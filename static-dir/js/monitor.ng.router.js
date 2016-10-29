var monitorApp = angular.module('monitorApp', ['ngRoute', "ngResource", "ngCookies", "ngTable", 'ui.bootstrap']);

// config the http interceptor.
monitorApp.config(['$httpProvider', function($httpProvider){
    $httpProvider.interceptors.push('MHttpInterceptor')
}]);

monitorApp.config(["$routeProvider", function ($routeProvider) {
    $routeProvider.otherwise({redirectTo: "/login"})
        .when("/analysis", {templateUrl: "views/analysis.html", controller: "CAnalysis"})
        .when("/login", {templateUrl: "views/login.html", controller: "CLogin"})
        .when("/user", {templateUrl: "views/user.html", controller: "CUser"})
        .when("/devices", {templateUrl: "views/devices.html", controller: "CDevices"})
        .when("/records", {templateUrl: "views/records.html", controller: "CRecords"});
}]);

//default search and default sort
monitorApp.service("userData", function () {
    return {
        user: {
            table_filter: {
                name: ""
            },
            table_sort: {},
            page_count: 10,
            current_page: 1
        },
        device: {
            table_filter: {
                name: "",
                uid: "",
                specification: ""
            },
            table_sort: {},
            page_count: 10,
            current_page: 1
        },
        record: {
            table_filter: {
            },
            table_sort: {},
            page_count: 10,
            current_page: 1
        },
        query: {
            uid: ""
        }
    }
});

function DateController($scope) {
    $scope.date = {};
    $scope.date.date_tab = 1;
    $scope.date.to_date = get_today_date_objs();
    $scope.date.from_date = get_before_days_date_objs(0);
    $scope.is_date_set = function(tabValue){
        return $scope.date.date_tab === tabValue;
    };
    $scope.get_day_data = function (tab_value, date_diff_from_today) {
        $scope.date.date_tab = tab_value;
        $scope.date.from_date = get_before_days_date_objs(date_diff_from_today);
        $scope.date.to_date = get_today_date_objs();
        $scope.get_date_data();
    };
    $scope.get_select_date_data = function(tab_value) {
        $scope.date.date_tab = tab_value;
        $scope.date.from_date.ts = Math.round(($scope.date.from_date.value).getTime() / 1000);
        $scope.date.to_date.ts = Math.round(($scope.date.to_date.value).getTime() / 1000);
        $scope.get_date_data();
    };
    $scope.get_tab_days = function() {
        if ($scope.date.date_tab == 2) {
            return 3;
        }
        if ($scope.date.date_tab == 3) {
            return 7;
        }
        if ($scope.date.date_tab == 4) {
            return 30;
        }
        return 1;
    }
}

// config the route
monitorApp.controller("CMain", ["$scope", "$location", "$cookies", "$interval", "$db_utility", "$db_nav",
    function ($scope, $location, $cookies, $interval, $db_utility, $db_nav) {
        $scope.logs = [];
        // remove expired alert.
        $interval(function () {
            for (var i = 0; i < $scope.logs.length; i++) {
                var log = $scope.logs[i];
                if (log.create + 10000 < new Date().getTime()) {
                    $scope.logs.splice(i, 1);
                    break;
                }
            }
        }, 3000);
        // handler system log event, from $db_utility service.
        $scope.$on("$db_utility_log", function (event, level, msg) {
            var log = {
                level: level, msg: msg, create: new Date().getTime()
            };
            // only show 3 msgs.
            while ($scope.logs.length > 2) {
                $scope.logs.splice(0, 1);
            }
            $scope.logs.push(log);
        });

        // handle system error event, from $db_utility service.
        $scope.$on("$db_utility_http_error", function (event, status, response) {
            if (status != 200) {
                if (!status && !response) {
                    response = "无法访问服务器";
                } else {
                    response = "HTTP/" + status + ", " + response;
                }
            } else {
                var map = {};
                if (map[response.code]) {
                    response = "code=" + response.code + ", " + map[response.code];
                } else {
                    resonse = "code=" + response.code + ", 系统错误";
                }
            }

            $db_utility.log("warn", response);
        });

        // jump to other pages with query.
        $scope.jump = function (to) {
            $db_nav.jump(to, $location);
        };
        $scope.logout = function () {
            var x = $cookies.getAll();
            $cookies.remove("kl_monitor_user");
            $db_nav.go_login($location);
        };
        $scope.user = function () {
            return $cookies.get("kl_monitor_user");
        };
        $scope.is_nav_selected = function (nav) {
            return $db_nav.is_selected(nav);
        };
    }]);

