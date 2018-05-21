package db

import (
	"fmt"
	"log"
	"sync"

	"github.com/boltdb/bolt"
)

var synconce sync.Once
var kdvInstance *bolt.DB

// KVManager : boltdb 管理器
type KVManager struct {
}

// GetKVInstance 获取boltdb实例
func GetKVInstance() *KVManager {
	return &KVManager{}
}

// Init 初始化 boltdb
func (s KVManager) Init() {

	synconce.Do(func() {
		if kdvInstance == nil {
			var err error
			kdvInstance, err = bolt.Open("yeetiku.db", 0600, nil)
			if err != nil {
				panic("failed to connect kv database")
			}
		}
	})
}

// Set 设置键值
// @params bucketname  数据表名
// @params key
// @params value
func (s KVManager) Set(bucketName []byte, key, value string) error {
	err := kdvInstance.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(value))
	})
	return err
}

// Delete 删除键
// @params bucketName
// key
func (s KVManager) Delete(bucketName []byte, key string) (err error) {
	if err = kdvInstance.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketName).Delete([]byte(key))
	}); err != nil {
		log.Fatal(err)
	}
	return err
}

// Get 获取值
// @params bucketname
// @params key
func (s KVManager) Get(bucketName []byte, key string) (value []byte, err error) {
	err = kdvInstance.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucketName)
		}

		value = b.Get([]byte(key))
		return nil
	})
	return value, err
}
