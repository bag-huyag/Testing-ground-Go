package config

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db *gorm.DB
)

func Connect() {
	//b, err := gorm.Open("mysql", "Bag:BBQ@12@/simplerest?charset=utf8&parseTime=True&loc=Local")
	d, err := gorm.Open("mysql", "Bag:root@12@tcp(127.0.0.1:3306)/simplerest?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}
	db = d
}

func GetDB() *gorm.DB {
	return db
}
