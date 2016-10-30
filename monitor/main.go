package main

import (
    "chnvideo.com/cloud/common/core"
    "chnvideo.com/cloud/common/mysql"
    ol "github.com/ossrs/go-oryx-lib/logger"
    ojson "github.com/ossrs/go-oryx-lib/json"
    oo "github.com/ossrs/go-oryx-lib/options"
    oh "github.com/ossrs/go-oryx-lib/http"
    "os"
    "fmt"
    "net/http"
    "strconv"
    "io/ioutil"
    "encoding/json"
)

const (
    version string = "1.0.1"

    UpdateAction = "update"
    DeleteAction = "delete"

    ErrorCreateUser oh.SystemError = 1000 + iota
    ErrorPutUser
    ErrorGetUser
    ErrorCreateDevice
    ErrorPutDevice
    ErrorGetDevice
    ErrorCreateRecord
    ErrorPutRecord
    ErrorGetRecord
    ErrorAnalysis
    ErrorQuery
    ErrorLoginFailed
)

var server = fmt.Sprintf("DataMonitor/%v", version)

// 配置文件的解析和检查
type MonitorConfig struct {
    core.Config
    mysql.SqlCommonConfig
}

func (v *MonitorConfig) Loads(conf string) (err error) {
    var f *os.File
    if f, err = os.Open(conf); err != nil {
        ol.E(nil, fmt.Sprintf("open %v failed, err is %v", conf, err))
        return
    }
    defer f.Close()

    if err = ojson.Unmarshal(f, v); err != nil {
        return
    }

    return v.Validate()
}

func (v *MonitorConfig) Validate() (err error) {
    if err = v.Config.Validate(); err != nil {
        ol.E(nil, fmt.Sprintf("validate core config failed, err is %v", err))
        return
    }
    if err = v.SqlCommonConfig.Validate(); err != nil {
        ol.E(nil, fmt.Sprintf("validate mysql config failed, err is %v", err))
        return
    }
    return
}

// monitor main struct
type Monitor struct {
    config *MonitorConfig
    sql *SqlServer
    users *DbUsers
    devices *DbDevices
    records *DbRecords
}

func NewMonitor(c *MonitorConfig) *Monitor {
    v := &Monitor{
        config: c,
        sql: NewSqlServer(c),
    }
    v.users = NewDbUsers(v.sql)
    v.devices = NewDbDevices(v.sql)
    v.records = NewDbRecords(v.sql)
    return v
}

func (v *Monitor) Initialize() (err error) {
    if err = v.sql.Open(); err != nil {
        ol.E(nil, "open sql failed, err is", err)
        return
    }
    return
}

func (v *Monitor) Close() {
    v.sql.Close()
}

