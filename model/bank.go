package model

import (
	"encoding/json"
	"strings"
	"time"
)

type Allow struct {
	Type string   `json:"type"`
	Keys []uint64 `json:"keys"`
}

type Bank struct {
	ID          uint64     `json:"id" gorm:"primary_key"`
	Creator     uint64     `json:"creator" gorm:"not null;"`
	Users       []User     `json:"users" gorm:"many2many:bank_users;"`
	Groups      []Group    `json:"groups" gorm:"many2many:bank_groups;"`
	Tags        []Tag      `json:"tags" gorm:"many2many:bank_tags;"`
	Questions   []Question `json:"questions" gorm:"many2many:bank_questions;"`
	Btags       []Btags    `json:"banks" gorm:"many2many:banks_btags;"`
	Name        string     `json:"name" gorm:"not null;unique;"`
	Limited     bool       `json:"limited" gorm:"not null;"` //是否受限制，如果受限制，则根据AuthorityType判断授权给谁可访问
	AllowType   string     `json:"allow_type" `              //授权给 用户、组、标签
	Disable     bool       `json:"disable" gorm:"not null;"` //是否启用
	Image       string     `json:"image"`
	Description string     `json:"description"`
	Total       int        `json:"total"` //记录题目数量
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t *Bank) Update() (err error) {
	tx := mydb.Begin()
	err = tx.Model(Bank{}).Where("id = ?", t.ID).Updates(t).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (t Bank) GetRecords() (bank Bank) {
	return t
}

func (t *Bank) Save(keys []uint64) (err error) {
	tx := mydb.Begin()
	// update or create
	if t.ID > 0 {
		err = tx.Save(&t).Error
	} else {
		err = tx.Create(&t).Error
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	t.RemoveAssociation(t.ID)
	if t.Limited != false {
		switch t.AllowType {
		case "groups":
			var groups []Group
			for _, id := range keys {
				groups = append(groups, Group{ID: id})
			}
			err = tx.Model(&t).Association("Groups").Append(groups).Error
		case "users":
			var users []User
			for _, id := range keys {
				users = append(users, User{ID: id})
			}
			err = tx.Model(&t).Association("Users").Append(users).Error
		case "tags":
			var tags []Tag
			for _, id := range keys {
				tags = append(tags, Tag{ID: id})
			}
			err = tx.Model(&t).Association("Tags").Append(tags).Error
		}
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (t Bank) SaveQuestions(questions []uint64) (err error) {
	tx := mydb.Begin()
	var qs []Question
	for _, id := range questions {
		qs = append(qs, Question{ID: id})
	}

	tx.Model(&t).Association("Questions").Clear()
	mydb.First(&t)
	t.Total = len(questions)
	if err := tx.Save(&t).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Model(&t).Association("Questions").Append(qs).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (t Bank) DeleteQuestions(questions []uint64) (err error) {
	tx := mydb.Begin()

	var qs []Question
	for _, id := range questions {
		qs = append(qs, Question{ID: id})
	}

	if err = tx.Model(&t).Association("Questions").Delete(qs).Error; err != nil {
		tx.Rollback()
		return err
	}

	//更新 bank total 记录
	total := mydb.Model(&t).Association("Questions").Count()
	mydb.First(&t)
	t.Total = total
	mydb.Save(&t)

	tx.Commit()
	return nil
}

func (t Bank) Get() (item Bank) {
	mydb.Select("id,creator,name,limited,description,disable,image,allow_type").First(&item, t.ID)
	switch item.AllowType {
	case "groups":
		mydb.Model(&item).Association("Groups").Find(&item.Groups)
	case "users":
		mydb.Model(&item).Select("id,email").Association("Users").Find(&item.Users)
	case "tags":
		mydb.Model(&item).Association("Tags").Find(&item.Tags)
	}

	mydb.Model(&item).Association("Questions").Find(&item.Questions)
	return item
}

//GetRelatedQuestions :
//@params page
//@params pageSize
func (t Bank) GetRelatedQuestions(start, page, pageSize uint64) (qs []Question, total int) {
	var offset = (page-1)*pageSize + start
	mydb.Model(&t).Offset(offset).Limit(pageSize).Order("id").Association("Questions").Find(&qs)
	total = mydb.Model(&t).Association("Questions").Count()
	//get answers for every questions
	for index, q := range qs {
		qs[index].Options = q.GetOptions()
	}

	return qs, total
}

//GetAll :
func (t Bank) GetAll(page, pageSize uint64, where string, whereKeyword string) (banks []Bank, total int) {
	var offset = (page - 1) * pageSize
	fieldsStr := "id,name,description,disable,limited,image,total,allow_type" //默认情况下请求的字段

	m := mydb.Select(fieldsStr)
	if len(where) > 0 {
		if strings.EqualFold(where, "name") {
			m = m.Where(where+" LIKE  ? ", "%"+whereKeyword+"%")
		} else {
			m = m.Where(where+" = ? ", whereKeyword)
		}
	}

	m.Offset(offset).Limit(pageSize).Find(&banks)

	if err := mydb.Model(&t).Count(&total).Error; err != nil {
		total = 0
	}
	return banks, total
}

//GetAllEnable : get banks which is enable
func (t Bank) GetAllEnable(page, pageSize uint64, where string, whereKeyword string) (banks []Bank, total int) {
	var offset = (page - 1) * pageSize
	fieldsStr := "id,name,description,disable,limited,image,total,allow_type" //默认情况下请求的字段

	m := mydb.Select(fieldsStr)
	if len(where) > 0 {
		if strings.EqualFold(where, "name") {
			m = m.Where(where+" LIKE  ?", "%"+whereKeyword+"%")
		} else {
			m = m.Where(where+" = ?", whereKeyword)
		}
	}

	m.Where("disable=false").Offset(offset).Limit(pageSize).Find(&banks)

	if err := mydb.Model(&t).Count(&total).Error; err != nil {
		total = 0
	}
	return banks, total
}

//GetByUser : 查看用户正在练习或练习的题库
func (t Bank) GetByUser(page, pageSize, userID uint64) (banks []Bank, total int, err error) {
	var offset = (page - 1) * pageSize
	fieldsStr := "id,name,description,disable,limited,image,total,allow_type" //默认情况下请求的字段

	records, total, _ := BankRecords{UserID: userID}.GetByUser(page, pageSize)

	var bankIDs []uint64
	for _, record := range records {
		bankIDs = append(bankIDs, record.BankID)
	}
	mydb.Select(fieldsStr).Where("id IN (?)  AND disable=false ", bankIDs).Order("id").Offset(offset).Find(&banks)
	return banks, total, nil
}

func (t Bank) GetTags(bankID uint64) (tags []Btags, total int, err error) {
	t.ID = bankID
	mydb.Model(&t).Association("Btags").Find(&tags)
	total = mydb.Model(&t).Association("Btags").Count()
	return tags, total, err
}

//IsCreator :
func (t Bank) IsCreator(userID uint64, bankID uint64) bool {
	mydb.Where("id = ?", bankID).Find(&t)
	if t.Creator == userID {
		return true
	}
	return false
}

//RemoveAssociation :
func (t Bank) RemoveAssociation(bankID uint64) error {
	mydb.Model(&t).Association("Users").Clear()
	mydb.Model(&t).Association("Groups").Clear()
	mydb.Model(&t).Association("Tags").Clear()
	return nil
}

//Remove :
func (t Bank) Remove(bankID uint64) error {
	t.ID = bankID
	t.RemoveAssociation(bankID)
	mydb.Delete(&t)

	//删除相关记录
	record := &QuestionRecord{BankID: t.ID}
	record.RemoveBank()
	return nil
}

//SetEnable :
func (t Bank) SetEnable(bankID uint64) error {
	t.ID = bankID
	mydb.Model(&t).Where("id = ?", bankID).Update("disable", true)
	return nil
}

//SetDisable :
func (t Bank) SetDisable(bankID uint64) error {
	t.ID = bankID
	if err := mydb.Model(&t).Where("id = ?", bankID).Update("disable", t.Image).Error; err != nil {
		return err
	}

	return nil
}

//SaveImage :
func (t Bank) SaveImage() error {
	if err := mydb.Model(&t).Where("id = ?", t.ID).Update("image", false).Error; err != nil {
		return err
	}

	return nil
}

//MarshalJSON :
func (t *Bank) MarshalJSON() ([]byte, error) {
	type Alias Bank
	return json.Marshal(&struct {
		CreatedAt string `json:"create_time"`
		UpdatedAt string `json:"update_time"`
		*Alias
	}{
		CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(t),
	})
}
