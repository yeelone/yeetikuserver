package model

//Tag :
type Tag struct {
	ID    uint64 `json:"id" gorm:"primary_key"`
	Name  string `json:"name" gorm:"not null;unique"`
	Users []User `json:"users" gorm:"many2many:user_tags;"`
}

//GetAll :
func (t Tag) GetAll(page, pageSize uint64, where string, whereKeyword string) (ts []Tag, total uint64) {
	var offset = (page - 1) * pageSize
	fieldsStr := "ID,name"
	mydb.Select(fieldsStr).Offset(offset).Limit(pageSize).Find(&ts)
	mydb.Model(t).Count(&total)

	if len(where) > 0 {
		mydb.Select(fieldsStr).Offset(offset).Limit(pageSize).Where(where+" = ?", whereKeyword).Find(&ts)
		mydb.Model(t).Where(where+" = ?", whereKeyword).Count(&total)
	} else {
		mydb.Select(fieldsStr).Offset(offset).Limit(pageSize).Find(&ts)
		mydb.Model(t).Count(&total)
	}

	return ts, total
}

//Save :
func (t Tag) Save(users []uint64) (tag Tag, err error) {
	tx := mydb.Begin()
	// update or create
	if t.ID > 0 {
		err = tx.Save(&t).Error
	} else {
		err = tx.Create(&t).Error
	}

	if err != nil {
		tx.Rollback()
		return t, err
	}

	tx.Commit()
	t.relateUsers(users)
	return t, nil
}

// relateUsers
func (t Tag) relateUsers(keys []uint64) (err error) {
	tx := mydb.Begin()
	var users []User
	for _, id := range keys {
		users = append(users, User{ID: id})
	}

	tx.Model(&t).Association("Users").Clear()
	err = tx.Model(&t).Association("Users").Append(users).Error

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//Get :
func (t Tag) Get(withUsers bool) (result Tag) {
	if t.ID == 0 {
		return Tag{}
	}
	mydb.Select("id,name").First(&result, t.ID)

	if withUsers {
		mydb.Model(&result).Select("id").Association("Users").Find(&result.Users)
	}
	return result
}

//Remove :
func (t Tag) Remove(ids []uint64) (err error) {
	tx := mydb.Begin()
	if err = tx.Exec("DELETE FROM tags  WHERE id IN (?) ", ids).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Exec("DELETE FROM user_tags  WHERE tag_id IN (?) ", ids).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

//InitDefault :
func (t Tag) InitDefault() error {
	tx := mydb.Begin()
	// Create
	t.Name = "normal"
	if err := tx.Create(&t).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
