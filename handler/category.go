package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"../model"
	"../utils"
	"github.com/julienschmidt/httprouter"
)

type requestData struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Parent uint64 `json:"parent"`
}

func CreateCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var resdata requestData

	userid := utils.GetUserInfoFromContext(r.Context())

	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(data), &resdata)

	c := model.Category{Creator: userid, Name: resdata.Name, Parent: resdata.Parent}

	response := Response{}.Default()
	if id, err := c.Save(); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Code = StatusOK
		response.Body["ID"] = id
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetCategories(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte

	response := Response{}.Default()
	response.Body["categories"] = model.Category{}.GetAll()

	b, err = json.Marshal(response)
	fmt.Printf("%s", b)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func DeleteCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var resdata requestData
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(data), &resdata)
	c := model.Category{ID: resdata.ID}

	response := Response{}.Default()
	if err := c.Delete(); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Code = StatusOK
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func UpdateCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var resdata requestData

	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(data), &resdata)

	c := model.Category{ID: resdata.ID, Name: resdata.Name}

	response := Response{}.Default()
	if err := c.Update(); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Code = StatusOK
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
