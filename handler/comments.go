package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"yeetikuserver/model"
	"yeetikuserver/utils"

	"github.com/julienschmidt/httprouter"
)

//CreateComments : 创建评论
func CreateComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var parseModel model.Comments

	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(data), &parseModel)

	parseModel.Creator = utils.GetUserInfoFromContext(r.Context())
	response := Response{}.Default()
	if id, err := parseModel.Save(); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Code = StatusOK
		response.Body["id"] = id
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func LikeComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userid, _ := strconv.ParseUint(ps.ByName("userid"), 10, 64)
	commentid, _ := strconv.ParseUint(ps.ByName("commentid"), 10, 64)

	c := model.Comments{}
	response := Response{}.Default()
	if count, err := c.LikeComment(userid, commentid); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Code = StatusOK
		response.Body["count"] = count
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func DislikeComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userid, _ := strconv.ParseUint(ps.ByName("userid"), 10, 64)
	commentid, _ := strconv.ParseUint(ps.ByName("commentid"), 10, 64)

	c := model.Comments{}
	response := Response{}.Default()
	if count, err := c.DislikeComment(userid, commentid); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Code = StatusOK
		response.Body["count"] = count
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func UpdateComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var u model.User
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &u)

	var err error
	var resq []byte
	response := Response{}.Default()
	if u, err = u.Save(); err != nil {
		response.Status = http.StatusForbidden
		response.Code = StatusUnauthorized
		response.Message = err.Error()
	} else {
		response.Body["user"] = u
	}

	resq, err = json.Marshal(response)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	w.Write(resq)
}

// GetComments : 查询评论,后台专用API
func GetALlComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	query := parseQuery(r)
	var err error
	var b []byte

	response := Response{}.Default()
	userid := utils.GetUserInfoFromContext(r.Context())
	u := model.User{}

	if u.IsIDAdmin(userid) {
		response.Body["comments"], response.Body["total"] = model.Comments{}.GetAll(query.Page, query.PageSize, query.Field, query.Keyword)
	} else {
		response.Code = StatusNotAcceptable
		response.Message = "只有管理员才可以访问此API"
	}

	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

// GetQuestionComments : 查询评论
func GetQuestionComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	query := parseQuery(r)
	var err error
	var b []byte

	response := Response{}.Default()
	quesid, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	response.Body["comments"], response.Body["total"] = model.Comments{}.GetByQuestion(quesid, 0, query.Page, query.PageSize)
	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

// GetParentComments : 查询评论
func GetChildComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	query := parseQuery(r)
	parent, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	var err error
	var b []byte

	response := Response{}.Default()
	response.Body["comments"], response.Body["total"] = model.Comments{}.GetChilren(parent, query.Page, query.PageSize)
	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

// RemoveComments 删除评论
func DeleteComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		IDS []uint64 `json:"ids"`
	}

	var m parseModel
	var err error
	var b []byte

	userid := utils.GetUserInfoFromContext(r.Context())

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)

	comment := model.Comments{}
	u := model.User{}
	response := Response{}.Default()
	for _, id := range m.IDS {
		if comment.IsCreator(userid, id) || u.IsIDAdmin(userid) {
			if err := comment.Remove(id); err == nil {
				response.Message = "删除成功"
			} else {
				response.Code = StatusNotAcceptable
				response.Message = "无权删除，您非评论创建者或者管理员，请停止操作"
			}
		} else {
			response.Code = StatusNotAcceptable
			response.Message = "无权删除，您非题库创建者，请停止操作"
		}
	}
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
