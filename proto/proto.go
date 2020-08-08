package proto

type BaseResp struct {
	Ret     int64  `json:"ret"`
	ErrCode int64  `json:"errcode,omitempty"`
	Msg     string `json:"msg,omitempty"`
}

type OneModifySt struct {
	ID    int64  `gorm:"column:_id;primary_key;not_null;auto_increment" json:"_id"`
	Tid   int64  `gorm:"column:tid;not_null" json:"tid"`
	Key   string `gorm:"column:key;type:varchar(64);not_null" json:"key"`
	Val   string `gorm:"column:val;type:mediumtext;" json:"val"`
	St    int64  `gorm:"not_null;default:0" json:"st"`
	Et    int64  `gorm:"not_null;default:0" json:"et"`
	Ut    int64  `gorm:"not_null;default:0" json:"ut"`
	Del   bool   `gorm:"type:bool;not_null;default:false" json:"del,omitempty"`
	Extra string `gorm:"column:extra;type:mediumtext;" json:"extra"`
}

type RecordSt struct {
	*OneModifySt
}

type GroupSt struct {
	Id   int64  `gorm:"primary_key;not_null;auto_increment" json:"tid"`
	Name string `gorm:"type:varchar(64);not_null;" json:"exp_name"`
	St   int64  `gorm:"not_null;default:0" json:"st"`
	Et   int64  `gorm:"not_null;default:0" json:"et"`
	Ut   int64  `gorm:"not_null;default:0" json:"ut"`
	Del  bool   `gorm:"type:bool;not_null;default:false" json:"del,omitempty"`
}

type HistoryRecordSt struct {
	*OneModifySt
}

type AddConfigParam struct {
	App     string `json:"app"`
	Env     string `json:"env"`
	Tid     int64  `json:"tid"`
	ExpName string `json:"exp_name"`
	*OneModifySt
}

type AddConfigResp BaseResp

type GetGroupsParam struct {
	App string `json:"app"`
	Env string `json:"env"`
}
type GetGroupsResp struct {
	BaseResp
	Data *GetGroupsData `json:"data,omitempty"`
}
type GetGroupsData struct {
	List []*GetGroupsItem `json:"list,omitempty"`
}
type GetGroupsItem struct {
	*GroupSt
}

type GetConfigParam struct {
	App string `json:"app"`
	Env string `json:"env"`
	Tid int64  `json:"tid"`
	Key string `json:"key"`
}
type GetConfigResp struct {
	BaseResp
	Data *GetConfigData `json:"data,omitempty"`
}
type GetConfigData struct {
	*RecordSt
}

type ModConfigParam struct {
	App     string `json:"app"`
	Env     string `json:"env"`
	ExpName string `json:"exp_name"`
	*OneModifySt
	SetSt int64 `json:"set_st,omitempty"`
	SetEt int64 `json:"set_et,omitempty"`
}
type ModConfigResp BaseResp