// controller: CHomepage, for the view homepage.html.
monitorApp.controller('CAnalysis', ['$scope', "$location", "$cookies", "$db_nav", "MDBApi", "$db_utility",
    function($scope, $location, $cookies, $db_nav, MDBApi, $db_utility){
    $db_nav.in_analysis();

    // requires user to login first.
    if (!$cookies.get("kl_monitor_user")) {
        $db_nav.go_login($location);
        return;
    }

    $scope.search = {
        uid: ""
    };
    var analysisChart = echarts.init(document.getElementById('echart_analysis'));
    // 指定图表的配置项和数据
    var option = {
        title: {
            text: '设备运行数据分析'
        },
        tooltip: {},
        legend: {
            data:['测量值']
        },
        toolbox: {
            show : true,
            feature : {
                dataZoom : {show: true},
                dataView : {show: true, readOnly: true},
                restore : {show: true},
                saveAsImage : {show: true}
            }
        },
        dataZoom: {
            show : true,
            realtime : true,
            start : 0,
            end : 100
        },
        xAxis: {
            data: []
        },
        yAxis: {},
        series: [{
            name: '测量值',
            type: 'line',
            data: [],
            markPoint: {
                data: [
                    {type: 'max', name: '最大值'},
                    {type: 'min', name: '最小值'}
                ]
            },
            markLine: {
                data: [
                    {type: 'average', name: '平均值'}
                ]
            }
        }]
    };
    analysisChart.setOption(option);
    
    DateController($scope);
    
    $scope.valid = function () {
        return $scope.search.uid != "";
    };
    $scope.get_date_data = function() {
        MDBApi.analysis_query($scope.search.uid, $scope.date.from_date.ts, $scope.date.to_date.ts, function (data) {
            $scope.uid_records = data.data;
            $db_utility.log("trace", "获取"+$scope.search.uid + "[" + absolute_seconds_to_YYYYmmdd_hhmm($scope.date.from_date.ts) + " 到 " + absolute_seconds_to_YYYYmmdd_hhmm($scope.date.to_date.ts) + "]的巡检记录成功");
            var xaxis_data = [];
            var series_data = [];
            for (var i = 0; i < data.data.length; i++) {
                var record = data.data[i];
                xaxis_data.push(absolute_seconds_to_YYYYmmdd_hhmm(record["time"]));
                series_data.push(record["value"]);
            }
            analysisChart.setOption({
                xAxis: {
                    data: xaxis_data
                },
                series: [{
                    name: '测量值',
                    data: series_data
                }]
            });
        });
    };

    var get_first_record_uid = function () {
        MDBApi.query(function (data) {
            $scope.now_device = data.data;
            $scope.search.uid = $scope.now_device.uid;
            analysisChart.setOption({
                title: {
                    text: $scope.now_device.name
                },
                yAxis: {
                    name: $scope.now_device.unit
                }
            });
            $scope.get_date_data();
        })
    };

    if (!$scope.valid()) {
        get_first_record_uid();
    } else {
        $scope.get_date_data();
    }
}]);

