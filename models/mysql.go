package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sansna/expconf/proto"
)

var DB_CONF map[string]*gorm.DB

// 每个app_env一个db
func InitDB(dbname string) {
	if DB_CONF == nil {
		// 创建新的空db链接
	}
	urlpath := fmt.Sprintf("root:@tcp(localhost:3306)/%s?charset=utf8&parseTime=True&loc=Local", dbname)
	db, err := gorm.Open("mysql", urlpath)
	if err != nil {
		fmt.Println(err)
		return
	}
	//defer db.Close()
	db.DB().SetMaxOpenConns(4)
	db.DB().SetMaxIdleConns(1)
	db.DB().SetConnMaxLifetime(time.Hour)

	db.CreateTable(proto.RecordSt{
		OneModifySt: &proto.OneModifySt{
			Tid: 1,
		},
	}).AddUniqueIndex("uix_uk_r", "tid", "key", "et")
	db.CreateTable(proto.HistoryRecordSt{
		OneModifySt: &proto.OneModifySt{
			Tid: 1,
			Key: "xx",
		},
	}).AddUniqueIndex("uix_uk_h", "tid", "key", "et", "ut")
	db.CreateTable(proto.GroupSt{
		Name: "sdf",
	}).AddUniqueIndex("uix_uk_g", "name", "et")

	if DB_CONF == nil {
		DB_CONF = make(map[string]*gorm.DB)
	}
	DB_CONF[dbname] = db
	return
}

func GetConn(dbname string) *gorm.DB {
	if db, ok := DB_CONF[dbname]; ok {
		return db
	} else {
		// 初始化新db
		err := DB_CONF["default"].Exec("CREATE DATABASE ? CHARACTER SET utf8 collate utf8_bin;", dbname).Error
		if err != nil {
			fmt.Println(err, dbname)
		}
		new_url_path := fmt.Sprintf("root:@tcp(localhost:3306)/%s?charset=utf8&parseTime=True&loc=Local", dbname)
		newdb, err := gorm.Open("mysql", new_url_path)
		if err != nil {
			fmt.Println("fail open mysql.", new_url_path)
			return nil
		}
		DB_CONF[dbname] = newdb
		InitDB(dbname)
		return newdb
	}
}
