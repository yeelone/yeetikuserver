//message保存着键值对，使用的是boltdb进行存储，主要作用是保存那些不是非常重要的，却不需要永久保存的信息。
package model

import (
	"encoding/json"
	"time"

	db "yeetikuserver/db"
	"yeetikuserver/utils"
)

type QuestionImportMessage struct {
	UserID    uint64 `json:"userID"`
	UserName  string `json:"username"`
	Success   bool   `json:"success"`
	Content   string `json:"content"`
	Unread    int    `json:"unread"` // 1 为未读， 0为已读
	CreatedAt time.Time
}

func (qm QuestionImportMessage) Save() error {
	qm.CreatedAt = time.Now()
	encoded, err := json.Marshal(qm)
	if err != nil {
		return err
	}
	id := utils.Uint2Str(qm.UserID)
	kvdb.Set(db.SESSIONBUCKET, id, string(encoded))

	return nil
}

func (qm QuestionImportMessage) Query() (msg QuestionImportMessage, err error) {
	id := utils.Uint2Str(qm.UserID)
	value, err := kvdb.Get(db.SESSIONBUCKET, id)
	err = json.Unmarshal(value, &msg)
	return msg, err
}

func (qm QuestionImportMessage) Remove() (err error) {
	id := utils.Uint2Str(qm.UserID)
	err = kvdb.Delete(db.SESSIONBUCKET, id)
	return err
}
