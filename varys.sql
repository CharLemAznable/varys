
-- WECHAT_APP_TOKEN begin --

drop table if exists `wechat_app_config` ;

create table `wechat_app_config` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '公众号APP_ID',
  `app_secret` varchar(100) not null comment '公众号APP_SECRET',
  `enabled` tinyint not null default '1' comment '有效状态 0-无效 1-有效',
  primary key (`code_name`)
) comment='微信公众号接口access_token获取参数配置表';


drop table if exists `wechat_app_token` ;

create table `wechat_app_token` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '公众号APP_ID',
  `access_token` text comment '公众号ACCESS_TOKEN',
  `jsapi_ticket` text comment '公众号用于调用微信JS接口的JSAPI_TICKET',
  `updated` tinyint not null default '1' comment '是否最新记录 0-否 1-是',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`)
) comment='微信公众号接口access_token记录表';

-- WECHAT_APP_TOKEN end --


-- WECHAT_APP_THIRD_PLATFORM_TOKEN begin --

drop table if exists `wechat_app_third_platform_config` ;

create table `wechat_app_third_platform_config` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '第三方平台APP_ID',
  `app_secret` varchar(100) not null comment '第三方平台APP_SECRET',
  `token` varchar(100) not null comment '第三方平台接收消息的校验TOKEN',
  `aes_key` varchar(43) not null comment '第三方平台接收消息的AES加密Key',
  `auth_forward_url` text comment '第三方平台授权事件消息转发URL, varys将授权事件消息转发给业务服务做其他处理',
  `redirect_url` text comment '第三方平台授权回调URL',
  `enabled` tinyint not null default '1' comment '有效状态 0-无效 1-有效',
  primary key (`code_name`)
) comment='微信第三方平台配置表';


drop table if exists `wechat_app_third_platform_ticket` ;

create table `wechat_app_third_platform_ticket` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '第三方平台APP_ID',
  `ticket` varchar(100) not null comment '第三方平台component_verify_ticket',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  primary key (`code_name`)
) comment='微信第三方平台component_verify_ticket记录表';


drop table if exists `wechat_app_third_platform_token` ;

create table `wechat_app_third_platform_token` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '第三方平台APP_ID',
  `access_token` text comment '第三方平台component_access_token',
  `updated` tinyint not null default '1' comment '是否最新记录 0-否 1-是',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`)
) comment='第三方平台component_access_token记录表';


drop table if exists `wechat_app_third_platform_authorizer` ;

create table `wechat_app_third_platform_authorizer` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '第三方平台APP_ID',
  `authorizer_app_id` varchar(100) not null comment '授权方APP_ID',
  `authorization_state` tinyint not null default '1' comment '授权状态 0-未授权 1-已授权',
  `authorization_code` text comment '授权码(code)',
  `pre_auth_code` text comment '预授权码',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  primary key (`code_name`, `authorizer_app_id`)
) comment='第三方平台授权方状态表';


drop table if exists `wechat_app_third_platform_authorizer_token` ;

create table `wechat_app_third_platform_authorizer_token` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '第三方平台APP_ID',
  `authorizer_app_id` varchar(100) not null comment '授权方APP_ID',
  `authorizer_access_token` text comment '授权方接口调用凭据',
  `authorizer_refresh_token` text comment '接口调用凭据刷新令牌',
  `authorizer_jsapi_ticket` text comment '授权方用于调用微信JS接口的JSAPI_TICKET',
  `updated` tinyint not null default '1' comment '是否最新记录 0-否 1-是',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`, `authorizer_app_id`)
) comment='第三方平台授权方access_token记录表';

-- WECHAT_APP_THIRD_PLATFORM_TOKEN end --


-- WECHAT_CORP_TOKEN begin --

drop table if exists `wechat_corp_config` ;

create table `wechat_corp_config` (
  `code_name` varchar(42) not null comment '代号',
  `corp_id` varchar(100) not null comment '企业ID',
  `corp_secret` varchar(100) not null comment '应用的凭证密钥',
  `enabled` tinyint not null default '1' comment '有效状态 0-无效 1-有效',
  primary key (`code_name`)
) comment='企业微信接口access_token获取参数配置表';


drop table if exists `wechat_corp_token` ;

create table `wechat_corp_token` (
  `code_name` varchar(42) not null comment '代号',
  `corp_id` varchar(100) not null comment '企业ID',
  `access_token` text comment '企业微信ACCESS_TOKEN',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`)
) comment='企业微信接口access_token记录表';

-- WECHAT_CORP_TOKEN end --


-- WECHAT_CORP_THIRD_PLATFORM_TOKEN begin --

