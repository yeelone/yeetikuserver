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
	log "github.com/sirupsen/logrus"
)

type parseModel struct {
	ID          uint64      `json:"id"`
	Limit       bool        `json:"limit"`
	Name        string      `json:"name"`
	Allow       model.Allow `json:"allow"`
	Total       int         `json:"total"`
	Description string      `json:"description"`
	Image       string      `json:"image"`
	Disable     bool        `json:"disable"`
}

func parseQueryData(r *http.Request) parseModel {
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var m parseModel
	json.Unmarshal([]byte(result), &m)
	return m
}

func GetBank(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	bank := model.Bank{ID: id}
	response.Body["bank"] = bank.Get()
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetRelatedQuestions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	query := parseQuery(r)
	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	bank := model.Bank{ID: id}
	response.Body["questions"], response.Body["total"] = bank.GetRelatedQuestions(query.Start, query.Page, query.PageSize)
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func CreateBank(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := utils.GetUserInfoFromContext(r.Context())

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var m parseModel
	json.Unmarshal([]byte(result), &m)

	bank := &model.Bank{}
	bank.ID = m.ID
	bank.Creator = userID
	bank.Description = m.Description
	bank.Name = m.Name
	bank.Limited = m.Limit
	bank.AllowType = m.Allow.Type
	bank.Image = m.Image

	response := Response{}.Default()
	if err := bank.Save(m.Allow.Keys); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = "Create testbank failed!"
	} else {
		response.Body["id"] = bank.ID
	}

	b, err := json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	query := parseQuery(r)
	var err error
	var b []byte

	response := Response{}.Default()
	userid := utils.GetUserInfoFromContext(r.Context())
	bankID, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	bank := model.BankRecords{BankID: bankID, UserID: userid}
	response.Body["records"], response.Body["total"] = bank.GetAll(query.Page, query.PageSize)

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

//查询用户在题库下练习的所有记录，包括练习数量 ，做错题的数量 ，收藏
func QueryUserRecords(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	response := Response{}.Default()
	userid := utils.GetUserInfoFromContext(r.Context())
	bankid, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	record := model.QuestionRecord{}
	response.Body["total"], response.Body["wrong"] = record.CountByBankID(bankid, userid)

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func UpdateBank(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userid := utils.GetUserInfoFromContext(r.Context())
	query := parseQueryData(r)
	bank := &model.Bank{}
	bank.ID = query.ID
	bank.Creator = userid
	bank.Name = query.Name
	bank.Description = query.Description
	bank.Image = query.Image
	bank.Disable = query.Disable

	response := Response{}.Default()
	if err := bank.Update(); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Body["bank"] = bank.Get()
	}

	b, err := json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetBanks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte
	query := parseQuery(r)
	response := Response{}.Default()

	userID := utils.GetUserInfoFromContext(r.Context())
	u := model.User{}
	if u.IsIDAdmin(userID) {
		//admin can get all banks
		response.Body["banks"], response.Body["total"] = model.Bank{}.GetAll(query.Page, query.PageSize, query.Field, query.Keyword)
	} else {
		//other users can only get the enable banks
		response.Body["banks"], response.Body["total"] = model.Bank{}.GetAllEnable(query.Page, query.PageSize, query.Field, query.Keyword)
	}

	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetUserBanks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	r.ParseForm()
	response := Response{}.Default()
	userID := utils.GetUserInfoFromContext(r.Context())
	query := parseQuery(r)
	fmt.Printf("query %+v \n ", query)
	response.Body["banks"], response.Body["total"], err = model.Bank{}.GetByUser(query.Page, query.PageSize, userID)
	fmt.Printf("response.Body.banks, %+v \n ", response.Body["banks"])
	b, err = json.Marshal(response)

	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = "无法获取用户题库:" + err.Error()
	}
	w.Write(b)
}

func RemoveBank(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		ID uint64 `json:"id"`
	}

	var m parseModel
	var err error
	var b []byte

	userid := utils.GetUserInfoFromContext(r.Context())

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)

	bank := model.Bank{}

	response := Response{}.Default()
	if bank.IsCreator(userid, m.ID) {
		if err := bank.Remove(m.ID); err == nil {
			response.Message = "删除成功"
		} else {
			response.Code = StatusNotAcceptable
			response.Message = "无权删除，您非题库创建者，请停止操作"
		}
	} else {
		response.Code = StatusNotAcceptable
		response.Message = "无权删除，您非题库创建者，请停止操作"
	}

	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func AddRelateQuestions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	type parseModel struct {
		Questions []uint64 `json:"questions"`
	}
	var m parseModel

	var err error
	var b []byte

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	bank := model.Bank{ID: id}
	if err = bank.SaveQuestions(m.Questions); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func RemoveRelatedQuestions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	type parseModel struct {
		Questions []uint64 `json:"questions"`
	}
	var m parseModel
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	response := Response{}.Default()
	if len(m.Questions) > 0 {
		id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
		bank := model.Bank{ID: id}

		if err := bank.DeleteQuestions(m.Questions); err != nil {
			response.Code = StatusNotAcceptable
			response.Message = err.Error()
		}
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func ChangeStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		ID     uint64 `json:"id"`
		Status string `json:"status"`
	}

	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)

	bank := model.Bank{}
	response := Response{}.Default()
	if m.Status == "disable" {
		if err := bank.SetDisable(m.ID); err != nil {
			response.Message = "禁用成功"
		} else {
			response.Code = StatusNotAcceptable
			response.Message = "无权启用，请停止操作"
		}
	} else if m.Status == "enable" {
		if err := bank.SetEnable(m.ID); err != nil {
			response.Message = "启用成功"
		} else {
			response.Code = StatusNotAcceptable
			response.Message = "无权启用，请停止操作"
		}
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("errors :", err)
	}
	w.Write(b)
}

func UploadBankImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte
	imageInfo, _ := uploadImage(r, "banks", "bank-image")
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)
	bank := model.Bank{ID: id, Image: imageInfo.Url}
	if err = bank.SaveImage(); err != nil {
		fmt.Print(err)
	}
	b, err = json.Marshal(imageInfo)
	if err != nil {
		fmt.Print(err)
	}
	w.Write(b)
}

//由于 EnableBank 和 DisableBank大部分代码相关，所以封装在这里
func changeStatus(w http.ResponseWriter, r *http.Request, isDisable bool) {
	type parseModel struct {
		ID uint64 `json:"id"`
	}

	var m parseModel

	userid := utils.GetUserInfoFromContext(r.Context())

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	json.Unmarshal([]byte(result), &m)

	bank := model.Bank{}
	response := Response{}.Default()
	if bank.IsCreator(userid, m.ID) {
		if isDisable {
			if err := bank.SetDisable(m.ID); err != nil {
				response.Message = "禁用成功"
			} else {
				response.Code = StatusNotAcceptable
				response.Message = "无权启用，请停止操作"
			}
		} else {
			if err := bank.SetEnable(m.ID); err != nil {
				response.Message = "启用成功"
			} else {
				response.Code = StatusNotAcceptable
				response.Message = "无权启用，请停止操作"
			}
		}

	} else {
		response.Code = StatusNotAcceptable
		response.Message = "无权操作"
	}
	b, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("errors :", err)
	}
	w.Write(b)
}

func EnableBank(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	changeStatus(w, r, false)
}

func DisableIank(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	changeStatus(w, r, true)
}

//SaveRelatedBankTags :
func SaveRelatedBankTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		Tag  uint64 `json:"tag"`
		Bank uint64 `json:"bank"`
	}
	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	tag := model.Btags{ID: m.Tag}

	response := Response{}.Default()
	var err error
	if err = tag.RelateBank(m.Bank); err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetBankTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	response := Response{}.Default()
	id, _ := strconv.ParseUint(ps.ByName("id"), 10, 64)

	bank := model.Bank{}
	tags, total, err := bank.GetTags(id)

	if err != nil {
		userid := utils.GetUserInfoFromContext(r.Context())
		log.WithFields(log.Fields{
			"userID": userid,
		}).Info("用户获取题库标签出错, err:" + err.Error())
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Body["tags"] = tags
		response.Body["total"] = total
	}

	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)

}

func RemoveRelatedTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte

	type parseModel struct {
		TagID  uint64 `json:"tag"`
		BankID uint64 `json:"bank"`
	}
	var m parseModel
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	btag := model.Btags{}

	response := Response{}.Default()
	err = btag.RemoveBankRelated(m.BankID, m.TagID)
	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Message = "删除成功"
	}
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func DeleteBankTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var b []byte

	type parseModel struct {
		TagID uint64 `json:"tag"`
	}
	var m parseModel
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	btag := model.Btags{}

	response := Response{}.Default()
	err = btag.Delete(m.TagID)
	if err != nil {
		response.Code = StatusNotAcceptable
		response.Message = err.Error()
	} else {
		response.Message = "删除成功"
	}
	b, err = json.Marshal(response)
	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}

func GetAllBankTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var err error
	var b []byte

	response := Response{}.Default()
	query := parseQuery(r)

	btag := model.Btags{}
	if tags, total, err := btag.GetTagTree(query.Keyword); err != nil {
		response.Body["total"] = 0
	} else {
		response.Body["tags"] = tags
		response.Body["total"] = total
	}

	b, err = json.Marshal(response)
	if err != nil {
		log.Warn("在获取题库标签时出现错误 :" + err.Error())
	}
	w.Write(b)
}

func SaveBankTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type parseModel struct {
		Name   string `json:"name"`
		Parent uint64 `json:"parent"`
	}
	var m parseModel

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	json.Unmarshal([]byte(result), &m)

	tag := model.Btags{Name: m.Name, Parent: m.Parent}

	response := Response{}.Default()
	var err error
	if tag, err = tag.Save(); err != nil {
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

func GetRelatedBanks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	query := parseQuery(r)
	var err error
	var b []byte

	response := Response{}.Default()

	btags := model.Btags{ID: query.Tag}
	response.Body["banks"], response.Body["total"] = btags.GetRelatedBanks(query.Page, query.PageSize)

	b, err = json.Marshal(response)

	if err != nil {
		fmt.Println("errors :", err)
	}
	w.Write(b)
}
