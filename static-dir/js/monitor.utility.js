function Errors() {
}
Errors.Success = 0;
// errors >= 10000 are ui errors.
Errors.UIApiError = 10000;
Errors.UIUnAuthoriezed = 10001;
Errors.UINotFound = 10002;
Errors.SESSIONTIMEOUTError = 107;
// resolve error code
Errors.resolve = function(code, status) {
    var err_map = {
    };
    err_map[Errors.UIUnAuthoriezed] = "您没有登录或者登录超时，请重新登录";
    err_map[Errors.UINotFound] = "访问的资源不存在";
    err_map[Errors.UIApiError] = "服务器错误";
    err_map[Errors.SESSIONTIMEOUTError] = "系统: 找不到session中的api_key";
    var msg = "";

    // show status code when http unknown error.
    if (code == Errors.UIApiError) {
        msg += "HTTP：" + status;
    }

    // hide the detail when http known error.
    if (code <= Errors.UIApiError) {
        msg += "原因：" + code;
    }

    return msg;
};
/**
 * global error function for backend.
 * @param $location, the angularjs $location param, for page jump.
 * @param code the error code, from backend or js error code.
 * @param status http status code, if code is http unknown error, show it.
 */
function monitor_on_error($location, code, status) {
    // we parse the http error to system error code.
    var http_known_error = {
        401: Errors.UIUnAuthoriezed,
        404: Errors.UINotFound
    };
    if (code == Errors.UIApiError && http_known_error[status]) {
        code = http_known_error[status];
    }

    // process the system error.
    if (code == Errors.UIUnAuthoriezed || code == Errors.SESSIONTIMEOUTError) {
        logs.warn("请登录");
        jmp_to_user_login_page($location);
        return code;
    }

    // show error message to log. ignore errors:
    // 406: user test, has not login yet.
    if (code != 406) {
        var err_msg = Errors.resolve(code, status);
        console.log(err_msg);
    }
    return code;
}

function absolute_seconds_to_YYYYmmdd_hhmm(seconds) {
    var date = new Date();
    date.setTime(Number(seconds) * 1000);

    var ret = date.getFullYear()
        + "-" + padding(date.getMonth() + 1, 2, '0')
        + "-" + padding(date.getDate(), 2, '0')
        + " " + padding(date.getHours(), 2, '0')
        + ":" + padding(date.getMinutes(), 2, '0');

    return ret;
}

function get_today_date_objs() {
    var result = {};
    var date = new Date();
    result.ts = Math.round(date.getTime() / 1000);
    result.value = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    return result;
}

function get_before_days_date_objs(day) {
    var result = {};
    var date = new Date();
    date.setDate(date.getDate() - day);
    date.setHours(0);
    date.setMinutes(0);
    date.setSeconds(0);
    date.setMilliseconds(0);
    result.ts = Math.round(date.getTime() / 1000);
    result.value = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    return result;
}