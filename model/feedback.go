package model

import (
	"strings"
	"time"
)

type Feedback struct {
	ID        uint64 `json:"id" gorm:"primary_key"`
	User      User   `json:"user"`    //用户
	Content   string `json:"content"` //反馈内容
	Image     string `json:"image"`   //图片
	Close     bool   `json:"close"`   //图片
	Contact   string `json:"contact"` //联系方式
	CreatedAt time.Time
}

func (f Feedback) Save() (err error) {
	tx := mydb.Begin()
	err = tx.Create(&f).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (f Feedback) GetAll(page, pageSize uint64, where string, whereKeyword string) (feeds []Feedback, total uint64) {
	m := mydb.Select("id,content,image,close,contact,created_at")
	if len(where) > 0 {
		if strings.EqualFold(where, "content") {
			m = m.Where(where+" LIKE  ?", "%"+whereKeyword+"%")
		} else {
			m = m.Where(where+" = ?", whereKeyword)
		}
	}

	m.Find(&feeds)
	m.Count(&total)
	return feeds, total
}
