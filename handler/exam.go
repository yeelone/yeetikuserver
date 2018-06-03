package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"yeetikuserver/model"
	"yeetikuserver/utils"

	"github.com/julienschmidt/httprouter"
)

// CreateExam :
func CreateExam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	userID := utils.GetUserInfoFromContext(r.Context())
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	exam := model.Exam{ID: id}
	response.Body["exam"], _ = exam.RandomCreate(userID, 20)
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

// GetUserExams  :
func GetUserExams(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	query := parseQuery(r)
	response := Response{}.Default()

	userID := utils.GetUserInfoFromContext(r.Context())

	response.Body["exams"], response.Body["total"], _ = model.Exam{}.GetByCreator(userID, query.Page, query.PageSize)

	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

// GetExam :
func GetExam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	response := Response{}.Default()

	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	response.Body["exam"], _ = model.Exam{ID: id}.Get()
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}

	w.Write(b)
}

// UpdateExamScore  :
func UpdateExamScore(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		ID    uint64  `json:"id"`
		Score float64 `json:"score"`
	}

	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)
	exam := model.Exam{}
	response := Response{}.Default()
	if err := exam.UpdateScore(m.ID, m.Score); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = "无法更新分数，error : " + err.Error()
	} else {
		response.Message = "成功更新分数"
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
