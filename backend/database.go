package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	dsn := "root:password@tcp(db:3306)/circle?charset=utf8mb4&parseTime=True&loc=Local"
	fmt.Println("Connecting to dsn", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	Migrate(db)
	return db
}

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&TestTable{})
}
