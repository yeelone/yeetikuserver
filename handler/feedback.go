package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"yeetikuserver/model"
	"yeetikuserver/utils"

	"github.com/julienschmidt/httprouter"
)

func CreateFeedBack(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	response := Response{}.Default()
	r.ParseMultipartForm(32 << 20)
	imageInfo, err := uploadImage(r, "feedback", "picture")

	if err != nil {
		response.Status = http.StatusNotAcceptable
		response.Code = StatusNotAcceptable
		response.Message = "上传图片失败，请重新上传."
	} else {
		userid := utils.GetUserInfoFromContext(r.Context())
		feedback := model.Feedback{
			User:    model.User{ID: userid},
			Content: r.Form["content"][0],
			Contact: r.Form["contact"][0],
			Image:   imageInfo.Url,
		}

		if err = feedback.Save(); err != nil {
			response.Status = http.StatusNotAcceptable
			response.Code = StatusNotAcceptable
			response.Message = "发生错误 : " + err.Error()
		}
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetFeedBacks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()

	query := parseQuery(r)
	feeds := model.Feedback{}

	response.Body["feedbacks"], response.Body["total"] = feeds.GetAll(query.Page, query.PageSize, query.Field, query.Keyword)

	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
