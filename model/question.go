package model

import (
	"errors"
	"strings"
	"time"

	"yeetikuserver/utils"
)

//QLevel :
var QLevel = map[string]int64{"easy": 1, "normal": 2, "hard": 3}

//QType :
var QType = map[string]string{"单选题": "single", "多选题": "multiple", "判断题": "truefalse", "填空题": "filling", "问答题": "essay"}

//QuestionAnswerOptions : 建立一个中间表，记录题与选项之间的关系
//之所以不用gorm 的manytomany，是因为AnswerOption的content为unique,而在创建Question时，也会同时创建AnswerOption，此时我需要在创建AnswerOption时，
// 如果content已存在，则返回相应的ID。但如果使用manytomany,会直接报错并退回 。
type QuestionAnswerOptions struct {
	ID             uint64 `gorm:"primary_key"`
	QuestionID     uint64
	AnswerOptionID uint64
	IsCorrect      bool
}

// Save :
func (qao *QuestionAnswerOptions) Save() (id uint64, err error) {
	tx := mydb.Begin()
	if err := tx.Create(&qao).Error; err != nil {
		tx.Rollback()
		return 0, errors.New("无法为题目与选项建立关联")
	}
	tx.Commit()

	return qao.ID, nil
}

// DeleteByQuestionID :
func (qao QuestionAnswerOptions) DeleteByQuestionID() error {
	tx := mydb.Begin()
	if err := tx.Where("question_id = ?", qao.QuestionID).Delete(&qao).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()
	return nil
}

//AnswerOption :
type AnswerOption struct {
	ID        uint64 `json:"id" gorm:"primary_key"`
	Content   string `json:"content" gorm:"not null;unique"`
	IsCorrect bool   `json:"is_correct" gorm:"-"`
}

//Add :
func (op AnswerOption) Add() (uint64, error) {
	if err := mydb.Where("content = ?", op.Content).First(&op).Error; err == nil {
		return op.ID, nil
	}

	tx := mydb.Begin()
	if err := tx.Create(&op).Error; err != nil {
		tx.Rollback()
	}
	tx.Commit()

	return op.ID, nil
}

