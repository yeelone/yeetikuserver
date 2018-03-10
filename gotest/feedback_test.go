package gotest

import (
	"testing"
	"yeetikuserver/model"
)

func Test_Save_1(t *testing.T) {
	user := model.User{ID: 1}
	feedback := &model.Feedback{
		User:    user,
		Content: "this is test feedback",
	}
	if err := feedback.Save(); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("添加反馈数据测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("添加反馈数据测试通过") //记录一些你期望记录的信息
	}
}