func run() int {
    conf := oo.ParseArgv("conf/monitor.conf", version, server)

    ol.T(nil, "start to parse config file", conf)
    c := &MonitorConfig{}
    if err := c.Loads(conf); err != nil {
        ol.E(nil, "config error is", err)
        return -1
    }

    // start reload listen goroutine.
    go core.StartReload(func() (err error) {
        if err = c.Loads(conf); err != nil {
            return
        }
        c.Config.Initialize()
        return
    })

    work := func() (err error) {
        oh.Server = server
        ol.T(nil, "this is data monitor, collect, analysis and monitor data.")

        mt := NewMonitor(c)
        if err = mt.Initialize(); err != nil {
            return
        }
        defer mt.Close()

        initSession()

        ol.T(nil, "apply listen", c.Listen)
        ol.T(nil, "handle /")
        ol.T(nil, "handle /api/v1/login")
        ol.T(nil, "handle /api/v1/logout")
        ol.T(nil, "handle /api/v1/versions")
        ol.T(nil, "handle /api/v1/users")
        ol.T(nil, "handle /api/v1/devices")
        ol.T(nil, "handle /api/v1/records")
        ol.T(nil, "handle /api/v1/analysis")

        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            static_file := http.FileServer(http.Dir("static-dir"))
            static_file.ServeHTTP(w, r)
        })

        http.HandleFunc("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
            b, err := ioutil.ReadAll(r.Body)
            if err != nil {
                oh.Error(nil, fmt.Errorf("read cookie failed : %v", err))
                return
            }

            user := &User{}
            if err := json.Unmarshal(b, user); err != nil {
                oh.Error(nil, fmt.Errorf("json Unmarshal failed : %v", err))
                return
            }

            // check user
            if id, err := mt.users.check(user); err != nil {
                oh.WriteCplxError(nil, w, r, ErrorLoginFailed, err.Error())
                return
            } else {
                sess := globalSessions.SessionCreate(id, user.Name, w, r)
                oh.WriteData(nil, w, r, sess.sid)
                return
            }
        })

        http.HandleFunc("/api/v1/logout", func(w http.ResponseWriter, r *http.Request) {
            globalSessions.SessionDestroy(w, r)
            oh.WriteData(nil, w, r, nil)
        })

        http.HandleFunc("/api/v1/versions", func(w http.ResponseWriter, r *http.Request) {
            ver := struct {
                Version string `json:"version"`
            }{version}
            oh.WriteData(nil, w, r, ver)
        })

        http.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
            if r.Method == "POST" {// create user
                user := &User{}
                if err = user.Decode(r.Body); err != nil {
                    ol.E(nil, "user decode failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorCreateUser, fmt.Sprintf("user decode failed, err is %v", err))
                    return
                }

                if err := mt.users.create(user.Name, user.Passwd); err != nil {
                    ol.E(nil, "create user failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorCreateUser, fmt.Sprintf("create user failed, err is %v", err))
                    return
                }
                oh.WriteData(nil, w, r, nil)
                return
            }
            if r.Method == "PUT" {// delete or update user
                ua := &UserAction{}
                if err = ua.Decode(r.Body); err != nil {
                    ol.E(nil, "user action decode failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorPutUser, fmt.Sprintf("user action decode failed, err is %v", err))
                    return
                }
                if ua.Action == UpdateAction {
                    if err := mt.users.updatePasswd(ua.User.Id, ua.User.Passwd); err != nil {
                        ol.E(nil, "update user passwd failed, err is", err)
                        oh.WriteCplxError(nil, w, r, ErrorPutUser, fmt.Sprintf("update user passwd failed, err is %v", err))
                        return
                    }
                    oh.WriteData(nil, w, r, nil)
                    return
                }
                if ua.Action == DeleteAction {
                    if err := mt.users.delete(ua.User.Id); err != nil {
                        ol.E(nil, "delete user failed, err is", err)
                        oh.WriteCplxError(nil, w, r, ErrorPutUser, fmt.Sprintf("delete user failed, err is %v", err))
                        return
                    }
                    oh.WriteData(nil, w, r, nil)
                    return
                }
                oh.WriteCplxError(nil, w, r, ErrorPutUser, fmt.Sprintf("invalid user action %v", ua.Action))
                return
            }
            if r.Method == "GET" {// load users
                if users, err := mt.users.load(NormalUser); err != nil {
                    ol.E(nil, "load normal users failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorGetUser, fmt.Sprintf("get users failed, err is %v", err))
                    return
                } else {
                    oh.WriteData(nil, w, r, users)
                    return
                }
            }
        })

        http.HandleFunc("/api/v1/devices", func(w http.ResponseWriter, r *http.Request) {
            if r.Method == "POST" {// create device
                device := &Device{}
                if err = device.Decode(r.Body); err != nil {
                    ol.E(nil, "device decode failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorCreateDevice, fmt.Sprintf("device decode failed, err is %v", err))
                    return
                }

                if err := mt.devices.create(device); err != nil {
                    ol.E(nil, "create device failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorCreateDevice, fmt.Sprintf("create device failed, err is %v", err))
                    return
                }
                oh.WriteData(nil, w, r, nil)
                return
            }
            if r.Method == "PUT" {// delete or update device
                da := &DeviceAction{}
                if err = da.Decode(r.Body); err != nil {
                    ol.E(nil, "device action decode failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorPutDevice, fmt.Sprintf("device action decode failed, err is %v", err))
                    return
                }
                if da.Action == UpdateAction {
                    if err := mt.devices.update(da.Device); err != nil {
                        ol.E(nil, "update user passwd failed, err is", err)
                        oh.WriteCplxError(nil, w, r, ErrorPutDevice, fmt.Sprintf("update device failed, err is %v", err))
                        return
                    }
                    oh.WriteData(nil, w, r, nil)
                    return
                }
                if da.Action == DeleteAction {
                    if err := mt.devices.delete(da.Device.Id); err != nil {
                        ol.E(nil, "delete device failed, err is", err)
                        oh.WriteCplxError(nil, w, r, ErrorPutDevice, fmt.Sprintf("delete device failed, err is %v", err))
                        return
                    }
                    oh.WriteData(nil, w, r, nil)
                    return
                }
                oh.WriteCplxError(nil, w, r, ErrorPutDevice, fmt.Sprintf("invalid device action %v", da.Action))
                return
            }
            if r.Method == "GET" {// load devices
                if devices, err := mt.devices.load(); err != nil {
                    ol.E(nil, "load devices failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorGetDevice, fmt.Sprintf("get devices failed, err is %v", err))
                    return
                } else {
                    oh.WriteData(nil, w, r, devices)
                    return
                }
            }
        })

        http.HandleFunc("/api/v1/records", func(w http.ResponseWriter, r *http.Request) {
            if r.Method == "POST" {// create device
                record := &Record{}
                if err = record.Decode(r.Body); err != nil {
                    ol.E(nil, "record decode failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorCreateRecord, fmt.Sprintf("record decode failed, err is %v", err))
                    return
                }

                if err := mt.records.create(record); err != nil {
                    ol.E(nil, "create record failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorCreateRecord, fmt.Sprintf("create record failed, err is %v", err))
                    return
                }
                oh.WriteData(nil, w, r, nil)
                return
            }
            if r.Method == "PUT" {// delete or update device
                ra := &RecordAction{}
                if err = ra.Decode(r.Body); err != nil {
                    ol.E(nil, "record action decode failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorPutRecord, fmt.Sprintf("record action decode failed, err is %v", err))
                    return
                }
                if ra.Action == UpdateAction {
                    if err := mt.records.update(ra.Record); err != nil {
                        ol.E(nil, "update record failed, err is", err)
                        oh.WriteCplxError(nil, w, r, ErrorPutRecord, fmt.Sprintf("update record failed, err is %v", err))
                        return
                    }
                    oh.WriteData(nil, w, r, nil)
                    return
                }
                if ra.Action == DeleteAction {
                    if err := mt.records.delete(ra.Record.Id); err != nil {
                        ol.E(nil, "delete record failed, err is", err)
                        oh.WriteCplxError(nil, w, r, ErrorPutRecord, fmt.Sprintf("delete record failed, err is %v", err))
                        return
                    }
                    oh.WriteData(nil, w, r, nil)
                    return
                }
                oh.WriteCplxError(nil, w, r, ErrorPutRecord, fmt.Sprintf("invalid record action %v", ra.Action))
                return
            }
            if r.Method == "GET" {// load records
                if records, err := mt.records.load(); err != nil {
                    ol.E(nil, "load records failed, err is", err)
                    oh.WriteCplxError(nil, w, r, ErrorGetRecord, fmt.Sprintf("get records failed, err is %v", err))
                    return
                } else {
                    oh.WriteData(nil, w, r, records)
                    return
                }
            }
        })

        http.HandleFunc("/api/v1/analysis", func(w http.ResponseWriter, r *http.Request) {
            // query records by uid and start, end time
            q := r.URL.Query()
            uid := q.Get("uid")

            from, err := strconv.ParseInt(q.Get("from"), 10, 64)
            if err != nil {
                oh.WriteCplxError(nil, w, r, ErrorAnalysis, fmt.Sprintf("get from param failed, err is %v", err))
                return
            }
            to, err := strconv.ParseInt(q.Get("to"), 10, 64)
            if err != nil {
                oh.WriteCplxError(nil, w, r, ErrorAnalysis, fmt.Sprintf("get to param failed, err is %v", err))
                return
            }
            if from > to {
                oh.WriteCplxError(nil, w, r, ErrorAnalysis, fmt.Sprintf("from %v should < to %v", from, to))
            }
            if from == to {
                to += 24 * 60 * 60
            }

            if data, err := mt.records.query(uid, from, to); err != nil {
                oh.WriteCplxError(nil, w, r, ErrorAnalysis, err.Error())
            } else {
                oh.WriteData(nil, w, r, data)
            }
        })

        http.HandleFunc("/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
            uid, err := mt.records.queryUid()
            if err != nil {
                oh.WriteCplxError(nil, w, r, ErrorQuery, fmt.Sprintf("get uid failed, err is %v", err.Error()))
                return
            }

            device, err := mt.devices.queryDevice(uid)
            if err != nil {
                oh.WriteCplxError(nil, w, r, ErrorQuery, fmt.Sprintf("get device failed, err is %v", err.Error()))
                return
            }
            oh.WriteData(nil, w, r, device)
        })

        listen := fmt.Sprintf("0.0.0.0:%v", c.Config.Listen)
        ol.T(nil, "listen at", listen)
        if err = http.ListenAndServe(listen, nil); err != nil {
            ol.E(nil, "http serve at", c.Listen, "failed, err is", err)
            return
        }
        return
    }

    return core.ServerRun(&c.Config, func() int {
        if err := work(); err != nil {
            return -1
        }
        return 0
    })
}

func main()  {
    os.Exit(run())
}