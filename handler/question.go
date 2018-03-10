package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"yeetikuserver/model"
	"yeetikuserver/utils"

	"github.com/julienschmidt/httprouter"
	"github.com/tealeg/xlsx"
)

type _option struct {
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"`
}

type _resData struct {
	ID             uint64    `json:"id"`
	Category       uint64    `json:"category"`
	Score          float64   `json:"score"`
	Subject        string    `json:"subject"`
	Level          uint64    `json:"level"`
	Type           string    `json:"type"`
	FillingAnswers []string  `json:"filling-answers"` //记录填空题的答案
	TrueFalse      bool      `json:"true_or_false"`   //记录判断题的答案
	Options        []_option `json:"options"`         //记录选择题的选项，
	CorrectOptions []string  `json:"correct_options"`
}

func GetQuestion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	question := model.Question{ID: id}
	response.Body["question"] = question.Get()
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetQuestions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()

	query := parseQuery(r)
	question := model.Question{Category: query.CategoryID}

	response.Body["questions"] = question.GetByCategory(query.Page, query.PageSize, query.Field, query.Keyword)
	response.Body["total"] = question.CountByCategory(query.Field, query.Keyword)

	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetUserFavorites(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	userID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	response := Response{}.Default()
	query := parseQuery(r)
	question := model.Question{Category: query.CategoryID}

	response.Body["questions"], response.Body["total"], err = question.GetUserFavorites(userID, query.Page, query.PageSize)

	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err = json.Marshal(response)

	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}
	w.Write(b)
}

func GetUserWrong(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	userID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	response := Response{}.Default()
	query := parseQuery(r)
	question := model.Question{Category: query.CategoryID}

	response.Body["questions"], response.Body["total"], err = question.GetUserWrong(userID, query.Page, query.PageSize)

	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err = json.Marshal(response)

	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}
	w.Write(b)
}

