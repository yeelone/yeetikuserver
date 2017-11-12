package utils

import (
	"sync"

	"github.com/jinzhu/gorm"
)

var once sync.Once
var dbInstance *gorm.DB = nil

func LogInstance() *gorm.DB {
	once.Do(func() {
		if dbInstance == nil {
			var err error
			dbInstance, err = gorm.Open("postgres", "host=localhost user=elone dbname=elone password=123456 sslmode=disable ")
			if err != nil {
				panic("failed to connect database")
			}
		}
	})
	return dbInstance
}
