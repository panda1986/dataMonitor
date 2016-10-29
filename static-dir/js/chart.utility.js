/**
 * Created by panda on 2015/3/27.
 */
/*
 * transform_data_to_server_samples
 * 将从server得到的监控数据转换为折线图指令需要的多维数据格式
 */
var transform_data_to_pie_samples = function (data, property) {
    var pie_chart_samples = {
        title: property.title,
        samples: []
    };
    for(var item in data) {
        if (data[item].clients> 0) {
            var sample = {y: data[item].clients, x: get_isp_mapping_info(item)};
            pie_chart_samples.samples.push(sample);
        }
    }
    return pie_chart_samples;
};

var transform_data_to_map_samples = function (data, property) {
    var map_dimensional_samples = {
        title: property.title,
        legend: property.title,                      //数据标签
        dataRange_min: 0,                          //值域选择的最小值
        dataRange_max: 0,                          //值域选择的最大值
        mark_item: "",
        samples: [],
        unit: property.unit
    };
    var max_data_range = 0;
    var min_data_range = 0;

    for(var item in data) {
        if (data[item].clients == 0) {
            continue
        }
        var sample = {x: get_area_mapping_info(item), y: data[item].clients};
        if (!max_data_range) {
            max_data_range = data[item].clients;
            map_dimensional_samples.mark_item = get_area_mapping_info(item);
        } else {
            if (data[item].clients > max_data_range) {
                max_data_range = data[item].clients;
                map_dimensional_samples.mark_item = get_area_mapping_info(item);
            }
        }

        if (!min_data_range) {
            min_data_range = data[item].clients;
        } else {
            if (data[item].clients < min_data_range) {
                min_data_range = data[item].clients;
            }
        }
        map_dimensional_samples.samples.push(sample);
    }
    map_dimensional_samples.dataRange_max = max_data_range;
    map_dimensional_samples.dataRange_min = min_data_range;
    return map_dimensional_samples;
};

var echart_set_pie_option_data = function (chart) {
    var option = {
        title : {
            text: chart.title,
            x:'left'
        },
        tooltip : {                          // 气泡提示配置
            trigger: 'item',                        // 触发类型，数据触发
            formatter: "{a} <br/>{b} : {c} ({d}%)"
        },
        toolbox: {
            show: true,
            orient : 'vertical',
            x: 'right',
            y: 'center',
            feature : {
                dataView : {show: true, readOnly: true},
                saveAsImage : {show: true}
            }
        },
        calculable : true,
        series : [
            {
                name: chart.title,
                type: 'pie',
                radius: '55%',
                center: ['50%', '60%'],
                data: []
            }
        ],
        color : [
            '#2ec7c9','#b6a2de','#5ab1ef','#ffb980','#d87a80',
            '#8d98b3','#e5cf0d','#97b552','#95706d','#dc69aa',
            '#07a2a4','#9a7fd1','#588dd5','#f5994e','#c05050',
            '#59678c','#c9ab00','#7eb00a','#6f5553','#c14089'
        ]
    };
    for(var i = 0; i < chart.samples.length; i ++){
        var sample = chart.samples[i];
        option.series[0].data.push({name:sample.x, value:sample.y});
    }
    return option;
};

var echart_set_map_option_data = function (chart) {
    var option = {
        title : {
            text: chart.title,
            x:'left'
        },
        tooltip : {
            trigger: 'item'
        },
        dataRange: {
            min: chart.dataRange_min,
            max: chart.dataRange_max,
            x: 'left',
            y: 'bottom',
            text:[(chart.dataRange_max).toString() + chart.unit, chart.dataRange_min.toString() + chart.unit],
            color: ["#006edd", "#e0ffff"],
            calculable: false
        },
        toolbox: {
            show: true,
            orient : 'vertical',
            x: 'right',
            y: 'center',
            feature : {
                dataView : {show: true, readOnly: true},
                saveAsImage : {show: true}
            }
        },
        series : [
            {
                name: chart.title,
                type: 'map',
                mapType: 'china',
                roam: false,
                itemStyle:{
                    normal:{label:{show:true}},
                    emphasis:{label:{show:true}}
                },
                data:[]
            }
        ]
    };
    for(var i = 0; i < chart.samples.length; i ++){
        var sample = chart.samples[i];
        option.series[0].data.push({name:sample.x, value:sample.y});
    }
    return option;
};