//controller:CUser,for the view user.html
monitorApp.controller('CUser',['$scope', "$location", "$cookies", "$db_nav", "userData", "MDBApi", "NgTableParams", "$uibModal", "$db_utility",
    function($scope, $location, $cookies, $db_nav, userData, MDBApi, NgTableParams, $uibModal, $db_utility){
    $db_nav.in_user();

    // requires user to login first.
    if (!$cookies.get("kl_monitor_user")) {
        $db_nav.go_login($location);
        return;
    }

    var load = function () {
        MDBApi.users_load(function (data) {
            $scope.users = data.data;
            var max_page_cnt = Math.ceil($scope.users.length / userData.user.page_count);
            userData.user.current_page = Math.min(userData.user.current_page, max_page_cnt);

            $scope.tableParams = new NgTableParams({
                page: userData.user.current_page,
                count: userData.user.page_count,
                filter: {
                    'name': userData.user.table_filter.name
                },
                sorting: userData.user.table_sort

            }, {
                counts: [10, 30, 50],
                dataset: $scope.users
            });

            //watch search choose
            $scope.$watch(function () {
                return $scope.tableParams.filter();
            }, function (newValue, oldValue) {
                userData.user.table_filter = newValue;
            }, true);

            //watch page count choose
            $scope.$watch(function () {
                return $scope.tableParams.count();
            }, function (newValue, oldValue) {
                userData.user.page_count = newValue;
            }, true);

            //watch when page choose
            $scope.$watch(function () {
                return $scope.tableParams.page();
            }, function (newValue, oldValue) {
                userData.user.current_page = newValue;
            }, true);
        })
    };

    $scope.newUser = function () {
        var modalInstance = $uibModal.open({
            animation: true,
            templateUrl: 'CCreateUser.html',
            controller: 'CNewUser',
            size: "sm",
            resolve: {}
        });

        modalInstance.result.then(function (user) {
            MDBApi.user_create(user, function (res) {
                $db_utility.log("trace", "创建用户【" + user.name + "】成功");
                load();
            })
        }, function () {
            console.log('Modal dismissed at: ' + new Date());
        });
    };

    $scope.modifyUser = function (user) {
        var modalInstance = $uibModal.open({
            templateUrl: 'CModifyUser.html',
            controller: 'CModifyUser',
            size: 'sm',
            resolve: {
                oldUser:function () {
                    return user;
                }
            }
        });

        modalInstance.result.then(function (newUser) {
            MDBApi.user_put();
            // MLabelManager.modify({},{action:"update", id:newLabel.id, value:newLabel.value}, function (res) {
            //     get_labels();
            // });
        });
    };

    var confirm_modal = function(title, context, okFunc) {
        var modalInstance = $uibModal.open({
            templateUrl: 'CConfirm.html',
            controller: 'ConfirmCtrl',
            resolve: {
                title: function() {
                    return title;
                },
                context: function() {
                    return context;
                }
            }
        });

        modalInstance.result.then(okFunc, function () {
            console.log("modal dismissed");
        });
    };

    $scope.deleteUser = function (user) {
        confirm_modal("提示", "确定要删除【" + user.id + "】【" + user.name + "】用户?", function() {
            // MLabelManager.delete({}, {action:"delete", id:label.id}, function (res) {
            //     get_labels();
            // });
        });
    };

    load();
}]);

