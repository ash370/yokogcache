package main

import (
	"context"
	"yokogcache/internal/db/dbservice"
	"yokogcache/internal/db/model"
)

//初始化数据库，多次测试只需执行一次

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db := dbservice.NewStudentDB(ctx)
	db.CreateTable(&model.Student{})

	name := []string{
		"Ella Robinson", "Alexander Williams", "James Franklin",
	}
	score := []string{
		"67.8", "56.5", "34",
	}

	var records []model.Student
	for i, v := range name {
		records = append(records, model.Student{})
		records[i].Name = v
	}
	for i, v := range score {
		records[i].Score = v
	}
	db.CreateRecord(&records)
}