func SaveQuestion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var resData _resData
	json.Unmarshal([]byte(result), &resData)
	m := model.Question{
		ID:       resData.ID,
		Creator:  utils.GetUserInfoFromContext(r.Context()),
		Score:    resData.Score,
		Level:    resData.Level,
		Subject:  resData.Subject,
		Type:     resData.Type,
		Category: resData.Category,
	}

	switch resData.Type {
	case "truefalse":
		m.SetTrueFalse(resData.TrueFalse)
	case "filling":
		m.SetFillingAnswers(strings.Join(resData.FillingAnswers, "||"))
	}

	response := Response{}.Default()
	id, err := m.Save()
	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	//当创建题目成功后，创建选项
	if saveOptions(resData.Options, id) != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	if err == nil {
		response.Code = StatusOK
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func ChangeCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var resData _resData
	json.Unmarshal([]byte(result), &resData)
	m := model.Question{
		ID:       resData.ID,
		Category: resData.Category,
	}

	response := Response{}.Default()
	var err error
	q := model.Question{}
	if err = q.UpdateCategory(m.ID, m.Category); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

/*
  客户端发来的数据格式：
	options = [
            { content: 'content a ', is_correct: false  },
            { content: 'content b ', is_correct: false  },
            { content: 'content c ', is_correct: true   },
            { content: 'content d ', is_correct: false  },
        ]
   现在需要将这些选项一条一条保存进数据库，并取出正确选项的ID，返回给Question
*/
func saveOptions(options []_option, questionID uint64) error {
	answer := &model.AnswerOption{}
	//在新增选项之前 ，需先将选项与题目之间的关联清除掉，以去除旧的关联
	qao := &model.QuestionAnswerOptions{QuestionID: questionID}
	if err := qao.DeleteByQuestionID(); err != nil {
		return err
	}

	for _, v := range options {
		answer.Content = v.Content
		var id uint64
		var err error

		if id, err = answer.Add(); err != nil {
			return err
		}

		answerops := model.QuestionAnswerOptions{QuestionID: questionID, AnswerOptionID: id, IsCorrect: v.IsCorrect}
		if _, err := answerops.Save(); err != nil {
			return err
		}
	}
	return nil
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userid := utils.GetUserInfoFromContext(r.Context())

	type parseModel struct {
		IDS []uint64 `json:"ids"`
	}
	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)
	question := model.Question{}

	response := Response{}.Default()
	var notAllowIDList []string
	var deleteFailedIDList []string
	for _, id := range m.IDS {
		if question.IsCreator(userid, id) {
			if err := question.Remove(id); err != nil {
				deleteFailedIDList = append(deleteFailedIDList, utils.Uint2Str(id))
			}
		} else {
			notAllowIDList = append(notAllowIDList, utils.Uint2Str(id))
		}
	}

	if len(deleteFailedIDList) > 0 {
		response.Code = StatusNotAcceptable
		response.Message = "[ " + strings.Join(deleteFailedIDList, "、") + "] 无法删除，请稍后重新再试！\n"
	}

	if len(notAllowIDList) > 0 {
		response.Code = StatusNotAcceptable
		response.Message = response.Message + "[ " + strings.Join(notAllowIDList, "、") + "] 无法删除，请稍后重新再试！\n"
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func AddFavorites(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := utils.GetUserInfoFromContext(r.Context())
	questionID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)

	favorites := model.QuestionFavorites{}
	response := Response{}.Default()
	if err := favorites.Save(userID, questionID); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func IsUserFavorites(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	userID := utils.GetUserInfoFromContext(r.Context())
	questionID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	favorites := model.QuestionFavorites{}

	if response.Body["isFavorites"], err = favorites.IsFavorites(userID, questionID); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func RemoveFavorites(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := utils.GetUserInfoFromContext(r.Context())
	questionID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)

	favorites := model.QuestionFavorites{}
	response := Response{}.Default()
	if err := favorites.Remove(userID, questionID); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

//插入做题的记录
func InsertRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	response := Response{}.Default()

	type recordModel struct {
		UserID    uint64             `json:"user_id"`
		BankID    uint64             `json:"bank_id"`
		Questions []model.RecordItem `json:"questions"`
		Current   int                `json:"current"`
	}

	m := &recordModel{}
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), m)

	userID := utils.GetUserInfoFromContext(r.Context())

	if userID != m.UserID {
		response.Code = StatusNotAcceptable
		response.Message = "无权进行操作"
	} else {
		//先保存练习题的记录
		record := model.QuestionRecord{}

		//再把最后练习题的编号保存到题库记录中
		//如果 bankID == 0 , 表示更新题库记录，因为bank_id == 0 的情况的话，表示用户可能只是在练习错误题集，错误题不用归属到指定的题库中
		if m.BankID > 0 {
			record.Save(m.BankID, userID, m.Questions)

			bankRecord := model.BankRecords{BankID: m.BankID, UserID: userID, LatestIndex: m.Current}
			if err = bankRecord.Insert(); err != nil {
				response.Code = StatusNotAcceptable
				response.Message = "无法为题库保存记录:" + err.Error()
			}
		}

		//如果用户是在练习错误题集，则更新QuestionRecord中对应题目的记录
		if m.BankID == 0 {
			record.UpdateQuestion(userID, m.Questions)
		}
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

// ImportFromExcel : 从excel中导入数据到数据库
func ImportFromExcel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	response := Response{}.Default()
	userID := utils.GetUserInfoFromContext(r.Context())
	path, filename, _, err := uploadFile(r, "questions", "excel-file")

	go SaveExcelToDB(userID, "."+path+"/"+filename)

	if err != nil {
		fmt.Println(err)
	}
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

/*SaveExcelToDB : 将excel表中的内容保存到数据库
excel模板的格式如下：
分类	题型	题干	试题分数	题目难度	试题解析	答案	选项A	选项B	选项C	选项D	选项E	选项F

@params 分类 ： [ 父目录.子目录.节点 ] 用 " . " 来区隔
@params 题型
*/
func SaveExcelToDB(userID uint64, filename string) (err error) {
	xlFile, err := xlsx.OpenFile(filename)
	if err != nil {
		fmt.Println("error", err)
	}
	sheet := xlFile.Sheets[0]
	feedbackFile := xlsx.NewFile()
	writeSheet, err := feedbackFile.AddSheet("错误")
	writeSheetFirstRow := writeSheet.AddRow()
	for _, cell := range sheet.Rows[0].Cells {
		errcell := writeSheetFirstRow.AddCell()
		errcell.Value = cell.String()
	}

	errorRowsCount := 0

	categoryMap := make(map[string]uint64) //记录分类的ID
	// 转化为map
	for index, row := range sheet.Rows[1:] {
		if len(row.Cells) < 1 {
			continue
		}
		cells := row.Cells
		category := cells[0].String()

		if _, ok := categoryMap[category]; !ok {
			cat, err := getCategoryFromDB(category)

			//如果类别不存在，则报错，错误信息写回excel表中
			if err != nil {
				copyRowToSheet(writeSheet, row, index, "分类不存在，请先创建分类")
				errorRowsCount++
				continue
			} else {
				categoryMap[category] = cat
			}
		}

		//map[int]map[string][]interface{}
		question, err := genQuesStructFromRow(userID, categoryMap[category], row)

		if err != nil {
			copyRowToSheet(writeSheet, row, index, err.Error())
			errorRowsCount++
			continue
		}

		//如果是选择题，则要对选项做处理
		var options []_option
		if question.Type == "single" || question.Type == "multiple" {
			options, err = genOptionStructFromRow(row)
		}

		questionID, err := question.Save()
		if err != nil {
			copyRowToSheet(writeSheet, row, index, err.Error())
			errorRowsCount++
			continue
		}

		if question.Type == "single" || question.Type == "multiple" {
			if saveOptions(options, questionID) != nil {
				copyRowToSheet(writeSheet, row, index, err.Error())
				errorRowsCount++
				continue
			}
		}

	}

	quesMessage := &model.QuestionImportMessage{UserID: userID}

	if errorRowsCount > 0 {
		savePath := "/download/temp/"
		if err = MkDir("." + savePath); err != nil {
			return err
		}

		newFilename := savePath + `error-批量导入题目-` + strconv.FormatInt(time.Now().Unix(), 10) + `.xlsx`
		err = feedbackFile.Save("." + newFilename)
		quesMessage.Success = false
		quesMessage.Unread = 1
		quesMessage.Content = newFilename
		quesMessage.Save()
	} else {
		quesMessage.Success = true
		quesMessage.Save()
	}

	if err != nil {
		fmt.Printf(err.Error())
	}

	return nil
}

/*
	@return id  : 分类对应的数据库ID
	@return err

	cat的格式是这样的：父目录.子目录.节点
*/
func getCategoryFromDB(cat string) (id uint64, err error) {

	catModel := &model.Category{}
	id, err = catModel.QueryByName(cat)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func genQuesStructFromRow(userID, categoryID uint64, row *xlsx.Row) (question model.Question, err error) {
	cell := row.Cells
	qType := ""
	answer := cell[6].String()
	truefalseAnswer := false
	level, _ := strconv.ParseUint(cell[4].String(), 10, 64)
	score, scoreErr := cell[3].Float()

	if scoreErr != nil {
		score = 1
	}
	if level == 0 {
		level = 1
	}

	if value, ok := model.QType[cell[1].String()]; ok {
		qType = value
	} else {
		return question, errors.New("不存在的题目类型")
	}

	if qType == "truefalse" {
		switch answer {
		case "A":
			fallthrough
		case "true":
			fallthrough
		case "TRUE":
			fallthrough
		case "正确":
			truefalseAnswer = true
		case "B":
			fallthrough
		case "false":
			fallthrough
		case "FALSE":
			fallthrough
		case "错误":
			truefalseAnswer = false
		}
	}

	question = model.Question{
		Creator:  userID,
		Level:    level,
		Score:    score,
		Subject:  cell[2].String(),
		Type:     qType,
		Category: categoryID,
	}

	switch qType {
	case "truefalse":
		question.SetTrueFalse(truefalseAnswer)
	case "filling":
		question.SetFillingAnswers(answer)
	}

	return question, err
}

func genOptionStructFromRow(row *xlsx.Row) (options []_option, err error) {
	answer := strings.Split(row.Cells[6].String(), ",")
	sort.Strings(answer)

	for index, cell := range row.Cells[7:] {
		option := _option{
			Content:   cell.String(),
			IsCorrect: false,
		}
		if len(option.Content) > 0 {
			//将选项转化为大写字母
			c := string(rune(index + 65))
			//u判断是否正确答案
			for _, v := range answer {
				if v == c {
					option.IsCorrect = true
				}
			}
		} else {
			continue
		}

		options = append(options, option)
	}
	return options, nil
}

func copyRowToSheet(sheet *xlsx.Sheet, row *xlsx.Row, index int, errMsg string) (err error) {
	errRow := sheet.AddRow()
	for _, cell := range row.Cells {
		errcell := errRow.AddCell()
		errcell.Value = cell.String()
	}
	errcell := errRow.AddCell()
	errcell.Value = `第` + strconv.Itoa(index+1) + `行`
	errcell = errRow.AddCell()
	errcell.Value = errMsg
	return nil
}

func GetQuestionImportResult(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)

	m := model.QuestionImportMessage{UserID: userID}
	response := Response{}.Default()
	if value, err := m.Query(); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Body["result"] = value
	}
	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func RemoveQuestionImportResult(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)

	m := model.QuestionImportMessage{UserID: userID}
	response := Response{}.Default()
	if err := m.Remove(); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
