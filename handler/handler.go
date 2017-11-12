package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type Query struct {
	Page       uint64 `json:"page"`
	PageSize   uint64 `json:"pageSize"`
	Start      uint64 `json:"start"`
	CategoryID uint64 `json:"category"`
	Field      string `json:"field"` //用于admin中，/admin/users?field=name&keyword=yeel
	Keyword    string `json:"keyword"`
	FilterdBy  string `json:"filter_by"`
	FilterdID  uint64 `json:"filter_id"`
	Tag        uint64 `json:"tag"`
}

func parseQuery(r *http.Request) (query Query) {

	query.PageSize = DefaultPageSize
	query.Page = 1
	query.Start = 0
	if len(r.Form) > 0 {
		if len(r.Form["page"]) > 0 {
			query.Page, _ = strconv.ParseUint(r.Form["page"][0], 10, 64)
		}

		if len(r.Form["pageSize"]) > 0 {
			query.PageSize, _ = strconv.ParseUint(r.Form["pageSize"][0], 10, 64)
		}

		if len(r.Form["start"]) > 0 {
			query.Start, _ = strconv.ParseUint(r.Form["start"][0], 10, 64)
		}

		if len(r.Form["category"]) > 0 {
			query.CategoryID, _ = strconv.ParseUint(r.Form["category"][0], 10, 64)
		} else {
			query.CategoryID = 0
		}

		if len(r.Form["field"]) > 0 {
			query.Field = r.Form["field"][0]
		}

		if len(r.Form["keyword"]) > 0 {
			query.Keyword = r.Form["keyword"][0]
		}

		if len(r.Form["filter_by"]) > 0 {
			query.FilterdBy = r.Form["filter_by"][0]
		}

		if len(r.Form["filter_id"]) > 0 {
			query.FilterdID, _ = strconv.ParseUint(r.Form["filter_id"][0], 10, 64)
		}

		if len(r.Form["tag"]) > 0 {
			query.Tag, _ = strconv.ParseUint(r.Form["tag"][0], 10, 64)
		}

	}
	return query
}

type Response struct {
	Status    int                    `json:"status"` // always 200
	Code      int                    `json:"code"`
	Body      map[string]interface{} `json:"body"`
	Token     string                 `json:"token"`
	Message   string                 `json:"message"`
	RequestID int                    `json:"requestId"`
	Err       int                    `json:"error"`
}

func (resp Response) Default() Response {
	resp.Status = http.StatusOK
	resp.Code = StatusOK
	resp.Body = make(map[string]interface{})
	resp.Token = ""
	resp.Message = ""
	resp.RequestID = 0
	resp.Err = 0
	return resp
}

func (resp *Response) SetData(interface{}) {

}

func Home(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w, "this is home page")
}
