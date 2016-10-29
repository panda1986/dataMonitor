package main

import (
    "chnvideo.com/cloud/common/core"
    "chnvideo.com/cloud/common/mysql"
    "database/sql"
    "errors"
    "fmt"
    "strings"
)

const (
    NormalUser string = "lv0"
)

var ErrorNoRows = errors.New("sql: no rows in result set")

type SqlServer struct {
    sql *mysql.SqlClient
}

func NewSqlServer(c mysql.SqlConfig) *SqlServer {
    s := &SqlServer{}
    s.sql = mysql.NewSqlClient(c)
    return s
}

func (s *SqlServer) Open() error {
    return s.sql.Open()
}

func (s *SqlServer) Close() {
    s.sql.Close()
}

func (s *SqlServer) Exec(query string, args ...interface{}) (int64, int64, error) {
    return s.sql.Exec(query, args...)
}

func (s *SqlServer) QueryRow(query string, args ...interface{}) *sql.Row {
    return s.sql.QueryRow(query, args...)
}

func (s *SqlServer) Query(query string, args ...interface{}) (*sql.Rows, error) {
    return s.sql.Query(query, args...)
}

func (s *SqlServer) Scan(r *sql.Row, dest ...interface{}) (err error) {
    return s.sql.Scan(r, dest...)
}

// for users crud
func (s *SqlServer) GetNormalusers() ([]*User, error) {
    users := []*User{}

    query := fmt.Sprintf("select id, name from user_data where role = '%v' ", NormalUser)
    rows, err := s.sql.Query(query)
    if err != nil {
        if err == mysql.ErrorNoRows {
            core.LoggerWarn.Println("get 0 user from user_data, query is", query)
            return users, nil
        }
        core.LoggerError.Println("get users failed, query is", query, "err is", err)
        return users, err
    }

    defer rows.Close()
    for rows.Next() {
        var id int
        var name string
        if err = rows.Scan(&id, &name); err != nil {
            core.LoggerError.Println("row scan user item failed, err is", err)
            return users, err
        }
        users = append(users, &User{Id: id, Name: name})
    }

    err = rows.Err()
    if err != nil {
        core.LoggerError.Println("not encounter the eof of the rows, err is", err)
    }

    return users, err
}

func (s *SqlServer) CheckUser(name, passwd string) (id int, err error) {
    query := fmt.Sprintf("select id from user_data where name='%v' and passwd='%v'", name, passwd)
    row := s.sql.QueryRow(query)

    if err = row.Scan(&id); err != nil {
        core.LoggerError.Println("row scan user id failed, err is", err)
        err = fmt.Errorf("can't find name:%v in db", name)
        return
    }

    return
}

func (s *SqlServer) CheckUserExist(id int) (exist bool) {
    query := fmt.Sprintf("select count(id) from user_data where id = %v ", id)
    row := s.sql.QueryRow(query)

    var count int
    if err := row.Scan(&count); err != nil {
        core.LoggerError.Println("row scan user count failed, err is", err)
        return
    }
    if count > 0 {
        return true
    }
    return
}

func (s *SqlServer) CreateUser(name, passwd string) error {
    query := "insert into user_data(name, passwd) values(?, ?)"
    row_count, uid, err := s.sql.Exec(query, name, passwd)
    if err != nil {
        core.LoggerError.Println("create user_data", name, "name failed, err is", err)
        return err
    }
    core.LoggerTrace.Println("create user_data", name, "success, id=", uid, "row_count=", row_count)

    return nil
}

func (s *SqlServer) UpdateUser(id int, passwd string) (error) {
    if !s.CheckUserExist(id) {
        return fmt.Errorf("user id %v not exist", id)
    }

    query := "update user_data set passwd=? where id=?"
    row_count, _, err := s.sql.Exec(query, passwd, id)
    if err != nil {
        core.LoggerError.Println(fmt.Sprintf("update user_data id=%v passwd failed err is %v", id, err))
        return err
    }
    core.LoggerTrace.Println("update user_data passwd success row_count=", row_count, "id=", id)

    return nil
}

func (s *SqlServer) DeleteUser(id int) (error) {
    query := "delete from user_data where id = ?"
    row_count, uid, err := s.sql.Exec(query, id)
    if err != nil {
        core.LoggerError.Println("delete user failed, err is", err)
        return err
    }
    core.LoggerTrace.Println("delete user success, id=", uid, "row_count=", row_count)

    return nil
}

