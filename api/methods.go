package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sansna/expconf/models"
	"github.com/sansna/expconf/proto"
)

func GetDbName(app string, env string) string {
	if strings.Contains(app, "$") || strings.Contains(env, "$") {
		return ""
	}
	if len(app) == 0 || len(env) == 0 {
		return "app$env"
	}
	return app + "$" + env
}

// retval: > 0: got column
// = 0: not found.
func GetGroupOrAdd(conn *gorm.DB, tid int64, name string) int64 {
	gst := proto.GroupSt{}
	err := conn.Where("id=? and del = false", tid).Find(&gst).Error
	if err != nil {
		fmt.Println(err)
	}
	if gst.Id > 0 {
		return gst.Id
	}

	gst.Et = 0
	gst.St = 0
	gst.Ut = time.Now().Unix()
	gst.Name = name
	err = conn.Create(&gst).Error
	if err != nil {
		return 0
	}
	return gst.Id
}

// retval: false: do conflict
// true: no conflict
func FindNoConflictRecord(conn *gorm.DB, r *proto.OneModifySt) bool {
	other := proto.RecordSt{}
	//conn.Limit(1).Find(&other, conn.Where())
	conn.Where("tid=? and `key`=? and ((st < ? and ? < et) or (st < ? and ? < et))", r.Tid, r.Key, r.Et, r.Et, r.St, r.St).Limit(1).Find(&other)
	fmt.Println("found other as dup", other.ID, r.Et, r.St, other.Et, other.St)
	if len(other.Key) > 0 {
		return false
	}
	return true
}

func AddConfig(param *proto.AddConfigParam) (err error) {
	//fun := "AddConfig"

	if param.St < 0 || param.Et < 0 || len(param.Val) == 0 || param.Tid < 0 || len(param.Key) == 0 {
		return nil
	}
	if param.St >= param.Et && param.Et > 0 {
		return nil
	}
	if param.Tid == 0 && len(param.ExpName) == 0 {
		return nil
	}
	param.Ut = time.Now().Unix()

	dbname := GetDbName(param.App, param.Env)
	if len(dbname) > 0 {
		db := models.GetConn(dbname)
		fmt.Println(models.DB_CONF)
		// 开始事务
		db = db.Begin()
		defer func() {
			if r := recover(); r != nil {
				db.Rollback()
			} else if err != nil {
				db.Rollback()
			}
		}()
		if db == nil {
			return errors.New(" no conn get:" + dbname)
		}
		tid := GetGroupOrAdd(db, param.Tid, param.ExpName)
		if tid <= 0 {
			return errors.New("tid not found.")
		}
		param.OneModifySt.Tid = tid
		if FindNoConflictRecord(db, param.OneModifySt) {
			db.Create(&proto.RecordSt{
				OneModifySt: param.OneModifySt,
			})
		} else {
			return errors.New(" conflict happ.")
		}

		db.Commit()
	} else {
		return errors.New("db name err")
	}
	return nil
}

func GetGroups(param *proto.GetGroupsParam) (data *proto.GetGroupsData, err error) {
	app := param.App
	env := param.Env
	data = &proto.GetGroupsData{
		List: make([]*proto.GetGroupsItem, 0),
	}
	dbname := GetDbName(app, env)
	db := models.GetConn(dbname)
	list := make([]proto.GroupSt, 0)
	db.Find(&list)
	for _, l := range list {
		data.List = append(data.List, &proto.GetGroupsItem{
			GroupSt: &l,
		})
	}
	return
}

func GetConfig(param *proto.GetConfigParam) (data *proto.GetConfigData, err error) {
	app := param.App
	env := param.Env
	tid := param.Tid
	key := param.Key

	now := time.Now().Unix()

	r := proto.RecordSt{}
	dbname := GetDbName(app, env)
	db := models.GetConn(dbname)
	db.Where("tid = ? and `key` = ? and (st < ? < et or et = 0)", tid, key, now).Order("et desc").Find(&r)
	data = &proto.GetConfigData{
		RecordSt: &r,
	}

	fmt.Println(tid, key, now, r)
	return
}
