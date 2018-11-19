DROP TABLE IF EXISTS `APP_CONFIG` ;

CREATE TABLE `APP_CONFIG` (
  `CONFIG_NAME` varchar(100) NOT NULL COMMENT '配置参数名',
  `CONFIG_VALUE` text NOT NULL COMMENT '配置参数值',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`CONFIG_NAME`)
) COMMENT='应用配置表';

DROP TABLE IF EXISTS `WECHAT_API_TOKEN_CONFIG` ;

CREATE TABLE `WECHAT_API_TOKEN_CONFIG` (
  `APP_ID` varchar(100) NOT NULL COMMENT '公众号APP_ID',
  `APP_SECRET` varchar(100) NOT NULL COMMENT '公众号APP_SECRET',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`APP_ID`)
) COMMENT='微信公众号接口access_token获取参数配置表';

DROP TABLE IF EXISTS `WECHAT_API_TOKEN` ;

CREATE TABLE `WECHAT_API_TOKEN` (
  `APP_ID` varchar(100) NOT NULL COMMENT '公众号APP_ID',
  `ACCESS_TOKEN` text COMMENT '公众号ACCESS_TOKEN',
  `UPDATED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否最新记录 0-否 1-是',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `EXPIRE_TIME` timestamp COMMENT '过期时间',
  PRIMARY KEY (`APP_ID`)
) COMMENT='微信公众号接口access_token记录表';

DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_CONFIG` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_CONFIG` (
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `APP_SECRET` varchar(100) NOT NULL COMMENT '第三方平台APP_SECRET',
  `TOKEN` varchar(100) NOT NULL COMMENT '第三方平台接收消息的校验TOKEN',
  `AES_KEY` varchar(43) NOT NULL COMMENT '第三方平台接收消息的AES加密Key',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`APP_ID`)
) COMMENT='微信第三方平台配置表';

DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_TICKET` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_TICKET` (
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `TICKET` varchar(100) NOT NULL COMMENT '第三方平台component_verify_ticket',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`APP_ID`)
) COMMENT='微信第三方平台component_verify_ticket记录表';

DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_TOKEN` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_TOKEN` (
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `COMPONENT_ACCESS_TOKEN` text COMMENT '第三方平台component_access_token',
  `UPDATED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否最新记录 0-否 1-是',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `EXPIRE_TIME` timestamp COMMENT '过期时间',
  PRIMARY KEY (`APP_ID`)
) COMMENT='第三方平台component_access_token记录表';

DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE` (
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `PRE_AUTH_CODE` text COMMENT '第三方平台pre_auth_code',
  `UPDATED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否最新记录 0-否 1-是',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `EXPIRE_TIME` timestamp COMMENT '过期时间',
  PRIMARY KEY (`APP_ID`)
) COMMENT='第三方平台pre_auth_code记录表';
