package api

import (
	"encoding/json"
	"errors"
	//"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sansna/expconf/models"
	"github.com/sansna/expconf/proto"
	"github.com/sansna/expconf/utils/logger"
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
	fun := "GetGroupOrAdd"
	gst := proto.GroupSt{}
	err := conn.Where("id=? and del = false", tid).Find(&gst).Error
	if err != nil {
		logger.Errorf("%s: fail get %v, err: %v", fun, tid, err)
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

// 如果指定except_et >= 0，则过滤该指定et的配置, < 0情况不过滤
// retval: false: do conflict
// true: no conflict
func FindNoConflictRecord(conn *gorm.DB, r *proto.OneModifySt, except_et int64) bool {
	fun := "FindNoConflictRecord"
	other := proto.RecordSt{}
	if except_et >= 0 {
		conn.Where("tid=? and `key`=? and ((st < ? and ? < et) or (st < ? and ? < et) or (? < st and et < ?)) and et != ?", r.Tid, r.Key, r.Et, r.Et, r.St, r.St, r.St, r.Et, except_et).Limit(1).Find(&other)
	} else {
		conn.Where("tid=? and `key`=? and ((st < ? and ? < et) or (st < ? and ? < et) or (? < st and et < ?))", r.Tid, r.Key, r.Et, r.Et, r.St, r.St, r.St, r.Et).Limit(1).Find(&other)
	}
	logger.Debugf("%s: r: %v, other:%v", fun, r, other)

	if len(other.Key) > 0 {
		return false
	}
	return true
}

func AddConfig(param *proto.AddConfigParam) (err error) {
	fun := "AddConfig"

	dt := param.DataType
	switch dt {
	case "json":
		fallthrough
	default:
		val := param.OneModifySt.Val
		if len(val) > 0 {
			x := struct{}{}
			err := json.Unmarshal([]byte(val), &x)
			if err != nil {
				logger.Errorf("%s: fail err: %v", fun, err)
				return err
			}
		} else {
			param.Val = "{}"
		}
	}
	extra := param.OneModifySt.Extra
	if len(extra) > 0 {
		x := struct{}{}
		err := json.Unmarshal([]byte(extra), &x)
		if err != nil {
			logger.Errorf("%s: fail err: %v", fun, err)
			return err
		}
	} else {
		param.Extra = "{}"
	}

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
		logger.Infof("%s: got db : %v", fun, models.DB_CONF)
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
		if FindNoConflictRecord(db, param.OneModifySt, -1) {
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
	for i := 0; i < len(list); i++ {
		data.List = append(data.List, &proto.GetGroupsItem{
			GroupSt: &list[i],
		})
	}
	return
}

func GetConfig(param *proto.GetConfigParam) (data *proto.GetConfigData, err error) {
	fun := "GetConfig"
	app := param.App
	env := param.Env
	tid := param.Tid
	key := param.Key
	i := 1 / param.Tid
	logger.Debugf("%s: got i: %v", fun, i)

	data = &proto.GetConfigData{}
	now := time.Now().Unix()

	r := proto.RecordSt{}
	dbname := GetDbName(app, env)
	db := models.GetConn(dbname)
	//err = db.Table("record_sts").Where("tid = ? and `key` = ? and ((st < ? and ? < et) or et = 0)", tid, key, now, now).Order("et desc").Limit(1).First(&r).Error
	err = db.Where("tid = ? and `key` = ? and ((st < ? and ? < et) or et = 0)", tid, key, now, now).Order("et desc").Limit(1).First(&r).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
		return
	} else if err != nil {
		logger.Errorf("%s: fail . err: %v", fun, err)
		return
	}
	data.RecordSt = &r

	logger.Debugf("%s: got param: %v, data: %v", fun, param, data)
	return
}

func ModConfig(param *proto.ModConfigParam) (err error) {
	fun := "ModConfig"
	app := param.App
	env := param.Env
	tid := param.OneModifySt.Tid
	key := param.OneModifySt.Key
	et := param.OneModifySt.Et
	val := param.OneModifySt.Val
	set_st := param.SetSt
	set_et := param.SetEt
	del := param.Del
	extra := param.Extra

	dt := param.DataType
	switch dt {
	case "json":
		fallthrough
	default:
		if len(val) > 0 {
			x := struct{}{}
			err := json.Unmarshal([]byte(val), &x)
			if err != nil {
				logger.Errorf("%s: got err: %v", fun, err)
				return err
			}
		} else {
			param.Val = "{}"
		}
	}

	if len(extra) > 0 {
		x := struct{}{}
		err := json.Unmarshal([]byte(extra), &x)
		if err != nil {
			logger.Errorf("%s: got err: %v", fun, err)
			return err
		}
	} else {
		param.Extra = "{}"
	}
	//byt, _ := json.Marshal(param)

	dbname := GetDbName(app, env)
	if len(dbname) == 0 {
		return errors.New("no db specified.")
	}

	db := models.GetConn(dbname)
	logger.Debugf("%s: got db: %v", fun, models.DB_CONF)
	// 开始事务
	db = db.Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("%s: got recover: %v", fun, r)
			db.Rollback()
		} else if err != nil {
			logger.Errorf("%s: got recover: %v", fun, err)
			db.Rollback()
		}
	}()
	if db == nil {
		logger.Infof("%s: got db nil: %v", fun, param)
		return errors.New(" no conn get:" + dbname)
	}

	if del {
		cnt := db.Where("tid=? and `key`=? and et=?", tid, key, et).Delete(&proto.RecordSt{}).RowsAffected
		logger.Infof("%s: deleted: %v, cnt: %v", fun, param, cnt)
		//if err != nil {
		//	fmt.Println("fail del ", param, err)
		//	return
		//}
	} else {
		// modify
		map_updates := make(map[string]interface{})
		map_updates["val"] = val
		map_updates["extra"] = extra
		map_updates["ut"] = time.Now().Unix()
		// 针对修改生效起时时间的要检查是否与其他规则冲突
		orig_record := proto.RecordSt{}
		err = db.Where("tid=? and `key`=? and et=? and del=false", tid, key, et).Limit(1).Find(&orig_record).Error
		if err != nil {
			logger.Errorf("%s: param :%v, got err: %v", fun, param, err)
			return
		}

		if set_st >= 0 {
			map_updates["st"] = set_st
			orig_record.St = set_st
		}
		if set_et >= 0 {
			map_updates["et"] = set_et
			orig_record.Et = set_et
		}

		if !FindNoConflictRecord(db, orig_record.OneModifySt, et) {
			// conflict
			err = errors.New("conflicted, modfication not take eff.")
			logger.Errorf("%s: param :%v, got err: %v", fun, param, err)
			return
		}

		logger.Infof("%s: got param: %v, mpa: %v", fun, param, map_updates)
		err = db.Model(&proto.RecordSt{}).Where("tid=? and `key`=? and et=?", tid, key, et).Updates(map_updates).Error
		if err != nil {
			logger.Errorf("%s: param :%v, got err: %v", fun, param, err)
			return errors.New(" fail update config.")
		}
	}

	db.Commit()
	return
}
