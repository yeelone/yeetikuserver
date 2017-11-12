// exceltest.go
package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/tealeg/xlsx"
)

type TestAccount struct {
	ID      uint64  `json:"id" gorm:"primary_key"`
	Address string  `json:"address"  ` // gorm:"unique_index:idx_address_name"
	Name    string  `json:"name" gorm:"unique_index:idx_name"`
	Count   int64   `json:"count"`
	Amount  float64 `json:"amount"`
	Account string  `json:"account"`
}

var once sync.Once
var dbInstance *gorm.DB = nil

func GetInstance() *gorm.DB {
	once.Do(func() {
		if dbInstance == nil {
			var err error
			dbInstance, err = gorm.Open("postgres", "host=localhost user=elone dbname=elone password=123456 sslmode=disable ")
			if err != nil {
				panic("failed to connect database")
			}
		}
	})
	return dbInstance

}

var db *gorm.DB

//想批量保存几千条数据到数据库，先分段插入，如果有错误的，再继续分段 ，直到判断出哪些行出错，哪些没有出错
var errorRows []*xlsx.Row

func main() {
	db = GetInstance()
	var test_account TestAccount
	db.AutoMigrate(&test_account)

	excelFileName := "测试数据.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Printf("error", err)
	}
	sheet := xlFile.Sheets[0]

	//记录开始时间
	start := time.Now()

	// importToDB2(sheet.Rows[5:1413])
	importToDB3(sheet.Rows[5:10464])
	end := time.Now()

	//输出执行时间，单位为毫秒。
	fmt.Println("运行时间：", end.Sub(start))
	fmt.Printf("error -- rows %d \n ", len(errorRows))
}

//递归
func importToDB(rows []*xlsx.Row) {
	sqlStr := "INSERT INTO test_accounts( address,name ,count ,amount,account ) VALUES "
	var sqlArr []string

	for _, row := range rows {
		sqlArr = append(sqlArr, `( '`+row.Cells[1].String()+`'  , '`+row.Cells[2].String()+`' , `+row.Cells[3].String()+` , `+row.Cells[4].String()+` , `+row.Cells[5].String()+` )`)
	}

	s := strings.Join(sqlArr, ",")
	if err := db.Exec(sqlStr + s).Error; err != nil {
		length := len(rows)
		if length > 1 {
			end := length / 2
			if end >= 0 {
				importToDB(rows[0:end])
				importToDB(rows[end:length])
			}
		}

		if length == 1 {
			errorRows = append(errorRows, rows[0])
		}
	}

}

//改写迭代
func importToDB2(rows []*xlsx.Row) {
	sqlStr := "INSERT INTO test_accounts( address,name ,count ,amount,account ) VALUES "

	distance := 1 //测试结果，越小越快
	start := 0
	end := distance

	length := len(rows)
	if end > length {
		end = length
	}

	for end <= length {
		rs := rows[start:end]
		var sqlArr []string

		for _, row := range rs {
			sqlArr = append(sqlArr, `( '`+row.Cells[1].String()+`'  , '`+row.Cells[2].String()+`' , `+row.Cells[3].String()+` , `+row.Cells[4].String()+` , `+row.Cells[5].String()+` )`)
		}

		s := strings.Join(sqlArr, ",")
		if err := db.Exec(sqlStr + s).Error; err != nil {
			if (end - start) == 1 {
				//剩下最后一个的时候就证明是这个出现错误。记录起错误行，然后将start往前移动1，继续执行后面的操作
				errorRows = append(errorRows, rs[0])
				start += 1
				end = start + distance
			} else {
				//每次都取一半的数据进行查找错误行
				end = (end - start) / 2
			}
		} else {
			start = end
			end += distance
		}

		if end > length {
			end = length
		}

		if end <= start {
			end = start + 1
		}
	}
}

func importToDB3(rows []*xlsx.Row) {

	for _, row := range rows {
		count, _ := row.Cells[3].Int64()
		amount, _ := row.Cells[4].Float()
		m := TestAccount{
			Address: row.Cells[1].String(),
			Name:    row.Cells[2].String(),
			Count:   count,
			Amount:  amount,
			Account: row.Cells[5].String(),
		}

		db.Create(&m)

	}
}
