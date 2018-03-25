package model

import (
	"errors"
	"strings"
	"time"
)

type Comments struct {
	ID        uint64 `json:"id" gorm:"primary_key"`
	Creator   uint64 `json:"creator" gorm:"not null;"`
	User      User   `json:"user" gorm:"-"`
	Question  uint64 `json:"question"`
	Image     string `json:"image"`
	Parent    uint64 `json:"parent"`
	Content   string `json:"content"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Save : 新增评论
func (c *Comments) Save() (id uint64, err error) {
	if len(c.Content) == 0 {
		return 0, errors.New("content is empty")
	}

	parent := Comments{ID: c.Parent}
	if c.Parent != 0 {
		//检查父ID是否存在
		result := mydb.First(&parent)
		if result.Error != nil {
			return 0, result.Error
		}
	}

	tx := mydb.Begin()
	// update or create
	if c.ID > 0 {
		err = tx.Save(&c).Error
	} else {
		err = tx.Create(&c).Error
	}

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return c.ID, nil
}

// GetAll : 获取所有评论,仅限管理员后台调用
func (c Comments) GetAll(page, pageSize uint64, where string, whereKeyword string) (cs []Comments, total uint64) {
	var offset = (page - 1) * pageSize
	m := mydb.Model(&c)
	if len(where) > 0 {
		if strings.EqualFold(where, "content") {
			m = m.Where(where+" LIKE  ?", "%"+whereKeyword+"%")
		} else {
			m = m.Where(where+" = ?", whereKeyword)
		}
	}
	m.Offset(offset).Limit(pageSize).Find(&cs)
	mydb.Model(&c).Count(&total)
	return cs, total
}

//GetByQuestion :
func (c Comments) GetByQuestion(question, page, pageSize uint64) (cs []Comments, total uint64) {
	var offset = (page - 1) * pageSize
	m := mydb.Model(&c).Where(&Comments{Parent: 0, Question: question})

	m.Offset(offset).Limit(pageSize).Find(&cs)
	m.Count(&total)

	return cs, total
}

// GetAllParent : 获取所有一级评论
func (c Comments) GetAllParent(page, pageSize uint64, where string, whereKeyword string) (cs []Comments, total uint64) {
	var offset = (page - 1) * pageSize
	m := mydb.Model(&c).Where(&Comments{Parent: 0})

	if len(where) > 0 {
		if strings.EqualFold(where, "content") {
			m = m.Where(where+" LIKE  ?", "%"+whereKeyword+"%")
		} else {
			m = m.Where(where+" = ?", whereKeyword)
		}
	}
	m.Offset(offset).Limit(pageSize).Find(&cs)
	mydb.Model(&c).Where(&Comments{Parent: 0}).Count(&total)

	// for index, item := range cs {
	// 	if item.Creator > 0 {
	// 		cs[index].User, _ = User{ID: item.Creator}.Get()
	// 	}
	// }
	return cs, total
}

//GetChilren 获取指定一级评论的子评论
func (c Comments) GetChilren(parent, page, pageSize uint64) (cs []Comments, total uint64) {
	var offset = (page - 1) * pageSize
	ids := Tree{}.GetChilrenID("comments", c.ID)
	mydb.Model(&c).Where("parent= ?", parent).Count(&total)
	m := mydb.Select("id,creator,content,parent,question,image").Where("id IN (?)", ids)
	if m.Offset(offset).Limit(pageSize).Find(&cs).Error != nil {
		return nil, 0
	}
	return cs, total
}

//IsCreator :
func (c Comments) IsCreator(userID uint64, commentID uint64) bool {
	mydb.Where("id = ?", commentID).Find(&c)
	if c.Creator == userID {
		return true
	}
	return false
}

//Remove :
func (c Comments) Remove(commentID uint64) error {
	c.ID = commentID
	return mydb.Where("id =? OR parent = ?", commentID, commentID).Delete(&c).Error
}
