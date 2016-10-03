package main

import (
)

type User struct {
    Id int `json:"id"`
    Name string `json:"name"`
    Password string `json:"passwd"`
    Role string `json:"role"`
}

// 数据库用户
type DbUsers struct {

}

// create user, if ok, return user id.
func (v *DbUsers) create() (userId int, err error) {
    return
}

// load users of specified role from db
func (v *DbUsers) load(role string) ([]*User) {
    return
}

// get user of specified id
func (v *DbUsers) get(userId int) (*User, error) {
    return
}

// update user info, only support update name
func (v *DbUsers) updateName(userId int, name string) (error) {
    return
}

func (v *DbUsers) updatePasswd(userId int, passwd string) (error) {
    return
}

// delete user info from db
func (v *DbUsers) delete(userId int) (error) {
    return
}