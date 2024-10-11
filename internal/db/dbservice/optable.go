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

func (t *TestDB) CreateTable() {
	table := &model.Student{}
	err := t.AutoMigrate(table)
	if err != nil {
		logger.LogrusObj.Error("Create Table Error: ", err.Error())
	}
}

func (t *TestDB) CreateRecord(name string, score float64) (err error) {
	record := model.Student{
		Name:  name,
		Score: score,
	}
	if err = t.Model(&model.Student{}).Create(&record).Error; err != nil {
		logger.LogrusObj.Error("Insert User Error: ", err.Error())
		return
	}
	return
}