// Question :
type Question struct {
	ID             uint64         `json:"id" gorm:"primary_key"`
	Creator        uint64         `json:"creator" gorm:"not null;"`
	Type           string         `json:"type" gorm:"not null"`     //单选题，多选题，判断题
	Category       uint64         `json:"category" gorm:"not null"` //分类
	Subject        string         `json:"subject"`
	Score          float64        `json:"score"`
	Level          uint64         `json:"level"`
	Explanation    string         `json:"explanation"`
	Options        []AnswerOption `json:"options" gorm:"-"`
	CorrectOptions []uint64       `json:"correct_options" gorm:"-"`
	CorrectAnswers string         `json:"correct_answers"` //记录正确答案，用于填空题和是非题，是非题值为true/false,填空题为以逗号分隔的字符串
	TrueFalse      bool           `json:"true_or_false"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Get :
func (q Question) Get() (qs Question) {
	mydb.Select("id,creator,type,category,subject,score,level,explanation,correct_answers,true_false").First(&qs, q.ID)
	qs.Options = qs.GetOptions()
	return qs
}

// GetOptions :
func (q Question) GetOptions() []AnswerOption {
	//get the options
	if (q.Type == "single") || (q.Type == "multiple") {
		var ops []QuestionAnswerOptions
		var ids []uint64
		correctIDs := make(map[uint64]bool)
		mydb.Where(&QuestionAnswerOptions{QuestionID: q.ID}).Find(&ops)
		for _, value := range ops {
			ids = append(ids, value.AnswerOptionID)
			if value.IsCorrect {
				correctIDs[value.AnswerOptionID] = true
			}
		}
		mydb.Where("id in (?)", ids).Find(&q.Options)

		for index, value := range q.Options {
			if correctIDs[value.ID] {
				q.Options[index].IsCorrect = true
			}
		}
	}
	return q.Options
}

// GetAll :
func (Question) GetAll(where string, whereKeyword string) (qs []Question) {
	m := mydb.Select("id,creator,type,category,subject,score,level,explanation,correct_answers,true_false")
	if len(where) > 0 {
		if strings.EqualFold(where, "subject") {
			m = m.Where(where+" LIKE  ?", "%"+whereKeyword+"%")
		} else {
			m = m.Where(where+" = ?", whereKeyword)
		}
	}

	m.Find(&qs)
	return qs
}

//GetByCategory :
func (q Question) GetByCategory(page, pageSize uint64, where string, whereKeyword string) (qs []Question) {
	var offset = (page - 1) * pageSize
	cat := Category{ID: q.Category}
	if q.Category == 0 {
		qs = q.GetAll(where, whereKeyword)
	} else {
		m := mydb.Select("id,type,category,level,explanation,subject,score").Where("category IN (?)", cat.GetChild())
		if len(where) > 0 {
			if strings.EqualFold(where, "subject") {
				m = m.Where(where+" LIKE  ?", "%"+whereKeyword+"%")
			} else {
				m = m.Where(where+" = ?", whereKeyword)
			}
		}

		m.Offset(offset).Limit(pageSize).Find(&qs)

	}
	return qs
}

//UpdateCategory :
func (q Question) UpdateCategory(qusID, catID uint64) (err error) {
	q.ID = qusID
	tx := mydb.Begin()
	if err := tx.Model(&q).Update("category", catID).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return err
}

//GetUserFavorites : 获取用户收藏的题
//todo: 修复这里的bug
func (q Question) GetUserFavorites(userID, page, pageSize uint64) (result []Question, total uint64, err error) {
	var offset = (page - 1) * pageSize
	var questions []Question
	err = mydb.Model(&QuestionFavorites{}).Select("question_id as id").Offset(offset).Limit(pageSize).Where("user_id= ?", userID).Order("id").Scan(&questions).Error
	var ids []uint64
	for _, question := range questions {
		ids = append(ids, question.ID)
	}
	mydb.Model(&q).Where("id IN (?)", ids).Order("id").Scan(&questions)
	mydb.Model(&QuestionFavorites{}).Where("user_id = ?", userID).Count(&total)

	for index, q := range questions {
		questions[index].Options = q.GetOptions()
	}

	return questions, total, err
}

//GetUserWrong : 获取用户错题
func (q Question) GetUserWrong(userID, bankID, page, pageSize uint64) (result []Question, total uint64, err error) {
	var offset = (page - 1) * pageSize
	var questions []Question
	if bankID == 0 {
		err = mydb.Model(&QuestionRecord{}).Select("question_id as id").Offset(offset).Limit(pageSize).Where("user_id = ? AND result= ?", userID, bankID, false).Order("id").Scan(&questions).Error
		mydb.Model(&QuestionRecord{}).Where("user_id = ? AND result=false ", userID).Count(&total)
	} else {
		err = mydb.Model(&QuestionRecord{}).Select("question_id as id").Offset(offset).Limit(pageSize).Where("user_id = ? AND bank_id = ? AND result= ?", userID, bankID, false).Order("id").Scan(&questions).Error
		mydb.Model(&QuestionRecord{}).Where("user_id = ? AND bank_id = ? AND result=false ", userID, bankID).Count(&total)
	}

	var ids []uint64
	for _, question := range questions {
		ids = append(ids, question.ID)
	}

	mydb.Model(&q).Where("id IN (?)", ids).Order("id").Scan(&questions)

	for index, q := range questions {
		questions[index].Options = q.GetOptions()
	}
	return questions, total, err
}

//CountByCategory :  计算分类目录的数目
func (q Question) CountByCategory(where string, whereKeyword string) uint64 {
	var count uint64
	cat := Category{ID: q.Category}
	m := mydb.Model(q)
	if q.Category == 0 {
		if len(where) > 0 {
			if strings.EqualFold(where, "subject") {
				m = m.Where(where+" LIKE  ?", "%"+whereKeyword+"%")
			} else {
				m = m.Where(where+" = ?", whereKeyword)
			}
		} else {
			m.Count(&count)
		}
	} else {
		m.Where("category IN (?)", cat.GetChild()).Count(&count)
	}

	return count
}

//DeleteByCategory :
func (q Question) DeleteByCategory() error {
	tx := mydb.Begin()
	if err := tx.Where("category", q.Category).Delete(&q).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()
	return nil
}

// SetTrueFalse :
func (q *Question) SetTrueFalse(answer bool) {
	q.TrueFalse = answer
}

// SetFillingAnswers :
func (q *Question) SetFillingAnswers(answers string) {
	q.CorrectAnswers = answers
}

//Save :
func (q Question) Save() (uint64, error) {
	if q.ID > 0 {
		mydb.Save(&q)
	} else {
		tx := mydb.Begin()
		if err := tx.Create(&q).Error; err != nil {
			tx.Rollback()
			return 0, errors.New("无法创建问题")
		}
		tx.Commit()
	}
	return q.ID, nil
}

//Remove :
func (q Question) Remove(questionID uint64) error {
	q.ID = questionID
	answer := QuestionAnswerOptions{QuestionID: questionID}
	answer.DeleteByQuestionID()
	mydb.Model(&q).Association("Questions").Clear()

	mydb.Delete(&q)
	return nil
}

//IsCreator :
func (q Question) IsCreator(userID uint64, questionID uint64) bool {
	mydb.Where("id = ?", questionID).Find(&q)
	if q.Creator == userID {
		return true
	}
	return false

}

//QuestionFavorites : 题目收藏
type QuestionFavorites struct {
	ID         uint64 `json:"id" gorm:"primary_key"`
	UserID     uint64 `json:"user_id" gorm:"unique_index:idx_fav_user_question"` //与questionID建立联合唯一约束
	QuestionID uint64 `json:"question_id" gorm:"unique_index:idx_fav_user_question"`
	CreatedAt  time.Time
}

//Save :
func (q QuestionFavorites) Save(userID uint64, questionID uint64) error {
	sqlStr := "INSERT INTO question_favorites( user_id ,question_id ) VALUES (" + utils.Uint2Str(userID) + "," + utils.Uint2Str(questionID) + ")  ON CONFLICT (user_id,question_id) DO NOTHING;"
	mydb.Exec(sqlStr)
	return nil
}

//Remove :
func (q QuestionFavorites) Remove(userID uint64, questionID uint64) error {
	mydb.Where("user_id = ? AND question_id=? ", userID, questionID).Delete(&q)
	return nil
}

//IsFavorites :
func (q QuestionFavorites) IsFavorites(userID uint64, questionID uint64) (bool, error) {

	if err := mydb.Where("user_id = ? AND question_id=? ", userID, questionID).First(&q).Error; err != nil {
		return false, nil
	}

	if q.ID > 0 {
		return true, nil
	}

	return false, errors.New("do not exist")

}

//CountByUser : 计算用户的收藏题目数量
func (q QuestionFavorites) CountByUser(userID uint64) (total uint64, err error) {
	if err := mydb.Model(&q).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		total = 0
		err = errors.New("用户没有收藏")
	}
	return total, nil
}
