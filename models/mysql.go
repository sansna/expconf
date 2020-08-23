package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sansna/expconf/proto"
	"github.com/sansna/expconf/utils/logger"
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
	fun := "CreateDB"
	sqlstr := GetCreateDbSql(dbname)
	err := conn.Exec(sqlstr).Error
	if err != nil {
		logger.Errorf("%s: fail create: %v, err: %v", fun, dbname, err)
		return nil
	}
	urlpath := GetDbPath(dbname)
	db, _ := gorm.Open("mysql", urlpath)
	return db
}

func AppendDBCONF(dbname string, db *gorm.DB) {
	fun := "AppendDBCONF"
	logger.Debugf("%s: name: %v", fun, dbname)
	db.DB().SetMaxOpenConns(4)
	db.DB().SetMaxIdleConns(1)
	db.DB().SetConnMaxLifetime(time.Hour)
	DB_CONF[dbname] = db
	return
}

// input: information_schema: output: default
// input: default: output: default
// input: dbname output: dbname
func ConnectOrCreateAndConnect(dbname string) (*gorm.DB, bool) {
	fun := "ConnectOrCreateAndConnect"
	created := false

	if DB_CONF == nil {
		DB_CONF = make(map[string]*gorm.DB)
	}
	// 有直接返回
	if db, ok := DB_CONF[dbname]; ok {
		return db, created
	}
	// 尝试直接连接
	urlpath := GetDbPath(dbname)
	db, err := gorm.Open("mysql", urlpath)
	if err != nil {
		logger.Infof("%s: fail direct con: %v", fun, dbname)
	} else if db != nil {
		// 连接成功
		AppendDBCONF(dbname, db)
		if dbname == "information_schema" {
			// 创建default，并返回default连接
			db = CreateDB(db, DEFAULT_DB_NAME)
			if db != nil {
				AppendDBCONF(DEFAULT_DB_NAME, db)
			} else {
				return nil, created
			}
		}
		return db, created
	}

	// 其他正式db情况创建
	if dbname != DEFAULT_DB_NAME {
		db := DB_CONF[DEFAULT_DB_NAME]
		if db == nil {
			db, created = ConnectOrCreateAndConnect("information_schema")
		}
		if db == nil {
			logger.Errorf("fail conn information_schema db for create database")
			return nil, created
		}
		db = CreateDB(db, dbname)
		if db != nil {
			created = true
			AppendDBCONF(dbname, db)
			return db, created
		} else {
			return nil, created
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
	db, created := ConnectOrCreateAndConnect(dbname)

	if dbname != DEFAULT_DB_NAME && created {
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
