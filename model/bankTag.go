package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"../utils"
	log "github.com/sirupsen/logrus"
)

//Btags : 题库的标签化管理 ，最多三级标签
type Btags struct {
	ID     uint64 `json:"id" gorm:"primary_key"`
	Name   string `json:"name" gorm:"not null;unique"`
	Parent uint64 `json:"parent" gorm:"default:0"` //默认为0
	Level  int    `json:"level" gorm:"default:1"`
	Banks  []Bank `json:"banks" gorm:"many2many:banks_btags;"`
}

func (t Btags) Save() (tag Btags, err error) {

	if len(t.Name) == 0 {
		return tag, errors.New("name is empty")
	}

	parent := Btags{ID: t.Parent}
	if t.Parent != 0 {
		//检查父标签是否存在
		if err = mydb.First(&parent).Error; err != nil {
			return tag, err
		}
	}

	tx := mydb.Begin()
	err = mydb.Where("name = ? AND parent = ?", t.Name, t.Parent).First(&tag).Error
	if err != nil {
		if t.Parent != 0 {
			t.Level = parent.Level + 1
			if t.Level > 4 {
				return Btags{}, errors.New("最多只能有三级标签")
			}
		} else {
			t.Level = 1
		}
		err = tx.Create(&t).Error
	} else {
		t = tag
	}

	if err != nil {
		tx.Rollback()
		return t, err
	}

	tx.Commit()
	// t.relateBank(bank)
	return t, nil
}

func (t Btags) GetRelatedBanks(page, pageSize uint64) (banks []Bank, total int) {
	var offset = (page - 1) * pageSize

	//根据标签ID以及所属子标签的ID，查询所有关联的题库ID
	ids := t.GetChild(t.ID)
	rows, _ := mydb.Raw("SELECT bank_id as bank FROM banks_btags WHERE btags_id IN (?) offset ? limit ?;", ids, offset, pageSize).Rows()
	defer rows.Close()
	var bankIDs []uint64
	var bank uint64
	for rows.Next() {
		rows.Scan(&bank)
		bankIDs = append(bankIDs, bank)
	}
	fieldsStr := "id,name,description,disable,limited,image,total,allow_type,created_at,updated_at" //默认情况下请求的字段
	if err := mydb.Select(fieldsStr).Where("id IN (?)  AND disable=false ", bankIDs).Find(&banks).Error; err != nil {
		return banks, 0
	}

	// mydb.Exec("SELECT count(*) FROM banks_btags WHERE btags_id IN (?)", ids).Scan(&total)

	return banks, total
}

//relateBank : 关联题库跟标签
func (t Btags) RelateBank(bankID uint64) (err error) {
	tx := mydb.Begin()
	bank := Bank{ID: bankID}

	err = tx.Model(&t).Association("Banks").Append(bank).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//RemoveBankRelated : 移除与指定题库的关联关系
func (t Btags) RemoveBankRelated(bankID uint64, tagID uint64) (err error) {
	tx := mydb.Begin()
	err = mydb.Where("id = ?", tagID).First(&t).Error
	if err != nil {
		log.Warn("标签不存在")
		return err
	}

	bank := Bank{ID: bankID}
	if err = tx.Model(&t).Association("Banks").Delete(&bank).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

//Delete ： 删除标签
func (t Btags) Delete(tagID uint64) (err error) {
	tx := mydb.Begin()
	ids := t.GetChild(tagID)
	err = tx.Delete(Btags{}, "id IN (?)", ids).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, id := range ids {
		tag := Btags{ID: id}
		if err = tx.Model(&tag).Association("Banks").Clear().Error; err != nil {
			fmt.Println("err", err)
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (t Btags) GetTagTree(query string) (tree interface{}, total int, err error) {
	m := mydb.Select("id,name,parent,level")
	tags := []Btags{}

	if len(query) > 0 {
		m = m.Where("name LIKE ?", "%"+query+"%")
	}

	if err = m.Order("level").Find(&tags).Error; err != nil {
		log.Warn("无法获取题库标签 , error :" + err.Error())
		return tags, 0, err
	}

	if err = mydb.Model(&t).Count(&total).Error; err != nil {
		total = 0
	}

	tree = formatTagsToTree(tags)
	return tree, total, nil
}

/**formatTags :
	因为数据库有对level进行排序且level最多为3，所以在转换成树状结构时会简单很多。
	{id: 1, name: "tag1", level: 1, parent: 0}
	{id: 4, name: "tag4", level: 1, parent: 0}
	{id: 2, name: "tag2", level: 2, parent: 1}
	{id: 5, name: "tag5", level: 2, parent: 4}
	{id: 6, name: "tag6", level: 3, parent: 5}
	{id: 3, name: "tag3", level: 3, parent: 2}
**/
func formatTagsToTree(tags []Btags) (tagsTree interface{}) {
	type TreeStruct struct {
		Tag      Btags                 `json:"tag"`
		Children map[uint64]TreeStruct `json:"children"`
	}

	m := map[uint64]TreeStruct{}

	levelRecord := make(map[uint64]uint64) //用于记录第二层与第一层的映射关系
	for _, v := range tags {
		if v.Level == 1 {
			m[v.ID] = TreeStruct{Tag: v, Children: make(map[uint64]TreeStruct)}
		}

		if v.Level == 2 {
			parent := m[v.Parent]
			parent.Children[v.ID] = TreeStruct{Tag: v, Children: make(map[uint64]TreeStruct)}
			levelRecord[v.ID] = v.Parent
		}

		if v.Level == 3 {
			root := m[levelRecord[v.Parent]]
			parent := root.Children[v.Parent]
			parent.Children[v.ID] = TreeStruct{Tag: v, Children: make(map[uint64]TreeStruct)}
		}
	}

	return m
}

func (t Btags) GetChild(id uint64) []uint64 {
	level := TagsTree{}.generateTree(utils.Uint2Str(id))
	var ids []uint64
	for _, value := range level {
		ids = append(ids, value.ID)
	}
	return ids
}
func (t *Btags) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID     uint64 `json:"id" gorm:"primary_key"`
		Name   string `json:"name" gorm:"not null;unique"`
		Level  int    `json:"level" gorm:"default:1"`
		Parent uint64 `json:"parent"`
	}{
		ID:     t.ID,
		Name:   t.Name,
		Parent: t.Parent,
		Level:  t.Level,
	})
}

type TagsTree struct {
	ID     uint64
	Parent uint64
	Level  int
	Branch string
}

func (t TagsTree) generateTree(parent string) (tree []TagsTree) {
	sql := `select id,parent,level,branch from connectby('btags','id','parent','` + parent + `',0,'~') as t(id bigint, parent bigint,level integer ,branch text);`
	mydb.Raw(sql).Scan(&tree)
	return tree
}
