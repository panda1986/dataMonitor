#!/bin/bash

# calc the dir
echo "argv[0]=$0"
if [[ ! -f $0 ]]; then
    echo "directly execute the scripts on shell.";
    work_dir=`pwd`
else
    echo "execute scripts in file: $0";
    work_dir=`dirname $0`; work_dir=`(cd ${work_dir} && pwd)`
fi

objs=$work_dir/objs
release=$objs/_release
mkdir -p $objs
echo "work_dir: $work_dir"
echo "objs: $objs"
echo "release: $release"

function go_platform()
{
    # for go api
    go_blog="http://blog.csdn.net/win_lin/article/details/40618671"
    # check go
    go help >/dev/null 2>&1
    ret=$?; if [[ 0 -ne $ret ]]; then echo "go not install, see $go_blog. ret=$ret"; exit $ret; fi
    echo "go is ok"
    # check GOPATH
    if [[ -d $GOPATH ]]; then
        echo "GOPATH=$GOPATH";
    else
        echo "GOPATH not set.";
        echo "see $go_blog.";
        exit -1;
    fi
    echo "GOPATH is ok"
}

function install_pkg()
{

    # lib from go-oryx.
    if [[ ! -d $GOPATH/src/github.com/ossrs/go-oryx-lib ]]; then
        echo "install go-oryx-lib"
        mkdir -p $GOPATH/src/github.com/ossrs && cd $GOPATH/src/github.com/ossrs &&
        git clone https://github.com/ossrs/go-oryx-lib.git
        ret=$?; if [[ $ret -ne 0 ]]; then echo "build go-oryx-lib failed. ret=$ret"; exit $ret; fi
    fi
    echo "go-oryx-lib ok"

    if [[ ! -d $GOPATH/src/github.com/go-sql-driver/mysql ]]; then
        echo "install mysql"
        mkdir -p $GOPATH/src/github.com/go-sql-driver && cd $GOPATH/src/github.com/go-sql-driver &&
        rm -rf mysql-1.2 && tar xf $work_dir/../../../3rdparty/go/mysql-1.2.tar.gz &&
        rm -f mysql && ln -sf mysql-1.2 mysql
        ret=$?; if [[ 0 -ne $ret ]]; then echo "install github.com/go-sql-driver/mysql failed. ret=$ret"; exit $ret; fi
    fi
    echo "mysql ok"
}

function install_data_monitor()
{
    if [[ ! -d $GOPATH/src/github.com/panda1986/dataMonitor ]]; then
        mkdir -p $GOPATH/src/github.com/panda1986 &&
        ln -sf $work_dir $GOPATH/src/github.com/panda1986/dataMonitor
        ret=$?; if [[ 0 -ne $ret ]]; then echo "github.com/panda1986/dataMonitor failed. ret=$ret"; exit $ret; fi
    fi
    echo "dataMonitor ok"

    go build -o objs/kaloon.monitor github.com/panda1986/dataMonitor/monitor
    ret=$?; if [[ 0 -ne $ret ]]; then echo "build monitor failed. ret=$ret"; exit $ret; fi
    echo "build monitor ok"
}

# prepare the platform.
go_platform
install_pkg
install_data_monitor

echo ""
echo "* 启动数据统计功能:"
echo "      依赖 hdfs，SQL"
echo "      go build -o objs/kaloon.monitor github.com/panda1986/dataMonitor/monitor && ./objs/kaloon.monitor -conf conf/monitor.conf"
