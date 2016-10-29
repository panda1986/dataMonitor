#!/bin/bash
# user can config the following configs, then package.
#######################################
# user options: ./install.sh -4001
#######################################
PORT=4001
echo "arg0: port"
echo "for example, $0"
echo "for example, $0 4001"
if [[ $1 != "" ]]; then PORT=$1; fi

INSTALL="/opt/bravo/data_monitor"

if [[ $PORT == "" ]]; then
    SERVICE=kaloon-data-monitor
    TARGET_INSTALL=$INSTALL
else
    SERVICE=kaloon-data-monitor-${PORT}
    TARGET_INSTALL=${INSTALL}.${PORT}
fi
INITD=/etc/init.d/${SERVICE}
#######################################
# global config
#######################################
ok_msg(){
    echo -e "${1}${POS}${BLACK}[${GREEN}  OK  ${BLACK}]"
}

failed_msg(){
    echo -e "${1}${POS}${BLACK}[${RED}FAILED${BLACK}]"
}
##################################################################################
# discover the current work dir.
##################################################################################
echo "argv[0]=$0"
if [[ ! -f $0 ]]; then
    echo "directly execute the scripts on shell.";
    work_dir=`pwd`
else
    echo "execute scripts in file: $0";
    work_dir=`dirname $0`; work_dir=`(cd ${work_dir} && pwd)`
fi

# require sudo users
sudo echo "ok";
ret=$?; if [[ 0 -ne ${ret} ]]; then echo "[error]: you must be sudoer"; exit 1; fi

function check_process() {
    monitor_id=`ps aux| grep kaloon| grep monitor| gerp conf|grep ${PORT}|awk 'NR==1 {print $2}'`
    echo $monitor_id
    if [[ -z $monitor_id ]]; then return 1; fi

    ps -p ${monitor_id}>/dev/null;

    return $?;
}

function install(){
    # user must stop service first.
    if [[ -f ${INITD} ]]; then
        ${INITD} status >/dev/null 2>&1
        ret=$?; if [[ 0 -eq ${ret} ]]; then
            failed_msg "you must stop the service first: sudo ${INITD} stop";
            exit 1;
        fi
    fi

    install_root=$TARGET_INSTALL
    install_bin=$install_root/bms.live.data
    conf_file="$install_root/liveData.conf"
    sys_log_file="$install_root/sys.log"
    # backup installed apiserver.
    if [[ -d $install_root ]]; then
        version="unknown"
        if [[ -f $install_bin ]]; then
            version=`$install_bin -v 2>&1`
        fi

        backup_dir=${install_root}.`date "+%Y-%m-%d_%H-%M-%S"`.v-$version
        echo "backup installed dir, version=$version"
        echo "    to=$backup_dir"
        sudo mv $install_root $backup_dir
        ret=$?; if [[ 0 -ne ${ret} ]]; then failed_msg "backup installed dir failed"; return $ret; fi
        ok_msg "backup installed dir success"
    fi

    echo "create install dir"
    sudo mkdir -p $install_root
    ret=$?; if [[ 0 -ne ${ret} ]]; then failed_msg "create install dir failed"; return $ret; fi
    ok_msg "create install dir success"

    ok_msg "prepare components"
    (
        ok_msg "prepare init.d script" &&
        sed -i "s|PORT=.*$|PORT=${PORT}|g" etc/init.d/kl_monitor &&
        sed -i "s|ROOT=.*$|ROOT=\"${install_root}\"|g" etc/init.d/kl_monitor &&
        sed -i "s|APP=.*$|APP=\"${install_bin}\"|g" etc/init.d/kl_monitor &&
        sed -i "s|CONFIG=.*$|CONFIG=\"${conf_file}\"|g" etc/init.d/kl_monitor &&
        sed -i "s|SYSLOG=.*$|SYSLOG=\"${sys_log_file}\"|g" etc/init.d/kl_monitor &&
        sed -i "s|\"listen\": .*|\"listen\": ${PORT},|g" monitor.conf
    )

    echo "copy main file"
    (
        sudo mkdir -p $install_root &&
        sudo cp -f kaloon.monitor $install_root &&
        sudo cp -f monitor.conf $install_root &&
        sudo mkdir -p $install_root/static-dir &&
        sudo cp -af static-dir $install_root/static-dir/ &&
        sudo mkdir -p $install_root/etc/init.d &&
        sudo cp -rf etc/init.d/kl_monitor $install_root/etc/init.d
    )
    ret=$?; if [[ 0 -ne ${ret} ]]; then failed_msg "copy main file failed"; return $ret; fi
    ok_msg "copy main file success"

    ok_msg "install init.d scripts"
    (
        sudo rm -f ${INITD} &&
        sudo ln -sf $install_root/etc/init.d/kl_monitor ${INITD}
    )
    ret=$?; if [[ 0 -ne ${ret} ]]; then failed_msg "install init.d scripts failed"; return $ret; fi
    ok_msg "install init.d scripts success"
}

install
ret=$?; if [[ $ret -ne 0 ]]; then
    failed_msg " install go failed."
    exit $ret;
fi

# about
echo "before start, install data_monitor.sql"
echo "install success, you can):" &&
echo "      sudo ${INITD} start"
echo "kaloon_data_monitor root is ${TARGET_INSTALL}"

