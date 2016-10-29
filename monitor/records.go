package main

import (
    "fmt"
    "io"
    "io/ioutil"
    "encoding/json"
    ol "github.com/ossrs/go-oryx-lib/logger"
)

// 设备台账中的某台设备的巡检记录
type Record struct {
    Id int `json:"id"` //device id,generate by db
    Uid string `json:"uid"` //统一编号
    Value float64 `json:"value"` //测量值
    Watchers []string `json:"watchers"` //检查人
    UserId int `json:"user_id"` //用户Id
    Time int64 `json:"time"` //记录时间戳
    Desc string `json:"desc"` //记录描述
}

func (v *Record) String() string {
    return fmt.Sprintf("id:%v, uid:%v, value:%v, watchers:%v", v.Id, v.Uid, v.Value, v.Watchers)
}

func (v *Record) Decode(r io.Reader) (err error) {
    var data []byte
    if data, err = ioutil.ReadAll(r); err != nil {
        ol.E(nil, fmt.Sprintf("read users body failed, err is %v", err))
        return
    }
    if err = json.Unmarshal(data, v); err != nil {
        ol.E(nil, "parse user failed, err is", err)
        return
    }

    if len(v.Uid) == 0 || len(v.Watchers) == 0 {
        return fmt.Errorf("device variable %v should not empty", v)
    }
    return
}

type RecordAction struct {
    Action string `json:"action"`
    Record *Record `json:"record"`
}

func (v *RecordAction) Decode(r io.Reader) (err error) {
    var data []byte
    if data, err = ioutil.ReadAll(r); err != nil {
        ol.E(nil, fmt.Sprintf("read record action body failed, err is %v", err))
        return
    }
    if err = json.Unmarshal(data, v); err != nil {
        ol.E(nil, "parse record action failed, err is", err)
        return
    }
    return
}

// 巡检记录账簿
type DbRecords struct {
    sql *SqlServer
}

func NewDbRecords(sql *SqlServer) *DbRecords {
    v := &DbRecords{
        sql: sql,
    }
    return v
}

// create record.
func (v *DbRecords) create(record *Record) (err error) {
    return v.sql.CreateRecord(record)
}

// load records from db
func (v *DbRecords) load() (records []*Record, err error) {
    return v.sql.GetRecords()
}

// update record info
func (v *DbRecords) update(record *Record) (err error) {
    if record == nil {
        return fmt.Errorf("record not valid, is nil obj")
    }
    return v.sql.UpdateRecord(record)
}

// delete record from db
func (v *DbRecords) delete(recordId int) (err error) {
    return v.sql.DeleteRecord(recordId)
}

func (v *DbRecords) query(uid string, from, to int64) (records []*Record, err error) {
    return v.sql.QueryRecords(uid, from, to)
}

func (v *DbRecords) queryUid() (uid string, err error) {
    return v.sql.GetFirstUid()
}