//controller:CDevice,for the view devices.html
monitorApp.controller('CDevices',['$scope', "$location", "$cookies", "$db_nav", "userData", "MDBApi", "NgTableParams", "$uibModal", "$db_utility",
    function($scope, $location, $cookies, $db_nav, userData, MDBApi, NgTableParams, $uibModal, $db_utility){
        $db_nav.in_devices();

        // requires user to login first.
        if (!$cookies.get("kl_monitor_user")) {
            $db_nav.go_login($location);
            return;
        }
        var load = function () {
            MDBApi.devices_load(function (data) {
                $scope.devices = data.data;
                var max_page_cnt = Math.ceil($scope.devices.length / userData.device.page_count);
                userData.device.current_page = Math.min(userData.device.current_page, max_page_cnt);

                $scope.tableParams = new NgTableParams({
                    page: userData.device.current_page,
                    count: userData.device.page_count,
                    filter: {
                        'name': userData.device.table_filter.name
                    },
                    sorting: userData.device.table_sort

                }, {
                    counts: [10, 30, 50],
                    dataset: $scope.devices
                });

                //watch search choose
                $scope.$watch(function () {
                    return $scope.tableParams.filter();
                }, function (newValue, oldValue) {
                    userData.device.table_filter = newValue;
                }, true);

                //watch page count choose
                $scope.$watch(function () {
                    return $scope.tableParams.count();
                }, function (newValue, oldValue) {
                    userData.device.page_count = newValue;
                }, true);

                //watch when page choose
                $scope.$watch(function () {
                    return $scope.tableParams.page();
                }, function (newValue, oldValue) {
                    userData.device.current_page = newValue;
                }, true);
            })
        };

        $scope.newDevice = function () {
            var modalInstance = $uibModal.open({
                animation: true,
                templateUrl: 'CCreateDevice.html',
                controller: 'CNewDevice',
                windowClass: 'kl-modal-window'
            });

            modalInstance.result.then(function (device) {
                MDBApi.device_create(device, function (res) {
                    $db_utility.log("trace", "创建设备台账【" + device.name + "】成功");
                    load();
                })
            }, function () {
                console.log('Modal dismissed at: ' + new Date());
            });
        };

        $scope.modifyDevice = function (device) {
            var modalInstance = $uibModal.open({
                templateUrl: 'CModifyDevice.html',
                controller: 'CModifyDevice',
                windowClass: 'kl-modal-window',
                resolve: {
                    oldDevice:function () {
                        return device;
                    }
                }
            });

            modalInstance.result.then(function (newDevice) {
                MDBApi.device_put({action:"update", device:newDevice}, function (res) {
                    $db_utility.log("trace", "更新设备台账【" + newDevice.id + "】成功");
                    load();
                });
            });
        };

        var confirm_modal = function(title, context, okFunc) {
            var modalInstance = $uibModal.open({
                templateUrl: 'CConfirm.html',
                controller: 'ConfirmCtrl',
                resolve: {
                    title: function() {
                        return title;
                    },
                    context: function() {
                        return context;
                    }
                }
            });

            modalInstance.result.then(okFunc, function () {
                console.log("modal dismissed");
            });
        };

        $scope.deleteDevice = function (device) {
            confirm_modal("提示", "确定要删除【" + device.id + "】【" + device.name + "】设备台账?", function() {
                MDBApi.device_put({action:"delete", device:device}, function (res) {
                    $db_utility.log("trace", "删除设备台账【" + device.id + "】成功");
                    load();
                });
            });
        };

        load();
    }]);

