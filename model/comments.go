package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
	db "yeetikuserver/db"
	"yeetikuserver/utils"
)

type Comments struct {
	ID        uint64    `json:"id" gorm:"primary_key"`
	Creator   uint64    `json:"creator" gorm:"not null;"`
	User      User      `json:"user" gorm:"-"`
	Question  uint64    `json:"question"`
	Image     string    `json:"image"`
	Parent    uint64    `json:"parent"`
	Like      uint64    `json:"like"`
	DisLike   uint64    `json:"dislike"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"time"`
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
	fillUsers(cs)
	mydb.Model(&c).Count(&total)
	return cs, total
}

//GetByQuestion :
func (c Comments) GetByQuestion(question, parent, page, pageSize uint64) (cs []Comments, total uint64) {
	var offset = (page - 1) * pageSize
	m := mydb.Model(&c).Where(&Comments{Parent: parent, Question: question})
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
	fillUsers(cs)
	return cs, total
}

func fillUsers(cs []Comments) {
	for index, item := range cs {
		if item.Creator > 0 {
			cs[index].User, _ = User{ID: item.Creator}.Get()
		}
	}
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

// LikeComment :
func (c Comments) LikeComment(userID, commentID uint64) (count uint64, err error) {
	c.ID = commentID

	if mydb.First(&c).Error != nil {
		return 0, errors.New("查询失败")
	}
	key := "USER_" + utils.Uint2Str(userID) + "_COMMENT_" + utils.Uint2Str(commentID)

	value, err := kvdb.Get(db.USERLIKECOMMENTS, string(key))
	fmt.Printf("%s %v\n ", value, string(value) == "LIKE")
	if string(value) == "LIKE" { //曾经点击过Like,所以要减1
		c.Like = c.Like - 1
		kvdb.Delete(db.USERLIKECOMMENTS, string(key))
	} else if string(value) == "DISLIKE" { //曾经点击过Dislike,,所以要减1
		c.DisLike = c.DisLike - 1
		kvdb.Set(db.USERLIKECOMMENTS, string(key), string("LIKE"))
	} else {
		c.Like = c.Like + 1
		kvdb.Set(db.USERLIKECOMMENTS, string(key), string("LIKE"))
	}

	tx := mydb.Begin()
	err = tx.Save(&c).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()

	return c.Like, err
}

//DislikeComment :
func (c Comments) DislikeComment(userID, commentID uint64) (count uint64, err error) {
	c.ID = commentID

	if mydb.First(&c).Error != nil {
		return 0, errors.New("查询失败")
	}
	key := "USER_" + string(userID) + "_COMMENT_" + string(commentID)

	value, err := kvdb.Get(db.USERLIKECOMMENTS, string(key))
	if string(value) == "DISLIKE" {
		c.DisLike = c.DisLike - 1
		kvdb.Delete(db.USERLIKECOMMENTS, string(key))
	} else if string(value) == "LIKE" {
		c.Like = c.Like - 1
		kvdb.Set(db.USERLIKECOMMENTS, string(key), string("DISLIKE"))
	} else {
		c.DisLike = c.DisLike + 1
		kvdb.Set(db.USERLIKECOMMENTS, string(key), string("DISLIKE"))
	}

	tx := mydb.Begin()
	err = tx.Save(&c).Error

	if err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()

	return c.DisLike, err
}
