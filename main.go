package main

import (
    "chnvideo.com/cloud/common/core"
    ol "github.com/ossrs/go-oryx-lib/logger"
    ojson "github.com/ossrs/go-oryx-lib/json"
    oo "github.com/ossrs/go-oryx-lib/options"
    oh "github.com/ossrs/go-oryx-lib/http"
    "os"
    "fmt"
    "net/http"
)

const (
    version string = "1.0.0"
)

var server = fmt.Sprintf("DataMonitor/%v", version)

// 配置文件的解析和检查
type MonitorConfig struct {
    core.Config
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
    return
}

// monitor main struct
type Monitor struct {
    config *MonitorConfig
}

func NewMonitor(c *MonitorConfig) *Monitor {
    v := &Monitor{
        config: c,
    }
    return v
}

func (v *Monitor) Serve() {
    go func() {

    }()
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
        mt.Serve()

        ol.T(nil, "apply listen", c.Listen)
        ol.T(nil, "handle /api/v1/versions")
        ol.T(nil, "handle /api/v1/users")
        ol.T(nil, "handle /api/v1/warnings")
        ol.T(nil, "handle /api/v1/data")

        http.HandleFunc("/api/v1/versions", func(w http.ResponseWriter, r *http.Request) {
            ver := struct {
                Version string `json:"version"`
            }{version}
            oh.WriteData(nil, w, r, ver)
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