drop table if exists `wechat_corp_third_platform_config` ;

create table `wechat_corp_third_platform_config` (
  `code_name` varchar(42) not null comment '代号',
  `suite_id` varchar(100) not null comment '企业微信第三方应用SUITE_ID',
  `suite_secret` varchar(100) not null comment '企业微信第三方应用SUITE_SECRET',
  `token` varchar(100) not null comment '第三方应用接收消息的校验TOKEN',
  `aes_key` varchar(43) not null comment '第三方应用接收消息的AES加密Key',
  `redirect_url` text comment '第三方应用授权回调URL',
  `enabled` tinyint not null default '1' comment '有效状态 0-无效 1-有效',
  primary key (`code_name`)
) comment='企业微信第三方应用配置表';


drop table if exists `wechat_corp_third_platform_ticket` ;

create table `wechat_corp_third_platform_ticket` (
  `code_name` varchar(42) not null comment '代号',
  `suite_id` varchar(100) not null comment '企业微信第三方应用SUITE_ID',
  `ticket` varchar(100) not null comment '第三方应用suite_ticket',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  primary key (`code_name`)
) comment='企业微信第三方应用suite_ticket记录表';


drop table if exists `wechat_corp_third_platform_token` ;

create table `wechat_corp_third_platform_token` (
  `code_name` varchar(42) not null comment '代号',
  `suite_id` varchar(100) not null comment '企业微信第三方应用SUITE_ID',
  `access_token` text comment '企业微信第三方应用suite_access_token',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`)
) comment='企业微信第三方应用suite_access_token记录表';


drop table if exists `wechat_corp_third_platform_authorizer` ;

create table `wechat_corp_third_platform_authorizer` (
  `code_name` varchar(42) not null comment '代号',
  `suite_id` varchar(100) not null comment '企业微信第三方应用SUITE_ID',
  `corp_id` varchar(100) not null comment '授权方企业微信ID',
  `state` tinyint not null default '1' comment '授权状态 0-未授权 1-已授权',
  `permanent_code` text comment '企业微信永久授权码',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  primary key (`code_name`, `corp_id`)
) comment='企业微信授权方状态表';


drop table if exists `wechat_corp_third_platform_corp_token` ;

create table `wechat_corp_third_platform_corp_token` (
  `code_name` varchar(42) not null comment '代号',
  `suite_id` varchar(100) not null comment '企业微信第三方应用SUITE_ID',
  `corp_id` varchar(100) not null comment '授权方企业微信ID',
  `corp_access_token` text comment '企业微信授权方企业access_token',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`, `corp_id`)
) comment='企业微信授权方企业access_token记录表';

-- WECHAT_CORP_THIRD_PLATFORM_TOKEN end --


-- TOUTIAO_APP_TOKEN begin --

drop table if exists `toutiao_app_config` ;

create table `toutiao_app_config` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '字节小程序APP_ID',
  `app_secret` varchar(100) not null comment '字节小程序APP_SECRET',
  `enabled` tinyint not null default '1' comment '有效状态 0-无效 1-有效',
  primary key (`code_name`)
) comment='字节小程序access_token获取参数配置表';


drop table if exists `toutiao_app_token` ;

create table `toutiao_app_token` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '字节小程序APP_ID',
  `access_token` text comment '字节小程序ACCESS_TOKEN',
  `updated` tinyint not null default '1' comment '是否最新记录 0-否 1-是',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`)
) comment='字节小程序access_token记录表';

-- TOUTIAO_APP_TOKEN end --


-- FENGNIAO_APP_TOKEN begin --

drop table if exists `fengniao_app_config` ;

create table `fengniao_app_config` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '蜂鸟商户APP_ID',
  `secret_key` varchar(100) not null comment '蜂鸟商户SECRET_KEY',
  `callback_order_url` text comment '订单状态变更回调URL',
  `enabled` tinyint not null default '1' comment '有效状态 0-无效 1-有效',
  primary key (`code_name`)
) comment='蜂鸟商户access_token获取参数配置表';


drop table if exists `fengniao_app_token` ;

create table `toutiao_app_token` (
  `code_name` varchar(42) not null comment '代号',
  `app_id` varchar(100) not null comment '蜂鸟商户APP_ID',
  `access_token` text comment '蜂鸟商户ACCESS_TOKEN',
  `updated` tinyint not null default '1' comment '是否最新记录 0-否 1-是',
  `update_time` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
  `expire_time` timestamp comment '过期时间',
  primary key (`code_name`)
) comment='蜂鸟商户access_token记录表';

-- FENGNIAO_APP_TOKEN end --
