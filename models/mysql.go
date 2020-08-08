package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sansna/expconf/proto"
)

const DEFAULT_DB_NAME = "expconf"

var DB_CONF map[string]*gorm.DB

func GetDbPath(dbname string) string {
	return fmt.Sprintf("root:@tcp(localhost:3306)/%s?charset=utf8&parseTime=True&loc=Local", dbname)
}
func GetCreateDbSql(dbname string) string {
	return "create database if not exists " + dbname + " character set utf8 collate utf8_bin;"
}

func CreateDB(conn *gorm.DB, dbname string) *gorm.DB {
	sqlstr := GetCreateDbSql(dbname)
	err := conn.Exec(sqlstr).Error
	if err != nil {
		fmt.Println("fail create.", dbname, err)
		return nil
	}
	urlpath := GetDbPath(dbname)
	db, _ := gorm.Open("mysql", urlpath)
	return db
}

func AppendDBCONF(dbname string, db *gorm.DB) {
	fmt.Println(dbname)
	db.DB().SetMaxOpenConns(4)
	db.DB().SetMaxIdleConns(1)
	db.DB().SetConnMaxLifetime(time.Hour)
	DB_CONF[dbname] = db
	return
}

// input: information_schema: output: default
// input: default: output: default
// input: dbname output: dbname
func ConnectOrCreateAndConnect(dbname string) *gorm.DB {
	if DB_CONF == nil {
		DB_CONF = make(map[string]*gorm.DB)
	}
	// 有直接返回
	if db, ok := DB_CONF[dbname]; ok {
		return db
	}
	// 尝试直接连接
	urlpath := GetDbPath(dbname)
	db, err := gorm.Open("mysql", urlpath)
	if err != nil {
		fmt.Println("fail direct conn ", dbname)
	} else if db != nil {
		// 连接成功
		AppendDBCONF(dbname, db)
		if dbname == "information_schema" {
			// 创建default，并返回default连接
			db = CreateDB(db, DEFAULT_DB_NAME)
			if db != nil {
				AppendDBCONF(DEFAULT_DB_NAME, db)
			} else {
				return nil
			}
		}
		return db
	}

	// 其他正式db情况创建
	if dbname != DEFAULT_DB_NAME {
		db := DB_CONF[DEFAULT_DB_NAME]
		if db == nil {
			db = ConnectOrCreateAndConnect("information_schema")
		}
		if db == nil {
			fmt.Println("fail conn information_schema db for create database")
			return nil
		}
		db = CreateDB(db, dbname)
		if db != nil {
			AppendDBCONF(dbname, db)
			return db
		} else {
			return nil
		}
	} else {
		// 连接默认db情况
		return ConnectOrCreateAndConnect("information_schema")
	}

	//return nil
}

// 每个app_env一个db
// 仅default和业务db需要init
func InitDB(dbname string) *gorm.DB {
	db := ConnectOrCreateAndConnect(dbname)

	if dbname != DEFAULT_DB_NAME {
		// 其他从属库需要创建表
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
	}

	return db
}

func GetConn(dbname string) *gorm.DB {
	if db, ok := DB_CONF[dbname]; ok {
		return db
	} else {
		return InitDB(dbname)
	}
}
