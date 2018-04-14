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

const DefaultPageSize uint64 = 10

func ChangeAvatar(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	response := Response{}.Default()
	r.ParseMultipartForm(32 << 20)
	imageInfo, err := uploadImage(r, "avatar", "picture")

	if err != nil {
		response.Status = http.StatusNotAcceptable
		response.Code = StatusNotAcceptable
		response.Message = "上传图片失败，请重新上传."
	} else {
		user := model.User{
			ID:     utils.GetUserInfoFromContext(r.Context()),
			Avatar: imageInfo.Url,
		}
		if err = user.SetAvatar(); err != nil {
			response.Status = http.StatusNotAcceptable
			response.Code = StatusNotAcceptable
			response.Message = "更新头像失败，错误信息：" + err.Error()
		} else {
			response.Body["url"] = imageInfo.Url
			response.Message = "更新头像成功！"
		}
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func ChangePassword(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	response := Response{}.Default()

	type parseModel struct {
		ID          uint64 `json:"id"`
		OldPassword string `json:"oldpassword"`
		NewPassword string `json:"newpassword"`
	}
	var m parseModel
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	user, _ := model.User{ID: utils.GetUserInfoFromContext(r.Context())}.Get()
	if result := user.CheckPassword(m.OldPassword); result != true {
		response.Status = http.StatusNotAcceptable
		response.Code = StatusNotAcceptable
		response.Message = "旧密码验证不通过，请输入正确的密码"
	} else {
		if err = user.ResetPassword(user.Email, m.NewPassword); err != nil {
			response.Status = http.StatusNotAcceptable
			response.Code = StatusNotAcceptable
			response.Message = "修改密码失败，错误信息：" + err.Error()
		} else {
			response.Message = "成功更新密码！"
		}
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func SaveUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var u model.User
	var err error
	var resq []byte

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	err = json.Unmarshal([]byte(result), &u)

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

func DeleteUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		IDS []uint64 `json:"ids"`
	}
	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)

	user := model.User{}

	response := Response{}.Default()

	if err := user.Remove(m.IDS); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

//todo : 用这个函数替代 SaveUser
// func UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	var u model.User
// 	result, _ := ioutil.ReadAll(r.Body)
// 	r.Body.Close()
// 	json.Unmarshal([]byte(result), &u)
// 	fmt.Printf("user : %v \n", r.Body)
// 	var err error
// 	var resq []byte
// 	response := Response{}.Default()
// 	if u, err = u.Save(); err != nil {
// 		response.Status = http.StatusForbidden
// 		response.Code = StatusUnauthorized
// 		response.Message = err.Error()
// 	} else {
// 		response.Body["user"] = u
// 	}

// 	resq, err = json.Marshal(response)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 		return
// 	}
// 	w.Write(resq)
// }

func GetUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	if id == 0 {
		response.Code = StatusNotAcceptable
		response.Message = "cannot log in "
	} else {
		user := model.User{ID: id}
		response.Body["user"], _ = user.Get()
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)

}

func GetUserRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	response := Response{}.Default()
	userid := utils.GetUserInfoFromContext(r.Context())
	record := model.QuestionRecord{}
	response.Body["records"], err = record.GetByUser(userid)
	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = "无法获取用户记录"
	}
	b, err = json.Marshal(response)

	w.Write(b)
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	userid := utils.GetUserInfoFromContext(r.Context())

	user := model.User{ID: userid}
	response.Body["user"], _ = user.Get()

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
func GetUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	response := Response{}.Default()
	query := parseQuery(r)

	if query.FilterdBy == "group" {
		response.Body["users"], response.Body["total"] = model.User{}.GetAll(query.Page, query.PageSize, query.FilterdBy, query.FilterdID, query.Field, query.Keyword)
	} else if query.FilterdBy == "tag" {
		response.Body["users"], response.Body["total"] = model.User{}.GetAll(query.Page, query.PageSize, query.FilterdBy, query.FilterdID, query.Field, query.Keyword)
	} else {
		response.Body["users"], response.Body["total"] = model.User{}.GetAll(query.Page, query.PageSize, "", 0, query.Field, query.Keyword)
	}

	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}

	w.Write(b)
}

func GetTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	query := parseQuery(r)
	var err error
	var b []byte

	response := Response{}.Default()
	response.Body["tags"], response.Body["total"] = model.Tag{}.GetAll(query.Page, query.PageSize, query.Field, query.Keyword)
	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func SaveTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		ID    uint64   `json:"id"`
		Name  string   `json:"name"`
		Users []uint64 `json:"users"`
	}
	var m parseModel
	var err error

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	tag := model.Tag{ID: m.ID, Name: m.Name}

	response := Response{}.Default()
	if tag, err = tag.Save(m.Users); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Body["tag"] = tag
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func DeleteTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		IDS []uint64 `json:"ids"`
	}
	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)

	tag := model.Tag{}

	response := Response{}.Default()

	if err := tag.Remove(m.IDS); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
func GetTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	withUsers, err := strconv.ParseBool(r.Form["with_users"][0])
	tag := model.Tag{ID: id}
	response.Body["tag"] = tag.Get(withUsers)
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

// ResetPasswordUser : 只有管理员和用户本人才可以重置密码
func ResetPasswordUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte

	response := Response{}.Default()
	currentUserID := utils.GetUserInfoFromContext(r.Context())
	userID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)

	m := model.User{}
	m.ID = currentUserID
	currentUser, _ := m.Get()
	m.ID = userID
	user, _ := m.Get()

	if currentUser.ID == user.ID { //是用户本人
		err = m.ResetPassword(user.Email, "123456")
	} else if currentUser.IsSuper {
		err = m.ResetPassword(user.Email, "123456")
	}

	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = "重置密码出错，错误信息：" + err.Error()
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)

}
