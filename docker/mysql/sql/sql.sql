create database if not exists `go_im`;

use go_im;

CREATE TABLE if not exists `user`
(
    `id`         int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `username`   varchar(20)  NOT NULL COMMENT '用户名',
    `nickname`   varchar(20)  NOT NULL DEFAULT '' COMMENT '用户昵称',
    `password`   varchar(100) not null default '' comment '密码',
    `created_at` timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci
  ROW_FORMAT = DYNAMIC COMMENT ='用户表';