//for devices crud
func (s *SqlServer) GetDevices() ([]*Device, error) {
    devices := []*Device{}

    query := fmt.Sprintf("select * from devices")
    rows, err := s.sql.Query(query)
    if err != nil {
        if err == mysql.ErrorNoRows {
            core.LoggerWarn.Println("get 0 device from devices, query is", query)
            return devices, nil
        }
        core.LoggerError.Println("get devices failed, query is", query, "err is", err)
        return devices, err
    }

    defer rows.Close()
    for rows.Next() {
        device := &Device{}
        if err = rows.Scan(&device.Id, &device.Name, &device.Uid, &device.Spec, &device.Precision, &device.MeasureUnit, &device.MeasureMin, &device.MeasureMax); err != nil {
            core.LoggerError.Println("row scan user item failed, err is", err)
            return devices, err
        }
        devices = append(devices, device)
    }

    err = rows.Err()
    if err != nil {
        core.LoggerError.Println("not encounter the eof of the rows, err is", err)
    }

    return devices, err
}

func (s *SqlServer) CreateDevice(device *Device) error {
    query := "insert into devices(`name`, `uid`, `specification`, `precision`, `unit`, `min`, `max`) values(?, ?, ?, ?, ?, ?, ?)"
    row_count, uid, err := s.sql.Exec(query, device.Name, device.Uid, device.Spec, device.Precision, device.MeasureUnit, device.MeasureMin, device.MeasureMax)
    if err != nil {
        core.LoggerError.Println("create device", device, "failed, err is", err)
        return err
    }
    core.LoggerTrace.Println("create device", device, "success, id=", uid, "row_count=", row_count)

    return nil
}

func (s *SqlServer) CheckDeviceExist(id int) (exist bool) {
    query := fmt.Sprintf("select count(id) from devices where id = %v ", id)
    row := s.sql.QueryRow(query)

    var count int
    if err := row.Scan(&count); err != nil {
        core.LoggerError.Println("row scan device count failed, err is", err)
        return
    }
    if count > 0 {
        return true
    }
    return
}

func (s *SqlServer) CheckDeviceUidExist(uid string) (exist bool) {
    query := fmt.Sprintf("select count(uid) from devices where uid = '%v' ", uid)
    row := s.sql.QueryRow(query)

    var count int
    if err := row.Scan(&count); err != nil {
        core.LoggerError.Println("row scan uid device count failed, err is", err)
        return
    }
    if count > 0 {
        return true
    }
    return
}

func (s *SqlServer) QueryDevice(uid string) (device *Device, err error) {
    query := fmt.Sprintf("select * from devices where `uid`='%v'", uid)
    row := s.sql.QueryRow(query)

    device = &Device{}
    if err = row.Scan(&device.Id, &device.Name, &device.Uid, &device.Spec, &device.Precision, &device.MeasureUnit, &device.MeasureMin, &device.MeasureMax); err != nil {
        core.LoggerError.Println(fmt.Sprintf("row fetch uid %v device failed, err is %v", uid, err))
        return
    }

    return
}

func (s *SqlServer) UpdateDevice(device *Device) (error) {
    if !s.CheckDeviceExist(device.Id) {
        return fmt.Errorf("device id %v not exist", device.Id)
    }

    query := "update devices set `name`=?, `specification`=?, `precision`=?, `unit`=?, `min`=?, `max`=? where id=?"
    row_count, _, err := s.sql.Exec(query, device.Name, device.Spec, device.Precision, device.MeasureUnit, device.MeasureMin, device.MeasureMax, device.Id)
    if err != nil {
        core.LoggerError.Println(fmt.Sprintf("update device id=%v failed err is %v", device.Id, err))
        return err
    }
    core.LoggerTrace.Println("update device success row_count=", row_count, "id=", device.Id)

    return nil
}

func (s *SqlServer) DeleteDevice(id int) (error) {
    query := "delete from devices where id = ?"
    row_count, uid, err := s.sql.Exec(query, id)
    if err != nil {
        core.LoggerError.Println("delete device failed, err is", err)
        return err
    }
    core.LoggerTrace.Println("delete device success, id=", uid, "row_count=", row_count)

    return nil
}

func (s *SqlServer) GetFirstUid() (uid string, err error) {
    query := fmt.Sprintf("select uid from records order by time desc limit 1")
    row := s.sql.QueryRow(query)

    if err = row.Scan(&uid); err != nil {
        core.LoggerError.Println("row scan first record uid failed, err is", err)
        return
    }
    return
}