//controller:CDevice,for the view records.html
monitorApp.controller('CRecords',['$scope', "$location", "$cookies", "$db_nav", "userData", "MDBApi", "NgTableParams", "$uibModal", "$db_utility",
    function($scope, $location, $cookies, $db_nav, userData, MDBApi, NgTableParams, $uibModal, $db_utility){

        $db_nav.in_records();

        // requires user to login first.
        if (!$cookies.get("kl_monitor_user")) {
            $db_nav.go_login($location);
            return;
        }

        var load = function () {
            MDBApi.records_load(function (data) {
                $scope.records = data.data;
                var max_page_cnt = Math.ceil($scope.records.length / userData.record.page_count);
                userData.record.current_page = Math.min(userData.record.current_page, max_page_cnt);

                $scope.tableParams = new NgTableParams({
                    page: userData.record.current_page,
                    count: userData.record.page_count,
                    filter: {
                        'name': userData.record.table_filter.name
                    },
                    sorting: userData.record.table_sort

                }, {
                    counts: [10, 30, 50],
                    dataset: $scope.records
                });

                //watch search choose
                $scope.$watch(function () {
                    return $scope.tableParams.filter();
                }, function (newValue, oldValue) {
                    userData.record.table_filter = newValue;
                }, true);

                //watch page count choose
                $scope.$watch(function () {
                    return $scope.tableParams.count();
                }, function (newValue, oldValue) {
                    userData.record.page_count = newValue;
                }, true);

                //watch when page choose
                $scope.$watch(function () {
                    return $scope.tableParams.page();
                }, function (newValue, oldValue) {
                    userData.record.current_page = newValue;
                }, true);
            })
        };

        $scope.newRecord = function () {
            var modalInstance = $uibModal.open({
                animation: true,
                templateUrl: 'CCreateRecord.html',
                controller: 'CNewRecord',
                windowClass: 'kl-lg-modal-window'
            });

            modalInstance.result.then(function (record) {
                MDBApi.record_create(record, function (res) {
                    $db_utility.log("trace", "创建巡检记录【" + record.uid + "】成功");
                    load();
                })
            }, function () {
                console.log('Modal dismissed at: ' + new Date());
            });
        };

        $scope.modifyRecord = function (record) {
            var modalInstance = $uibModal.open({
                templateUrl: 'CModifyRecord.html',
                controller: 'CModifyRecord',
                windowClass: 'kl-lg-modal-window',
                resolve: {
                    oldRecord:function () {
                        return record;
                    }
                }
            });

            modalInstance.result.then(function (newRecord) {
                MDBApi.record_put({action:"update", record:newRecord}, function (res) {
                    $db_utility.log("trace", "更新巡检记录【" + newRecord.id + "】成功");
                    load();
                });
            });
        };

        var confirm_modal = function(title, context, okFunc) {
            var modalInstance = $uibModal.open({
                templateUrl: 'CConfirm.html',
                controller: 'ConfirmCtrl',
                resolve: {
                    title: function() {
                        return title;
                    },
                    context: function() {
                        return context;
                    }
                }
            });

            modalInstance.result.then(okFunc, function () {
                console.log("modal dismissed");
            });
        };

        $scope.deleteRecord = function (record) {
            confirm_modal("提示", "确定要删除【" + record.id + "】巡检记录?", function() {
                MDBApi.record_put({action:"delete", record:record}, function (res) {
                    $db_utility.log("trace", "删除巡检记录【" + record.id + "】成功");
                    load();
                });
            });
        };

        load();
    }]);

//controller:CLogin,for the view login.html
monitorApp.controller('CLogin',['$scope', "$location", "$cookies", "$db_nav", "$db_utility", "MDBApi",
    function($scope, $location, $cookies, $db_nav, $db_utility, MDBApi){
    if ($cookies.get("kl_monitor_user")) {
        $db_nav.back($location);
        return;
    }

    $db_utility.refresh.stop();
    $scope.on_press_enter_login = function(event) {
        if (event.keyCode !== 13) {
            return;
        }
        $scope.login();
    };
    // login to init user.
    $scope.user = $cookies.get("kl_monitor_user");
    $scope.login = function () {
        if (!$scope.user) {
            $db_utility.log("warn", "用户名不能为空.");
            return;
        }
        if (!$scope.password) {
            $db_utility.log("warn", "密码不能为空.");
            return;
        }

        MDBApi.login({name: $scope.user, passwd:$scope.password}, function (data) {
            $db_utility.log("trace", "login success.");
            $cookies.put("kl_monitor_user", $scope.user);
            $db_nav.back($location);
            return;
        });
    };
    $db_utility.log("warn", "Please login first.");

}]);

monitorApp.controller('CNewUser', ['$scope', '$uibModalInstance',
    function($scope, $uibModalInstance) {
        $scope.newUser = {
            name: "",
            passwd: ""
        };

        $scope.valid = function () {
            return $scope.newUser.name != "" && $scope.newUser.passwd != "";
        };

        $scope.confirm = function () {
            $uibModalInstance.close($scope.newUser);
        };

        $scope.cancel = function () {
            $uibModalInstance.dismiss('cancel');
        };
    }
]);

monitorApp.controller('CModifyUser', ['$scope', '$uibModalInstance', 'oldUser',
    function($scope, $uibModalInstance, oldUser){
        $scope.newUser = oldUser;
        $scope.confirm = function () {
            $uibModalInstance.close($scope.newUser);
        };

        $scope.cancel = function () {
            $uibModalInstance.dismiss('cancel');
        };
    }
]);

