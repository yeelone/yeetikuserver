package db

import (
	"fmt"
	"log"
	"sync"

	"github.com/boltdb/bolt"
)

var synconce sync.Once
var kdvInstance *bolt.DB
var SESSIONNAME = []byte("sessions")

type KVManager struct {
}

func GetKVInstance() *KVManager {
	return &KVManager{}
}

func (s KVManager) Init() {

	synconce.Do(func() {
		if kdvInstance == nil {
			var err error
			kdvInstance, err = bolt.Open("my.db", 0600, nil)
			if err != nil {
				panic("failed to connect kv database")
			}
		}
	})
}

func (s KVManager) Set(key, value string) error {
	err := kdvInstance.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(SESSIONNAME)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(value))
	})
	return err
}

func (s KVManager) Delete(key string) (err error) {
	if err = kdvInstance.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(SESSIONNAME).Delete([]byte(key))
	}); err != nil {
		log.Fatal(err)
	}
	return err
}

func (s KVManager) Get(id string) (value []byte, err error) {
	err = kdvInstance.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(SESSIONNAME)
		if b == nil {
			return fmt.Errorf("Bucket %q not found!", SESSIONNAME)
		}

		value = b.Get([]byte(id))
		return nil
	})
	return value, nil
}
