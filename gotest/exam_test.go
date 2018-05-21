package gotest

import (
	"testing"
	"yeetikuserver/model"
)

func Test_RandomCreate_1(t *testing.T) {
	ex := &model.Exam{}
	if result, err := ex.RandomCreate(3, 10); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("根据用户ID随机创建试卷没通过") // 如果不是如预期的那么就报错
	} else {
		t.Logf("%v \n", result)
		t.Log("根据用户ID随机创建试卷测试通过") //记录一些你期望记录的信息
	}
}

func Test_CheckResultAndUpdateScore_1(t *testing.T) {
	ex := &model.Exam{ID: 4}
	if err := ex.CheckResultAndUpdateScore(99); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("更新试卷测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("更新试卷测试通过") //记录一些你期望记录的信息
	}
}

func Test_GetByCreator_1(t *testing.T) {
	ex := model.Exam{Creator: 1}
	if items, total, err := ex.GetByCreator(1, 1, 10); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("获取所有试卷测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Logf("total : %d ; items : %v \n", total, items)
		t.Log("获取所有试卷测试通过") //记录一些你期望记录的信息
	}
}

func Test_Get_2(t *testing.T) {
	ex := model.Exam{ID: 10}
	if _, err := ex.Get(); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("获取所有试卷测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("获取所有试卷测试通过") //记录一些你期望记录的信息
	}
}