func (s *SqlServer) QueryRecords(uid string, from, to int64) ([]*Record, error)  {
    records := []*Record{}

    query := fmt.Sprintf("select * from records where `uid`='%v' and time between %v and %v order by time asc", uid, from, to)
    rows, err := s.sql.Query(query)
    if err != nil {
        if err == mysql.ErrorNoRows {
            core.LoggerWarn.Println("get 0 record from records, query is", query)
            return records, nil
        }
        core.LoggerError.Println("get records failed, query is", query, "err is", err)
        return records, err
    }

    defer rows.Close()
    for rows.Next() {
        record := &Record{}
        var watchers string
        if err = rows.Scan(&record.Id, &record.Uid, &record.Value, &watchers, &record.UserId, &record.Time, &record.Desc,); err != nil {
            core.LoggerError.Println("row scan user item failed, err is", err)
            return records, err
        }
        record.Watchers = strings.Split(watchers, ",")
        records = append(records, record)
    }

    err = rows.Err()
    if err != nil {
        core.LoggerError.Println("not encounter the eof of the rows, err is", err)
    }

    return records, err
}

// for records crud
func (s *SqlServer) GetRecords() ([]*Record, error) {
    records := []*Record{}

    query := fmt.Sprintf("select * from records")
    rows, err := s.sql.Query(query)
    if err != nil {
        if err == mysql.ErrorNoRows {
            core.LoggerWarn.Println("get 0 record from records, query is", query)
            return records, nil
        }
        core.LoggerError.Println("get records failed, query is", query, "err is", err)
        return records, err
    }

    defer rows.Close()
    for rows.Next() {
        record := &Record{}
        var watchers string
        if err = rows.Scan(&record.Id, &record.Uid, &record.Value, &watchers, &record.UserId, &record.Time, &record.Desc,); err != nil {
            core.LoggerError.Println("row scan user item failed, err is", err)
            return records, err
        }
        record.Watchers = strings.Split(watchers, ",")
        records = append(records, record)
    }

    err = rows.Err()
    if err != nil {
        core.LoggerError.Println("not encounter the eof of the rows, err is", err)
    }

    return records, err
}

func (s *SqlServer) CreateRecord(record *Record) error {
    if !s.CheckDeviceUidExist(record.Uid) {
        return fmt.Errorf("uid %v not exist", record.Uid)
    }

    watchers := strings.Join(record.Watchers, ",")
    query := "insert into records(`uid`, `value`, `watchers`, `user_id`, `time`, `desc`) values(?, ?, ?, ?, ?, ?)"
    row_count, uid, err := s.sql.Exec(query, record.Uid, record.Value, watchers, record.UserId, record.Time, record.Desc)
    if err != nil {
        core.LoggerError.Println("create record", record, "failed, err is", err)
        return err
    }
    core.LoggerTrace.Println("create record", record, "success, id=", uid, "row_count=", row_count)

    return nil
}

func (s *SqlServer) CheckRecordExist(id int) (exist bool) {
    query := fmt.Sprintf("select count(id) from records where id = %v ", id)
    row := s.sql.QueryRow(query)

    var count int
    if err := row.Scan(&count); err != nil {
        core.LoggerError.Println("row scan record count failed, err is", err)
        return
    }
    if count > 0 {
        return true
    }
    return
}

func (s *SqlServer) UpdateRecord(record *Record) (error) {
    if !s.CheckRecordExist(record.Id) {
        return fmt.Errorf("record id %v not exist", record.Id)
    }

    watchers := strings.Join(record.Watchers, ",")
    query := "update records set `uid`=?, `value`=?, `watchers`=?, `desc`=? where id=?"
    row_count, _, err := s.sql.Exec(query, record.Uid, record.Value, watchers, record.Desc, record.Id)
    if err != nil {
        core.LoggerError.Println(fmt.Sprintf("update record id=%v failed err is %v", record.Id, err))
        return err
    }
    core.LoggerTrace.Println("update record success row_count=", row_count, "id=", record.Id)

    return nil
}

func (s *SqlServer) DeleteRecord(id int) (error) {
    query := "delete from records where id = ?"
    row_count, uid, err := s.sql.Exec(query, id)
    if err != nil {
        core.LoggerError.Println("delete record failed, err is", err)
        return err
    }
    core.LoggerTrace.Println("delete record success, id=", uid, "row_count=", row_count)

    return nil
}