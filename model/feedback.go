package model

import "time"

type Feedback struct {
	ID   uint64 `json:"id" gorm:"primary_key"`
	User User `json:"user"`                       //用户
	Content string `json:"content"`				//反馈内容
	Image   string `json:"iamge"`					//图片
	Contact string `json:"contact"`              //联系方式
	CreatedAt time.Time
}

func (f Feedback) Save() ( err error ) {
	tx := mydb.Begin()
	err = tx.Create(&f).Error

	if err != nil {
		tx.Rollback()
		return  err
	}

	tx.Commit()
	return nil
}
