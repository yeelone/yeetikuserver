package model

import (
	"encoding/json"
	"errors"
	"strings"

	"yeetikuserver/utils"
)

type Category struct {
	ID      uint64 `gorm:"primary_key"`
	Creator uint64 `json:"creator" gorm:"not null;"`
	Name    string `json:"name" gorm:"not null ;"`
	Parent  uint64 `json:"parent"` //默认为0
}

func (c *Category) Save() (uint64, error) {
	if len(c.Name) == 0 {
		return 0, errors.New("name is empty")
	}
	parent := Category{ID: c.Parent}
	if c.Parent != 0 {
		//检查 父目录是否存在
		result := mydb.First(&parent)
		if result.Error != nil {
			return 0, result.Error
		}
	}

	tx := mydb.Begin()
	if err := tx.Create(&c).Error; err != nil {
		tx.Rollback()
		return 0, errors.New("题库名称已存在，请重新输入")
	}

	tx.Commit()

	return c.ID, nil
}

func (c Category) GetAll() (cats []Category) {
	mydb.Select("ID,name,parent").Find(&cats)
	return cats
}

//Delete : 删除分类时，须同时删除相关联的题目
func (c Category) Delete() error {
	tx := mydb.Begin()
	if err := tx.Delete(&c).Error; err != nil {
		tx.Rollback()
		return errors.New("无法删除")
	}
	tx.Commit()

	question := Question{Category: c.ID}
	question.DeleteByCategory()
	return nil
}

func (c Category) Update() error {
	tx := mydb.Begin()
	if err := tx.Model(&c).Update("name", c.Name).Error; err != nil {
		tx.Rollback()
		return errors.New("无法更新")
	}
	tx.Commit()
	return nil

}

/* QueryByName ：
   @params : name
	name 的格式： 父目录.子目录.节点

//通过postgresql的with recursive 功能来递归查看分类ID
//语法如下：
// with recursive cte as
// (
//   select a.id,cast(a.name as varchar(100)) from categories a where name='test1' //test1 为最顶层父分类
//   union all
//   select k.id,cast(c.name||'.'||k.name as varchar(100)) as name from categories k inner join cte c on c.id=k.parent
// )select id,name from cte where name='test1.test11.test111';

*/
func (c Category) QueryByName(name string) (id uint64, err error) {
	catArray := strings.Split(name, ".")
	type Result struct {
		ID   uint64
		Name string
	}
	var result Result
	sql := `with recursive cte as
			(
			select a.id,cast(a.name as varchar(100)) from categories a where name='` + catArray[0] + `'
			union all
			select k.id,cast(c.name||'.'||k.name as varchar(100)) as name from categories k inner join cte c on c.id=k.parent
			)select id,name from cte where name='` + name + `';`

	mydb.Raw(sql).Scan(&result)
	if result.ID > 0 {
		return result.ID, nil
	}

	return 0, errors.New("category not exist")

}

//GetChild : 同时需要缓存
func (c Category) GetChild() []uint64 {
	level := CategoryTree{}.generateTree(utils.Uint2Str(c.ID))
	var ids []uint64
	for _, value := range level {
		ids = append(ids, value.ID)
	}
	return ids
}

func (c *Category) MarshalJSON() ([]byte, error) {

	return json.Marshal(&struct {
		ID     uint64 `json:"id"`
		Name   string `json:"name"`
		Parent uint64 `json:"parent"` //默认为1
	}{
		ID:     c.ID,
		Name:   c.Name,
		Parent: c.Parent,
	})
}

type CategoryTree struct {
	ID     uint64
	Parent uint64
	Level  int
	Branch string
}

//运用postgresql的tablefunc扩展的connectby扩展函数
//todo： 后续应该将数据缓存起来
func (t CategoryTree) generateTree(parent string) (tree []CategoryTree) {
	sql := `select id,parent,level,branch from connectby('categories','id','parent','` + parent + `',0,'~') as t(id bigint, parent bigint,level integer ,branch text);`
	mydb.Raw(sql).Scan(&tree)
	return tree
}
