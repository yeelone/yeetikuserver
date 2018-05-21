package model

import (
	"yeetikuserver/config"
	"yeetikuserver/db"
	"yeetikuserver/utils"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var mydb *gorm.DB
var kvdb *db.KVManager

//InitDatabaseTable :
func InitDatabaseTable() {
	mydb = db.GetInstance()
	kvdb = db.GetKVInstance()
	var user User
	var group Group
	var tag Tag
	var question Question
	var answerOption AnswerOption
	var bank Bank
	var category Category
	var options QuestionAnswerOptions
	var records BankRecords
	var questionRecords QuestionRecord
	var favoritesQuestions QuestionFavorites
	var feedback Feedback
	var bankTag Btags
	var comments Comments
	var exam Exam
	mydb.AutoMigrate(&user, &group, &bank, &bankTag, &tag,
		&category, &question, &answerOption, &options,
		&records, &questionRecords, &favoritesQuestions,
		&feedback, &comments, &exam)
	initAdmin()
	initGuest()
}

//initAdmin: 初始化管理员账号
func initAdmin() {
	var cfg = config.GetConfig()
	u := User{}
	//查看账号是否存在
	err := mydb.Where("email = ?", cfg.AdminAccount).First(&u).Error
	if err != nil {
		log.Info("系统无管理账号，即将初始化管理员账号")
		u.Email = cfg.AdminAccount
		u.Password = cfg.AdminDefaultPassword
		u.IsSuper = true
		u.Save()
	}
}

//initAdmin: 初始化管理员账号
func initGuest() {
	var cfg = config.GetConfig()
	u := User{}
	//查看账号是否存在
	err := mydb.Where("email = ?", cfg.GuestAccount).First(&u).Error
	if err != nil {
		log.Info("系统无游客账号，即将初始化游客账号")
		u.Email = cfg.GuestAccount
		u.Password = cfg.GuestDefaultPassword
		u.IsSuper = false
		u.Save()
	}
}

// Tree :
type Tree struct {
	ID     uint64
	Parent uint64
	Level  int
	Branch string
}

func (t Tree) generateTree(tablename, parent string) (tree []Tree) {
	sql := `select id,parent,level,branch from connectby('` + tablename + `','id','parent','` + parent + `',0,'~') as t(id bigint, parent bigint,level integer ,branch text);`
	mydb.Raw(sql).Scan(&tree)
	return tree
}

// GetChilrenID :
func (t Tree) GetChilrenID(tablename string, parent uint64) []uint64 {
	level := t.generateTree(tablename, utils.Uint2Str(parent))
	var ids []uint64
	for _, value := range level {
		ids = append(ids, value.ID)
	}
	return ids
}
