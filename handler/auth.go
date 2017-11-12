package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"../db"
	"../model"
	"../utils"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func AdminLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var u model.User
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &u)

	var err error
	var b []byte
	if u.IsEmailAdmin(u.Email) {
		if u.Valid(u.Email, u.Password) {
			response := Response{}.Default()
			response.Body["user"] = u
			response.Body["username"] = "elone"
			response.Body["id"] = u.ID
			response.Token = utils.SetJWTToken(u.ID)
			b, err = json.Marshal(response)

			//保存session
			err := db.KVManager{}.Set(utils.Uint2Str(u.ID), response.Token)
			if err != nil {
				log.Warn("保存session出现错误 : " + err.Error())
			}
			log.WithFields(log.Fields{
				"email": u.Email,
			}).Info("登录后台系统")

		} else {
			b, err = json.Marshal(&Response{
				http.StatusForbidden,
				StatusUnauthorized,
				nil,
				"", //token is empty
				"cann't login ",
				1,
				1,
			})
		}
	} else {
		log.WithFields(log.Fields{
			"email": u.Email,
		}).Warn("正在尝试登录后台系统")
	}

	if err != nil {
		log.WithFields(log.Fields{
			"email": u.Email,
		}).Warn("登录失败，错误信息 : " + err.Error())
	}
	w.Write(b)
}

func Login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte

	u := &model.User{}
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), u)

	if u.Valid(u.Email, u.Password) {
		response := Response{}.Default()
		response.Body["user"] = u
		response.Body["username"] = "elone"
		response.Body["id"] = u.ID
		response.Token = utils.SetJWTToken(u.ID)
		b, err = json.Marshal(response)

		fmt.Printf("json b %s \n ", b)
		//保存session
		err := db.KVManager{}.Set(utils.Uint2Str(u.ID), response.Token)
		if err != nil {
			log.Warn("保存session出现错误 : " + err.Error())
		}

	} else {
		b, err = json.Marshal(&Response{
			http.StatusForbidden,
			StatusUnauthorized,
			nil,
			"", //token is empty
			"cann't login ",
			1,
			1,
		})
	}

	if err != nil {
		log.WithFields(log.Fields{
			"email": u.Email,
		}).Warn("登录失败，错误信息 : " + err.Error())
	}
	w.Write(b)
}

func Logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	ctx := r.Context()
	userid := utils.GetUserInfoFromContext(ctx)
	//保存session
	err = db.KVManager{}.Delete(utils.Uint2Str(userid))
	response := Response{}.Default()
	response.Status = http.StatusUnauthorized
	response.Code = StatusUnauthorized
	response.Message = "need login again!"

	b, err = json.Marshal(response)
	if err != nil {
		log.WithFields(log.Fields{
			"UserID": userid,
		}).Warn("注销失败，错误信息 : " + err.Error())
	}

	w.Write(b)
}

func Register(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var u model.User
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &u)

	var err error
	var resq []byte

	if u, err = u.Save(); err != nil {
		resq, err = json.Marshal(&Response{
			http.StatusForbidden,
			StatusUnauthorized,
			nil,
			"",
			"cann't create user ",
			1,
			1, // errors
		})
	} else {
		response := Response{}.Default()
		response.Body["username"] = "username"
		response.Token = utils.SetJWTToken(u.ID)
		resq, err = json.Marshal(response)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"Email": u.Email,
		}).Warn("注册出错，错误信息 : " + err.Error())
		return
	}
	w.Write(resq)

}
