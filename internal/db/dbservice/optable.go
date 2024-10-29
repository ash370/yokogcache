package dbservice

import (
	"context"
	"yokogcache/internal/db/model"
	"yokogcache/utils/logger"

	"gorm.io/gorm"
)

type TestDB struct {
	*gorm.DB
}

func NewStudentDB(ctx context.Context) *TestDB {
	return &TestDB{NewClient(ctx)}
}

func (t *TestDB) CreateTable(table *model.Student) {
	err := t.AutoMigrate(table)
	if err != nil {
		logger.LogrusObj.Error("Create Table Error: ", err.Error())
	}
}

func (t *TestDB) CreateRecord(r *[]model.Student) (err error) {
	if err = t.Create(r).Error; err != nil {
		logger.LogrusObj.Error("Insert User Error: ", err.Error())
		return
	}
	return
}

func (t *TestDB) Load(key string) ([]byte, error) {
	var record model.Student
	if err := t.Where("name=?", key).Find(&record).Error; err != nil {
		logger.LogrusObj.Error("Load Table Error: ", err.Error())
		return nil, err
	} else {
		return []byte(record.Score), nil
	}
}
