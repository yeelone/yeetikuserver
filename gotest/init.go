package gotest

import (
	"yeetikuserver/model"
	"yeetikuserver/config"
)

func init(){
	config.ParseConfig("../config/config.json")
	model.InitDatabaseTable()
}