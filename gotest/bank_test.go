package gotest

import (
	"testing"

	"../model"
)

func Test_GetByUser_1(t *testing.T) {
	g := &model.Bank{ID: 1}
	if result, _, err := g.GetByUser(1, 10, 13); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("根据用记ID获取题库没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log(result)
		t.Log("根据用记ID获取题库测试通过") //记录一些你期望记录的信息
	}
}

func Test_Delete_1(t *testing.T) {
	btags := &model.Btags{}
	if err := btags.Delete(78); err != nil {
		t.Error(err)
		t.Error("删除失败") // 如果不是如预期的那么就报错
	} else {
		t.Log("删除成功") //记录一些你期望记录的信息
	}
}

func Test_GetChild_2(t *testing.T) {
	btags := &model.Btags{}
	ids := btags.GetChild(78)
	t.Logf("%+v \n ", ids)
	t.Log("get ids ") //记录一些你期望记录的信息
}
