package main

import (
    "io"
    "io/ioutil"
    "fmt"
    "encoding/json"
    ol "github.com/ossrs/go-oryx-lib/logger"
)

// 设备台账中的某台设备
type Device struct {
    Id int `json:"id"` //device id,generate by db
    Name string `json:"name"` // 计量器具名称
    Uid string `json:"uid"` //统一编号
    Spec string `json:"specification"` //规格型号
    Precision string  `json:"precision"` //精度
    MeasureUnit string `json:"unit"` //测量单位
    MeasureMin float64 `json:"min"` //测量最小值
    MeasureMax float64 `json:"max"` //测量最大值
}

func (v *Device) String() string {
    return fmt.Sprintf("name:%v, uid:%v, spec:%v, prec:%v, measure:%v-%v %v", v.Name, v.Uid, v.Spec, v.Precision, v.MeasureMin, v.MeasureMax, v.MeasureUnit)
}

func (v *Device) Decode(r io.Reader) (err error) {
    var data []byte
    if data, err = ioutil.ReadAll(r); err != nil {
        ol.E(nil, fmt.Sprintf("read device body failed, err is %v", err))
        return
    }
    if err = json.Unmarshal(data, v); err != nil {
        ol.E(nil, "parse device failed, err is", err)
        return
    }

    if len(v.Name) == 0 || len(v.Uid) == 0 || len(v.Spec) == 0 || len(v.Precision) == 0 {
        return fmt.Errorf("device variable %v should not empty", v)
    }
    return
}

type DeviceAction struct {
    Action string `json:"action"`
    Device *Device `json:"device"`
}

func (v *DeviceAction) Decode(r io.Reader) (err error) {
    var data []byte
    if data, err = ioutil.ReadAll(r); err != nil {
        ol.E(nil, fmt.Sprintf("read device body failed, err is %v", err))
        return
    }
    if err = json.Unmarshal(data, v); err != nil {
        ol.E(nil, "parse device failed, err is", err)
        return
    }
    return
}

// 数据库台账
type DbDevices struct {
    sql *SqlServer
}

func NewDbDevices(sql *SqlServer) *DbDevices {
    v := &DbDevices{
        sql: sql,
    }
    return v
}

// create user.
func (v *DbDevices) create(device *Device) (err error) {
    return v.sql.CreateDevice(device)
}

// load devices from db
func (v *DbDevices) load() (devices []*Device, err error) {
    return v.sql.GetDevices()
}

// update device info
func (v *DbDevices) update(device *Device) (err error) {
    if device == nil {
        return fmt.Errorf("device not valid, is nil obj")
    }
    return v.sql.UpdateDevice(device)
}

// delete device info from db
func (v *DbDevices) delete(deviceId int) (err error) {
    return v.sql.DeleteDevice(deviceId)
}

func (v *DbDevices) queryDevice(uid string) (device *Device, err error) {
    return v.sql.QueryDevice(uid)
}