monitorApp.controller('CNewDevice', ['$scope', '$uibModalInstance',
    function($scope, $uibModalInstance) {
        $scope.newDevice = {
            name: "",
            uid: "",
            specification: "",
            precision: "",
            unit: "",
            min: 0,
            max: 1000
        };

        $scope.valid = function () {
            return $scope.newDevice.name != "" && $scope.newDevice.uid != "" && $scope.newDevice.specification != "" && $scope.newDevice.precision != "";
        };

        $scope.confirm = function () {
            $uibModalInstance.close($scope.newDevice);
        };

        $scope.cancel = function () {
            $uibModalInstance.dismiss('cancel');
        };
    }
]);

monitorApp.controller('CModifyDevice', ['$scope', '$uibModalInstance', 'oldDevice',
    function($scope, $uibModalInstance, oldDevice){
        $scope.newDevice = oldDevice;

        $scope.valid = function () {
            return $scope.newDevice.name != "" && $scope.newDevice.uid != "" && $scope.newDevice.specification != "" && $scope.newDevice.precision != "";
        };

        $scope.confirm = function () {
            $uibModalInstance.close($scope.newDevice);
        };

        $scope.cancel = function () {
            $uibModalInstance.dismiss('cancel');
        };
    }
]);

monitorApp.controller('CNewRecord', ['$scope', '$uibModalInstance',
    function($scope, $uibModalInstance) {
        $scope.watchers = "";
        $scope.create_time = new Date();
        $scope.newRecord = {
            uid: "",
            value: 0,
            watchers: "",
            desc: ""
        };

        $scope.valid = function () {
            return $scope.newRecord.uid != "" && $scope.watchers != "";
        };

        $scope.confirm = function () {
            $scope.newRecord.watchers = $scope.watchers.split(",");
            $scope.newRecord.time = Math.round($scope.create_time.getTime() / 1000);
            $uibModalInstance.close($scope.newRecord);
        };

        $scope.cancel = function () {
            $uibModalInstance.dismiss('cancel');
        };
    }
]);

monitorApp.controller('CModifyRecord', ['$scope', '$uibModalInstance', 'oldRecord',
    function($scope, $uibModalInstance, oldRecord){
        $scope.watchers = oldRecord.watchers.join(",");
        $scope.create_time = new Date();
        $scope.create_time.setTime(Number(oldRecord.time) * 1000);
        $scope.newRecord = oldRecord;

        $scope.valid = function () {
            return $scope.newRecord.uid != "" && $scope.watchers != "";
        };

        $scope.confirm = function () {
            $scope.newRecord.watchers = $scope.watchers.split(",");
            $scope.newRecord.time = Math.round($scope.create_time.getTime() / 1000);
            $uibModalInstance.close($scope.newRecord);
        };

        $scope.cancel = function () {
            $uibModalInstance.dismiss('cancel');
        };
    }
]);

// controller: ConfirmCtrl
monitorApp.controller('ConfirmCtrl', ['$scope', '$uibModalInstance', 'title', 'context',
    function($scope, $uibModalInstance, title, context){
        $scope.title = title;
        $scope.context = context;
        $scope.confirm = function () {
            $uibModalInstance.close();
        };

        $scope.cancel = function () {
            $uibModalInstance.dismiss('cancel');
        };
    }
]);

monitorApp.filter("db_filter_log_level", function () {
    return function (v) {
        return (v == "warn" || v == "error") ? "alert-warn" : "alert-success";
    };
});

monitorApp.filter("filter_record_watchers", function () {
    return function (v) {
        return v.join(", ")
    }
});

monitorApp.filter("filter_record_time", function () {
    return function (v) {
        return absolute_seconds_to_YYYYmmdd_hhmm(v)
    }
});

monitorApp.filter('date_button_active', function() {
        return function(is_active) {
            return is_active? "btn btn-primary": "btn btn-default";
        };
    });

