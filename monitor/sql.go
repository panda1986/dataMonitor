package main

import (
    "chnvideo.com/cloud/common/core"
    "chnvideo.com/cloud/common/mysql"
    "database/sql"
    "errors"
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

func
