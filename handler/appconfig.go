package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

type Config struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	APIPrefix   string `json:"apiPrefix"`
	SplashImage string `json:"splashImage"`
	Logo        string `json:"logoImage"`
}

func (c Config) toString() string {
	return toJson(c)
}

func toJson(cfg interface{}) string {
	bytes, err := json.Marshal(cfg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return string(bytes)
}

func GetAppConfig(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var cfg Config
	response := Response{}.Default()
	path := "./config/clientConfig.json"
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		response.Status = http.StatusNotAcceptable
		response.Code = StatusNotAcceptable
		response.Message = "解析配置文件出现错误."
	} else {
		json.Unmarshal(raw, &cfg)
		response.Body["config"] = cfg
	}

	resq, _ := json.Marshal(response)
	w.Write(resq)
}

//UploadClientSplashImage :
func UploadClientSplashImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	imageInfo, _ := uploadImage(r, "client", "splash-image")
	b, err = json.Marshal(imageInfo)
	if err != nil {
		fmt.Print(err)
	}
	w.Write(b)
}

//UploadClientIconImage :
func UploadClientIconImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	imageInfo, _ := uploadImage(r, "client", "logo-image")
	b, err = json.Marshal(imageInfo)
	if err != nil {
		fmt.Print(err)
	}
	w.Write(b)
}

//SaveAppConfig :
func SaveAppConfig(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var resq []byte
	var newCfg Config
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &newCfg)

	cfg, _ := json.Marshal(newCfg)
	ioutil.WriteFile("./config/clientConfig.json", cfg, 0644)

	response := Response{}.Default()
	resq, err = json.Marshal(response)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	w.Write(resq)
}
