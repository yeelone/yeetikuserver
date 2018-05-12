package model

import (
	"encoding/json"
	"time"
)

// BankRecords :
type BankRecords struct {
	ID          uint64    `json:"id" gorm:"primary_key"`
	BankID      uint64    `json:"bank_id"`
	UserID      uint64    `json:"user_id"`
	User        User      `json:"user" gorm:"-"`
	LatestIndex int       `json:"latest"` //记录最后做的题
	CreatedAt   time.Time `json:"create_time"`
	UpdatedAt   time.Time `json:"update_time"`
}

// Get :
func (r BankRecords) Get() (record BankRecords) {
	mydb.Where(&BankRecords{BankID: r.BankID, UserID: r.UserID}).First(&record)
	return record
}

// GetAll :
func (r BankRecords) GetAll(page, pageSize uint64) (records []BankRecords, total uint64) {
	var offset = (page - 1) * pageSize
	mydb.Model(&r).Offset(offset).Limit(pageSize).Where(&BankRecords{BankID: r.BankID}).First(&records)
	mydb.Model(&r).Where(&BankRecords{BankID: r.BankID}).Count(&total)

	for index, item := range records {
		if item.UserID > 0 {
			records[index].User, _ = User{ID: item.UserID}.Get()
		}
	}
	return records, total
}

// GetByUser :
func (r BankRecords) GetByUser(page, pageSize uint64) (records []BankRecords, total int, err error) {
	var offset = (page - 1) * pageSize
	mydb.Model(&r).Offset(offset).Limit(pageSize).Order("id").Where(&BankRecords{UserID: r.UserID}).Find(&records)
	mydb.Model(&r).Where(&BankRecords{UserID: r.UserID}).Count(&total)
	return records, total, nil
}

// Insert :
func (r BankRecords) Insert() (err error) {
	record := BankRecords{}
	tx := mydb.Begin()
	mydb.Where(&BankRecords{BankID: r.BankID, UserID: r.UserID}).First(&record)
	if record.ID > 0 {
		record.LatestIndex = r.LatestIndex
		err = tx.Save(&record).Error
	} else {
		err = tx.Create(&r).Error
	}
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// MarshalJSON :
func (r *BankRecords) MarshalJSON() ([]byte, error) {
	type Alias BankRecords
	return json.Marshal(&struct {
		CreatedAt string `json:"create_time"`
		UpdatedAt string `json:"update_time"`
		*Alias
	}{
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: r.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(r),
	})
}
