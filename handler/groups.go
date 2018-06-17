package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"yeetikuserver/model"

	"github.com/julienschmidt/httprouter"
)

//GetGroups :
func GetGroups(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	query := parseQuery(r)
	var err error
	var b []byte

	response := Response{}.Default()
	response.Body["groups"], response.Body["total"] = model.Group{}.GetAll(query.Page, query.PageSize, query.Field, query.Keyword)
	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

//GetRelatedUsers :
func GetRelatedUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte

	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	g := &model.Group{ID: id}
	if response.Body["users"], err = g.GetRelatedUsers(); err != nil {
		response.Status = http.StatusNotAcceptable
		response.Code = StatusNotAcceptable
	}
	b, err = json.Marshal(response)

	if err != nil {
		fmt.Printf("errors :", err)
	}
	w.Write(b)
}

func SaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		ID    uint64   `json:id`
		Name  string   `json:"name"`
		Users []uint64 `json:"users`
	}
	var m parseModel
	var err error
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	group := model.Group{ID: m.ID, Name: m.Name}
	response := Response{}.Default()
	if group, err = group.Save(m.Users); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Body["group"] = group
	}
	b, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("errors :", err)
	}
	w.Write(b)
}

func DeleteGroups(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		IDS []uint64 `json:"ids"`
	}
	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)

	group := model.Group{}

	response := Response{}.Default()

	if err := group.Remove(m.IDS); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("errors :", err)
	}
	w.Write(b)
}
func GetGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	with_users, err := strconv.ParseBool(r.Form["with_users"][0])
	group := model.Group{ID: id}
	response.Body["group"] = group.Get(with_users)
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Printf("errors :", err)
	}
	w.Write(b)
}

func AddRelatedUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	type parseModel struct {
		Users []uint64 `json:"users"`
	}
	var m parseModel
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	group := model.Group{ID: id}
	response := Response{}.Default()
	if err = group.AddUsers(m.Users); err != nil {
		response.Status = http.StatusNotAcceptable
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Printf("errors :", err)
	}
	w.Write(b)
}
