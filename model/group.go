package model

type Group struct {
	ID    uint64 `json:"id" gorm:"primary_key"`
	Name  string `json:"name" gorm:"not null;unique"`
	Users []User `json:"users" gorm:"many2many:user_groups;"`
}

func (g Group) GetAll(page, pageSize uint64, where string, whereKeyword string) (gs []Group, total uint64) {
	var offset = (page - 1) * pageSize
	fieldsStr := "ID,name"
	if len(where) > 0 {
		mydb.Select(fieldsStr).Offset(offset).Limit(pageSize).Where(where+" = ?", whereKeyword).Find(&gs)
		mydb.Model(g).Where(where+" = ?", whereKeyword).Count(&total)
	} else {
		mydb.Select(fieldsStr).Offset(offset).Limit(pageSize).Find(&gs)
		mydb.Model(g).Count(&total)
	}

	return gs, total
}

func (g Group) GetRelatedUsers() (users []User, err error) {
	if err := mydb.Model(&g).Association("Users").Find(&users).Error; err != nil {
		return users, err
	}
	return users, nil
}

func (g Group) Save(users []uint64) (group Group, err error) {
	tx := mydb.Begin()
	// update or create
	if g.ID > 0 {
		err = tx.Save(&g).Error
	} else {
		err = tx.Create(&g).Error
	}

	if err != nil {
		tx.Rollback()
		return g, err
	}

	tx.Commit()
	g.AddUsers(users)
	return g, nil
}

func (g Group) AddUsers(idList []uint64) (err error) {
	tx := mydb.Begin()
	var users []User
	for _, id := range idList {
		users = append(users, User{ID: id})
	}
	tx.Model(&g).Association("Users").Clear()
	err = tx.Model(&g).Association("Users").Append(users).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (g Group) Get(withUsers bool) (result Group) {
	if g.ID == 0 {
		return Group{}
	}
	mydb.Select("id,name").First(&result, g.ID)

	if withUsers {
		mydb.Model(&result).Select("id").Association("Users").Find(&result.Users)
	}
	return result
}

func (g Group) Remove(ids []uint64) (err error) {
	tx := mydb.Begin()
	if err = tx.Exec("DELETE FROM groups  WHERE id IN (?) ", ids).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Exec("DELETE FROM user_groups  WHERE group_id IN (?) ", ids).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (g Group) InitDefault() error {
	tx := mydb.Begin()
	// Create
	g.Name = "normal"
	if err := tx.Create(&g).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
