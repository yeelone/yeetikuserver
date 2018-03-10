package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"yeetikuserver/utils"
)

//QuestionRecord ： 记录用户答题情况
type QuestionRecord struct {
	ID             uint64         `json:"id" gorm:"primary_key"`
	UserID         uint64         `json:"user" gorm:"unique_index:idx_user_question_bank"` // 与questionID\ bankID 建立联合唯一约束
	QuestionID     uint64         `json:"question" gorm:"unique_index:idx_user_question_bank"`
	BankID         uint64         `json:"bank" gorm:"unique_index:idx_user_question_bank"` //记录答错时处于哪个题库
	Result         bool           `json:"result"`
	ErrorOptions   []AnswerOption `json:"error_correct_options"`
	ErrorAnswers   string         `json:"error_correct_answers"` //记录答错答案，用于填空题
	ErrorTrueFalse bool           `json:"error_true_or_false"`
	ErrorCount     uint64         `json:"error_count"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

//RecordItem ： 供请求数据做解析用的模型
type RecordItem struct {
	ID     uint64 `json:"id"`
	Result bool   `json:"result"`
}

// Save ：
// 记录不存在时则新增，存在时则更新error_count
// Postgresql语法示例：
// insert into question_records (user_id, question_id, bank_id, result,error_count ) values(13,1,1,true,1)
// ON CONFLICT (user_id,question_id,bank_id)
// DO UPDATE SET    //将错误记录+1
// result=(case EXCLUDED.result when true then true else false END),
// error_count=question_records.error_count + (case EXCLUDED.result when true then - 1 else  1 END);
func (q QuestionRecord) Save(bankID uint64, userID uint64, questions []RecordItem) error {
	sqlStr := "INSERT INTO question_records( bank_id,user_id ,question_id ,result,error_count) VALUES "
	var sqlArr []string
	for _, question := range questions {
		sqlArr = append(sqlArr, `( `+utils.Uint2Str(bankID)+` , `+utils.Uint2Str(userID)+` , `+utils.Uint2Str(question.ID)+` , `+strconv.FormatBool(question.Result)+` , 1 )`)
	}
	s := strings.Join(sqlArr, ",")
	s = s + ` ON CONFLICT (user_id,question_id,bank_id) DO UPDATE SET result=(case EXCLUDED.result when true then true else false END),error_count=question_records.error_count + (case EXCLUDED.result when true then 0 else  1 END)`
	mydb.Exec(sqlStr + s)
	return nil
}

// UpdateQuestion:  更新指定题目ID的记录
func (q QuestionRecord) UpdateQuestion(userID uint64, questions []RecordItem) (err error) {
	for _, question := range questions {
		if err = mydb.Model(&q).Where("user_id = ? AND question_id = ? ", userID, question.ID).Update("result", question.Result).Error; err != nil {
			return err
		}
	}

	return nil
}

//统计用户在指定的题库下练习的记录，返回全部记录数目，以及做错的记录
func (q QuestionRecord) CountByBankID(bankID uint64, userID uint64) (total, wrong uint64) {
	if err := mydb.Where("user_id = ? AND bank_id = ?", userID, bankID).Find(&q).Count(&total).Error; err != nil {
		total = 0
	}
	if err := mydb.Where("user_id = ? AND bank_id = ? AND result = false ", userID, bankID).Find(&q).Count(&wrong).Error; err != nil {
		wrong = 0
	}
	return total, wrong
}

//获取用户的做题记录，包括做的题数量，错题数量， 题所属题库有哪些等等
//todo: 获取的数据项比较多，看看是不是加缓存改善一下
func (q QuestionRecord) GetByUser(userID uint64) (result interface{}, err error) {

	type Item struct {
		ID     uint64 `json:"id"`
		Name   string `json:"name"` //题库的名
		Image  string `json:"image"`
		Done   uint64 `json:"done"`
		Wrong  uint64 `json:"wrong"`
		Latest uint64 `json:"latest"`
	}

	type resultModel struct {
		Total     uint64          `json:"total"`
		Wrong     uint64          `json:"wrong"`
		Favorites uint64          `json:"favorites"`
		Banks     map[uint64]Item `json:"banks"`
	}
	var m resultModel

	bankMap := make(map[uint64]Item)

	//获取总题数
	if err := mydb.Model(&q).Where("user_id = ?", userID).Count(&m.Total).Error; err != nil {
		m.Total = 0
	}
	//获取错题总数
	if err := mydb.Model(&q).Where("user_id = ? AND result = false ", userID).Count(&m.Wrong).Error; err != nil {
		m.Wrong = 0
	}

	m.Favorites, _ = QuestionFavorites{}.CountByUser(userID)
	//获取每个题库的错题总数
	var items []Item
	str := `select bank_id as id , count(*) done , sum(case when result=false then 1 else 0 end ) as wrong from question_records where user_id=` + utils.Uint2Str(userID) + ` group by bank_id  order by id;`
	mydb.Raw(str).Scan(&items)

	var itemIDs []uint64
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}

	var itemNames []Item
	var itemLatest []Item
	if len(itemIDs) > 0 {
		//根据题库ID ，获取题库名称
		if err := mydb.Model(&Bank{}).Select("id,name,image").Where("id IN (?)", itemIDs).Scan(&itemNames).Error; err != nil {
			fmt.Printf("error %s \n ", err.Error())
		}

		if err := mydb.Model(&BankRecords{}).Select("bank_id as id,latest_index as latest").Where("user_id = ? AND bank_id IN (?)", userID, itemIDs).Scan(&itemLatest).Error; err != nil {
			fmt.Printf("error %s \n ", err.Error())
		}

	}
	if len(itemNames) > 0 {
		for index := range items {
			items[index].Name = itemNames[index].Name
			items[index].Image = itemNames[index].Image

			for j := range itemLatest {
				if items[index].ID == itemLatest[j].ID {
					items[index].Latest = itemLatest[j].Latest
				}
			}
			bankMap[items[index].ID] = items[index]
		}
	}

	m.Banks = bankMap
	return m, nil

}

func (q QuestionRecord) RemoveBank() (err error) {
	if err = mydb.Where(" bank_id = ? ", q.BankID).Delete(&QuestionRecord{}).Error; err != nil {
		return err
	}
	return nil
}
