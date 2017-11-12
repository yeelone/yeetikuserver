package config

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

type ConfigStruct struct {
	Domain               string `json:"domain"`
	SecretKey            string `json:"secret_key"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	DB                   string `json:"db"`
	DBName               string `json:"db_name"`
	DBUser               string `json:"db_user"`
	DBPassword           string `json:"db_password"`
	AdminAccount         string `json:"admin_account"`
	AdminDefaultPassword string `json:"admin_default_password"`
	GuestAccount         string `json:"guest_account"`
	GuestDefaultPassword string `json:"guest_default_password"`
	PublicSalt           string `json:"public_salt"`
}

var Config ConfigStruct

func parseJsonFile(path string, v interface{}) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("配置文件读取失败:", err)
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(v)
	if err != nil {
		log.Fatal("配置文件解析失败:", err)
	}
}

func GetConfig() ConfigStruct {
	if Config.DB == "" {
		log.Fatal("数据库地址还没有配置,请到config.json内配置db字段.")
		os.Exit(1)
	}

	return Config
}

func ParseConfig(file string) {
	parseJsonFile(file, &Config)
	if Config.DB == "" {
		log.Fatal("数据库地址还没有配置,请到config.json内配置db字段.")
		os.Exit(1)
	}

}

// func init() {
// 	GetConfig()
// }
