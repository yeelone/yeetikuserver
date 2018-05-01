package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
	db "yeetikuserver/db"
	"yeetikuserver/utils"

	"github.com/pborman/uuid"
)

type User struct {
	ID              uint64  `json:"id" gorm:"primary_key"`
	Avatar          string  `json:"avatar"`
	Email           string  `json:"email" gorm:"not null;unique"`
	Name            string  `json:"name"`
	Nickname        string  `json:"nickname"`
	Age             uint64  `json:"age"`
	Sex             string  `json:"sex"`
	Address         string  `json:"address"`
	Phone           string  `json:"phone"`
	Groups          []Group `gorm:"many2many:user_groups;"`
	Tags            []Tag   `gorm:"many2many:user_tags;"`
	Password        string  `json:"password,omitempty" gorm:"not null"`
	PasswordConfirm string  `json:"password_confirm,omitempty"`
	Salt            string  `json:"salt,omitempty"`
	IsSuper         bool    `json:"is_super_user"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (u User) Get() (result User, err error) {
	if u.ID > 0 {
		value, err := kvdb.Get(db.USERBUCKET, string(u.ID))
		if err == nil {
			err = json.Unmarshal(value, &result)
			return result, nil
		}
	}

	if mydb.First(&result, u.ID).Error != nil {
		encoded, err := json.Marshal(result)
		if err != nil {
			return result, nil
		}
		kvdb.Set(db.USERBUCKET, string(result.ID), string(encoded))
	} else {
		return result, errors.New("cannot find any user")
	}
	return result, nil
}

//GetAll :
func (u User) GetAll(page, pageSize uint64, filterBy string, filterID uint64, where string, whereKeyword string) (users []User, total uint64) {
	users = []User{}
	var fieldsStr string
	var offset = (page - 1) * pageSize

	fieldsStr = "id,avatar,email,nickname,address,phone,name,age,sex,created_at" //默认情况下请求的字段

	if strings.EqualFold(filterBy, "group") {
		if len(where) > 0 {
			mydb.Select(fieldsStr).Joins("right join user_groups on users.id=user_groups.user_id and user_groups.group_id = ?", filterID).Offset(offset).Limit(pageSize).Where(where+" =  ?", whereKeyword).Find(&users)
		} else {
			mydb.Select(fieldsStr).Joins("right join user_groups on users.id=user_groups.user_id and user_groups.group_id = ?", filterID).Offset(offset).Limit(pageSize).Find(&users)
		}
	} else if strings.EqualFold(filterBy, "tag") {
		if len(where) > 0 {
			mydb.Select(fieldsStr).Joins("right join user_tags on users.id=user_tags.user_id and user_tags.tag_id = ?", filterID).Offset(offset).Limit(pageSize).Where(where+" = ?", whereKeyword).Find(&users)
		} else {
			mydb.Select(fieldsStr).Joins("right join user_tags on users.id=user_tags.user_id and user_tags.tag_id = ?", filterID).Offset(offset).Limit(pageSize).Find(&users)
		}
	} else {
		if len(where) > 0 {
			mydb.Select(fieldsStr).Offset(offset).Limit(pageSize).Where(where+" LIKE  ?", "%"+whereKeyword+"%").Find(&users)
			mydb.Model(u).Where(where+" LIKE  ?", "%"+whereKeyword+"%").Count(&total)
		} else {
			mydb.Select(fieldsStr).Offset(offset).Limit(pageSize).Find(&users)
			mydb.Model(u).Count(&total)
		}
	}

	return users, total
}

//Save :
func (u User) Save() (User, error) {
	tx := mydb.Begin()

	if u.IsSuper {
		u.IsSuper = true
	} else {
		u.IsSuper = false
	}
	if u.ID > 0 {
		tx.Model(&u).Where("id = ?", u.ID).Updates(u)
		encoded, err := json.Marshal(u)
		if err == nil {
			kvdb.Set(db.USERBUCKET, string(u.ID), string(encoded))
		}

	} else if len(u.Email) > 0 {

		u.Salt = strings.Replace(uuid.NewUUID().String(), "-", "", -1)
		u.Password = utils.EncryptPassword(u.Password, u.Salt)
		if err := tx.Create(&u).Error; err != nil {
			tx.Rollback()
			return u, err
		}
	}

	tx.Commit()

	return u, nil
}

//Remove :
func (u User) Remove(ids []uint64) (err error) {
	tx := mydb.Begin()
	if err = tx.Exec("DELETE FROM users  WHERE id IN (?) ", ids).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Exec("DELETE FROM user_groups  WHERE user_id IN (?) ", ids).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Exec("DELETE FROM user_tags  WHERE user_id IN (?) ", ids).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

//Valid :
func (u *User) Valid(email string, password string) bool {
	mydb.Where("email = ?", email).First(&u)
	if u.CheckPassword(password) {
		return true
	}
	return false
}

// IsEmailAdmin : check if the current user is admin
func (u User) IsEmailAdmin(email string) bool {
	err := mydb.Where("email = ?", email).First(&u).Error
	if err != nil {
		return false
	}
	return u.IsSuper
}

// IsIDAdmin : check if the current user is admin
func (u User) IsIDAdmin(id uint64) bool {
	err := mydb.Where("id = ?", id).First(&u).Error
	if err != nil {
		return false
	}
	return u.IsSuper
}

//SetAvatar :
func (u User) SetAvatar() (err error) {
	return mydb.Model(&u).Update("avatar", u.Avatar).Error
}

// CheckPassword  : check the user's password
func (u User) CheckPassword(password string) bool {
	return u.Password == utils.EncryptPassword(password, u.Salt)
}

// ResetPassword :
func (u User) ResetPassword(email, password string) (err error) {

	if err = mydb.Where("email = ?", email).First(&u).Error; err != nil {
		return err
	}

	u.Password = utils.EncryptPassword(password, u.Salt)
	return mydb.Model(&u).Update("password", u.Password).Error
}

func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		CreatedAt string `json:"createTime"`
		UpdatedAt string `json:"updateTime"`
		Password  string `json:"password,omitempty" gorm:"not null"`
		Salt      string `json:"salt,omitempty"`
		*Alias
	}{
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(u),
	})
}