monitorApp.factory("MDBApi", ["$http", function ($http) {
    return {
        users_load: function (success) {
            $http.get("/api/v1/users").success(success);
        },
        user_create: function (r, success) {
            $http.post("/api/v1/users", r).success(success);
        },
        user_put: function (r, success) {
            $http.put("/api/v1/users", r).success(success);
        },
        devices_load: function (success) {
            $http.get("/api/v1/devices").success(success);
        },
        device_create: function (r, success) {
            $http.post("/api/v1/devices", r).success(success);
        },
        device_put: function (r, success) {
            $http.put("/api/v1/devices", r).success(success);
        },
        records_load: function (success) {
            $http.get("/api/v1/records").success(success);
        },
        record_create: function (r, success) {
            $http.post("/api/v1/records", r).success(success);
        },
        record_put: function (r, success) {
            $http.put("/api/v1/records", r).success(success);
        },
        analysis_query: function (uid, from, to, success) {
            $http.get("/api/v1/analysis", {params:{uid:uid, from:from, to:to}}).success(success);
        },
        query: function (success) {
            $http.get("/api/v1/query").success(success);
        },
        login: function (r, success) {
            $http.post("/api/v1/login", r).success(success);
        }
    };
}]);

// the sc nav is the nevigator
monitorApp.provider("$db_nav", function () {
    this.$get = function () {
        return {
            selected: "/analysis",
            in_analysis: function () {
                this.selected = "/analysis";
            },
            in_user: function () {
                this.selected = "/user";
            },
            in_devices: function () {
                this.selected = "/devices";
            },
            in_records: function () {
                this.selected = "/records";
            },
            jump: function (to, $location) {
                $location.path(to);
            },
            back: function ($location) {
                $location.path(this.selected);
            },
            go_analysis: function ($location) {
                $location.path("/analysis");
            },
            go_user:  function ($location) {
                $location.path("/user");
            },
            go_devices:  function ($location) {
                $location.path("/devices");
            },
            go_records:  function ($location) {
                $location.path("/records");
            },
            go_login: function ($location) {
                $location.path("/login");
            },
            is_selected: function (v) {
                return v == this.selected;
            }
        };
    };
});

// the db utility is a set of helper utilities.
monitorApp.provider("$db_utility", function () {
    this.$get = ["$rootScope", function ($rootScope) {
        return {
            log: function (level, msg) {
                $rootScope.$broadcast("$db_utility_log", level, msg);
            },
            http_error: function (status, response) {
                $rootScope.$broadcast("$db_utility_http_error", status, response);
            },
            refresh: async_refresh2
        };
    }];
});

// config the http interceptor.
monitorApp.config(['$httpProvider', function ($httpProvider) {
    $httpProvider.interceptors.push('MHttpInterceptor');
}]);

// the http interceptor.
monitorApp.factory('MHttpInterceptor', ["$q", "$db_utility", "$location", "$cookies", function ($q, $db_utility, $location, $cookies) {
    // register the interceptor as a service
    // @see: https://code.angularjs.org/1.2.0-rc.3/docs/api/ng.$http
    // @remark: the function($q) should never add other params.
    return {
        'request': function (config) {
            return config || $q.when(config);
        },
        'requestError': function (rejection) {
            return $q.reject(rejection);
        },
        'response': function (response) {
            if (response.data.code && response.data.code != 0) {
                //$db_utility.http_error(response.status, response.data);
                // the $q.reject, will cause the error function of controller.
                // @see: https://code.angularjs.org/1.2.0-rc.3/docs/api/ng.$q
                return $q.reject(response);
            }
            return response || $q.when(response);
        },
        'responseError': function (rejection) {
            $db_utility.http_error(rejection.status, rejection.data);
            if(rejection.status === 401) {
                $cookies.remove("kl_monitor_user");
                $location.path('/login');
                return $q.reject(rejection);
            }
            else {
                return $q.reject(rejection);
            }
        }
    };
}]);
// Please note that $uibModalInstance represents a modal window (instance) dependency.
// It is not the same as the $uibModal service used above.
