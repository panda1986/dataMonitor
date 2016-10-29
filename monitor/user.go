package main

import (
    "io"
    "io/ioutil"
    "fmt"
    "encoding/json"
    ol "github.com/ossrs/go-oryx-lib/logger"
)

type User struct {
    Id int `json:"id"`
    Name string `json:"name"`
    Passwd string `json:"passwd"`
}

func (v *User) Decode(r io.Reader) (err error) {
    var data []byte
    if data, err = ioutil.ReadAll(r); err != nil {
        ol.E(nil, fmt.Sprintf("read users body failed, err is %v", err))
        return
    }
    if err = json.Unmarshal(data, v); err != nil {
        ol.E(nil, "parse user failed, err is", err)
        return
    }
    if len(v.Name) == 0 || len(v.Passwd) == 0 {
        ol.E(nil, fmt.Sprintf("name=%v, or passwd=%v is empty", v.Name, v.Passwd))
        return fmt.Errorf("name=%v and passwd=%v should not empty", v.Name, v.Passwd)
    }
    return
}

type UserAction struct {
    Action string `json:"action"`
    Id int `json:"id"`
    Passwd string `json:"passwd"`
}

func (v *UserAction) Decode(r io.Reader) (err error) {
    var data []byte
    if data, err = ioutil.ReadAll(r); err != nil {
        ol.E(nil, fmt.Sprintf("read users body failed, err is %v", err))
        return
    }
    if err = json.Unmarshal(data, v); err != nil {
        ol.E(nil, "parse user failed, err is", err)
        return
    }
    return
}

// 数据库用户
type DbUsers struct {
    sql *SqlServer
}

func NewDbUsers(sql *SqlServer) *DbUsers {
    v := &DbUsers{
        sql: sql,
    }
    return v
}

// create user, if ok, return user id.
func (v *DbUsers) create(name, passwd string) (err error) {
    return v.sql.CreateUser(name, passwd)
}

// load users of specified role from db
func (v *DbUsers) load(role string) (users []*User, err error) {
    return v.sql.GetNormalusers()
}

// get user of specified id
func (v *DbUsers) getById(userId int) (user *User, err error) {
    return
}

func (v *DbUsers) check(user *User) (id int, err error) {
    return v.sql.CheckUser(user.Name, user.Passwd)
}

// update user info, only support update name
func (v *DbUsers) updatePasswd(userId int, passwd string) (err error) {
    if len(passwd) == 0 {
        return fmt.Errorf("passwd should not empty")
    }
    return v.sql.UpdateUser(userId, passwd)
}

// delete user info from db
func (v *DbUsers) delete(userId int) (err error) {
    return v.sql.DeleteUser(userId)
}