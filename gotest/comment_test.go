package gotest

import (
	"testing"

	"yeetikuserver/model"
)

func Test_GetAllParent_1(t *testing.T) {
	m := &model.Comments{}
	result, _ := m.GetAllParent(1, 10)

	if result != nil {
		t.Log(len(result))
	}

}
