package gotest

import (
	"testing"
	"../model"
)
func Test_GetByUser_2(t *testing.T) {
	g := &model.QuestionRecord{}
	if result,err  := g.GetByUser(13); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("根据用户ID获取练习记录测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Logf("%+v", result)
		t.Log("根据用户ID获取练习记录测试通过") //记录一些你期望记录的信息
	}
}

//todo : 这里应该那家一个测试账号
//func Test_ResetPassword_1(t *testing.T) {
//	u := &model.User{}
//	if err := u.ResetPassword("admin@gmail.com","eee"); err != nil {
//		t.Error(err)
//		t.Error("修改密码测试没通过") // 如果不是如预期的那么就报错
//	}else{
//		t.Error(err)
//		t.Error("修改密码测试已通过") // 如果不是如预期的那么就报错
//	}
//}