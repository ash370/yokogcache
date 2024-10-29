package dbservice

import (
	"context"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var _db *gorm.DB

func init() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,   // Default length of String type fields
		DisableDatetimePrecision:  true,  // Disable datetime precision
		DontSupportRenameIndex:    true,  // When renaming an index, delete and create a new one
		DontSupportRenameColumn:   true,  // Rename the column with `change`
		SkipInitializeWithVersion: false, // Automatically configure based on version
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	_db = db
}

func NewClient(ctx context.Context) *gorm.DB {
	db := _db
	return db.WithContext(ctx)
}
