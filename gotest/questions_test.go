package gotest

import (
	"testing"

	"yeetikuserver/model"

	"../handler"
)

func Test_GetUserFavorites_3(t *testing.T) {
	f := &model.Question{}
	if result, total, err := f.GetUserFavorites(13, 1, 10); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("根据用户ID获取收藏题目测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Logf("%d : %+v", total, result)
		t.Log("根据用户ID获取收藏题目测试通过") //记录一些你期望记录的信息
	}
}

func Test_GetUserWrong_1(t *testing.T) {
	f := &model.Question{}
	if result, total, err := f.GetUserWrong(13, 1, 10); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("根据用户ID获取错误题目测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Logf("%d : %+v", total, result)
		t.Log("根据用户ID获取错误题目测试通过") //记录一些你期望记录的信息
	}
}

func Test_SaveExcelToDB_1(t *testing.T) {
	file := "../upload/file/questions/test.xlsx"

	handler.SaveExcelToDB(13, file)

}
