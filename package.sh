#!/bin/bash

#######################################
# user options: ./package.sh -j16
#######################################
# the jobs for make.
jobs=$1; if [[ "" == $jobs ]]; then jobs="-j1"; fi

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

build_objs="${work_dir}/objs"
package_dir=${build_objs}/package
kl_monitor="${work_dir}/objs/kaloon.monitor"

##################################################################################
# package the core components
##################################################################################

function package_kl_moitor(){
    release_version=`${kl_monitor} -v 2>&1`

    # svn_version is generated by package().
    package_name="data-monitor-${release_version}"
    package_file="${package_name}.zip"
    package_dirname="${package_dir}/${package_name}"
    monitor_path="${package_dir}/${package_file}"

    echo "start packaging"
    echo "package_file=${package_file}"

    (
        sudo rm -rf ${package_dirname} && mkdir -p ${package_dirname} &&
        cp ${work_dir}/objs/kaloon.monitor ${package_dirname} &&
        mkdir -p ${package_dirname}/static-dir &&
        cp -af ${work_dir}/static-dir ${package_dirname} &&
        cp ${work_dir}/install.sh ${package_dirname}/install.sh &&
        cp ${work_dir}/conf/monitor.conf ${package_dirname}/ &&
        cp ${work_dir}/data_monitor.sql ${package_dirname}/ &&
        mkdir -p ${package_dirname}/etc/init.d &&
        cp ${work_dir}/etc/init.d/kl_monitor ${package_dirname}/etc/init.d &&
        cd ${package_dir} && rm -f ${package_file} && zip -q -r ${package_file} ${package_name}
    )
    ret=$?; if [[ 0 -ne ${ret} ]]; then failed_msg "package kl monitor failed"; return $ret; fi

    ok_msg "success: ${monitor_path}"
}

echo "build core components"
bash build.sh $jobs;
ret=$?; if [[ $ret -ne 0 ]]; then
    failed_msg " build failed."
    exit $ret;
fi

go build -o objs/kaloon.monitor github.com/panda1986/dataMonitor/monitor

package_kl_moitor
ret=$?; if [[ $ret -ne 0 ]]; then
    failed_msg " package live data failed"
    exit $ret;
fi
ok_msg "package data monitor success"

echo "install data monitor: "
echo "      bash install.sh"

