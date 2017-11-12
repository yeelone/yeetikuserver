package gotest

import (
	"testing"
	"../model"
)
func Test_GetRelatedUsers_1(t *testing.T) {
	g := &model.Group{ID:1}
	if result,err  := g.GetRelatedUsers(); err != nil { //try a unit test on function
		t.Error(err)
		t.Error("获取组关联用户测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log(result)
		t.Log("获取组关联用户测试通过") //记录一些你期望记录的信息
	}
}
