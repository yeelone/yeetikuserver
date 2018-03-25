package db

import (
	"sync"

	c "yeetikuserver/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var once sync.Once
var dbInstance *gorm.DB

// GetInstance : 获取数据库实例
func GetInstance() *gorm.DB {
	once.Do(func() {
		if dbInstance == nil {
			var err error
			var cfg = c.GetConfig()

			dbInstance, err = gorm.Open("postgres", "host=localhost user="+cfg.DBUser+" dbname="+cfg.DBName+" password="+cfg.DBPassword+" sslmode=disable ")
			if err != nil {
				log.Fatal("failed to connect database:", err)
				panic("failed to connect database")
			} else {
				log.Println("db connected")
			}
		}
	})
	return dbInstance
}
