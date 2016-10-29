-- author:panda

CREATE DATABASE IF not EXISTS kl_data_monitor CHARACTER SET utf8 COLLATE utf8_general_ci;

USE kl_data_monitor;
SET NAMES 'utf8';

CREATE TABLE `user_data` (
	`id` int(32) NOT NULL AUTO_INCREMENT COMMENT '自动生成的用户id',
	`name` varchar(32) NOT NULL DEFAULT '' COMMENT '用户名，支持中文',
	`passwd` varchar(32) NOT NULL DEFAULT '' COMMENT '密码',
	`role` enum('administror','lv0') NOT NULL DEFAULT 'lv0' COMMENT '角色',
	PRIMARY KEY (`id`, `name`),
	UNIQUE `user_name` USING BTREE (`name`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8 AUTO_INCREMENT=0 ;

CREATE TABLE `records` (
	`id` int(32) NOT NULL AUTO_INCREMENT,
	`uid` varchar(64) NOT NULL COMMENT '记录的设备编号',
	`value` float(64,2) NOT NULL COMMENT '记录值',
	`watchers` varchar(64) NOT NULL COMMENT '记录检查员',
	`user_id` int(32) NOT NULL COMMENT '记录用户id',
	`time` int(64) NOT NULL COMMENT '记录时间',
	`desc` varchar(256) NOT NULL COMMENT '记录描述',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8 AUTO_INCREMENT=1000;

CREATE TABLE `devices` (
	`id` int(32) NOT NULL AUTO_INCREMENT,
	`name` varchar(32) NOT NULL COMMENT '计量器具名称',
	`uid` varchar(64) NOT NULL COMMENT '统一编号（位号）',
	`specification` varchar(64) NOT NULL,
	`precision` varchar(64) NOT NULL COMMENT '最大误差/准确度等级/分度值',
	`unit` varchar(32) NOT NULL COMMENT '测量单位',
	`min` float(32,2) NOT NULL COMMENT '测量最小值',
	`max` float(32,2) NOT NULL COMMENT '测量最大值',
	PRIMARY KEY (`id`, `uid`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8  AUTO_INCREMENT=200;

insert into `kl_data_monitor`.`user_data` ( `name`, `id`, `role`, `passwd`) values ( 'admin', '0', 'administror', 'kaloon2016');
