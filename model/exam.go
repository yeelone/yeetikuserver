package model

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// Exam : Exam struct
type Exam struct {
	ID          uint64     `json:"id" gorm:"primary_key"`
	Creator     uint64     `json:"creator" gorm:"not null;"`
	Name        string     `json:"name" gorm:"not null;unique;"`
	Users       []User     `json:"users" gorm:"many2many:exam_users;"`
	Groups      []Group    `json:"groups" gorm:"many2many:exam_groups;"`
	Tags        []Tag      `json:"tags" gorm:"many2many:exam_tags;"`
	Questions   []Question `json:"questions" gorm:"many2many:exam_questions;"`
	Score       float64    `json:"score" gorm:"default:0.00"`
	Quantity    uint64     `json:"quantity" gorm:"default:0"`
	Description string     `json:"description"`
	Expired     bool       `json:"expired" gorm:"default:false"` //是否过期
	Time        uint64     `json:"time" gorm:"defualt:120" `     // n分钟
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

//QueryFieldStr :
const QueryFieldStr = "id,creator,name,score,quantity,description,expired,time,created_at"

// Save :
func (ex *Exam) Save(keys []uint64) (err error) {
	tx := mydb.Begin()
	err = tx.Create(&ex).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

// RandomCreate : 在用户已练习的题目中随机挑选N道题级成一张试卷
//todo : 有重复题的可能
func (ex Exam) RandomCreate(userid, total uint64) (item Exam, err error) {
	tx := mydb.Begin()

	userQuestionRecordIDs := []QuestionRecord{}
	if err = mydb.Select("distinct(question_id)").Where("user_id  = ? ", userid).Order(gorm.Expr("random()")).Limit(total).Find(&userQuestionRecordIDs).Error; err != nil {
		log.WithFields(log.Fields{
			"USER ID ": userid,
		}).Info("无法找到用户记录, error: " + err.Error())
	}

	for _, record := range userQuestionRecordIDs {
		q, _ := Question{ID: record.QuestionID}.Get()
		ex.Questions = append(ex.Questions, q)
	}
	ex.Name = "复习测试 " + time.Now().Format("2006-01-02 15:04:05")
	ex.Creator = userid
	ex.Description = "Just for Test"
	ex.Time = 9999
	ex.Quantity = uint64(len(ex.Questions))

	err = tx.Create(&ex).Error

	if err != nil {
		tx.Rollback()
		return item, err
	}

	tx.Commit()
	return ex, nil
}

//Get : 获取单张试卷
func (ex Exam) Get() (item Exam, err error) {
	mydb.Select(QueryFieldStr).First(&item, ex.ID)
	mydb.Model(&item).Association("Questions").Find(&item.Questions)
	newQues := []Question{}
	for _, question := range item.Questions {
		q, err := Question{ID: question.ID}.Get()
		if err != nil {
			continue
		}
		newQues = append(newQues, q)
	}
	item.Questions = newQues

	return item, nil
}

// CheckResultAndUpdateScore :
func (ex *Exam) CheckResultAndUpdateScore(score float64) (err error) {
	if err = mydb.First(&ex, ex.ID).Error; err != nil {
		return err
	}
	ex.Score = score
	ex.Expired = true
	if ex.ID > 0 {
		tx := mydb.Begin()
		if err := tx.Save(&ex).Error; err != nil {
			tx.Rollback()
			return errors.New("无法更新试卷")
		}
		tx.Commit()
	}
	return nil
}

// UpdateScore :
func (ex *Exam) UpdateScore(examid uint64, score float64) (err error) {
	ex.ID = examid
	return mydb.Model(&ex).Update("score", score).Error
}

//GetAll :
func (ex Exam) GetAll(page, pageSize uint64) (exs []Exam, total int, err error) {
	var offset = (page - 1) * pageSize

	m := mydb.Select(QueryFieldStr)

	m.Offset(offset).Limit(pageSize).Find(&exs)

	if err = mydb.Model(&ex).Count(&total).Error; err != nil {
		total = 0
	}
	return exs, total, err
}

//GetByCreator :
func (ex Exam) GetByCreator(userid, page, pageSize uint64) (exs []Exam, total int, err error) {
	var offset = (page - 1) * pageSize
	mydb.Select(QueryFieldStr).Where("creator = ?", userid).Offset(offset).Order("updated_at desc").Limit(pageSize).Find(&exs)

	if err = mydb.Model(&ex).Count(&total).Error; err != nil {
		total = 0
	}
	return exs, total, err
}

//MarshalJSON :
func (ex Exam) MarshalJSON() ([]byte, error) {
	type Alias Exam
	return json.Marshal(&struct {
		CreatedAt string
		UpdatedAt string
		Alias
	}{
		CreatedAt: ex.CreatedAt.Format("2006-01-02"),
		UpdatedAt: ex.UpdatedAt.Format("2006-01-02"),
		Alias:     (Alias)(ex),
	})
}
