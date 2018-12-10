DROP TABLE IF EXISTS `APP_CONFIG` ;

CREATE TABLE `APP_CONFIG` (
  `CONFIG_NAME` varchar(100) NOT NULL COMMENT '配置参数名',
  `CONFIG_VALUE` text NOT NULL COMMENT '配置参数值',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`CONFIG_NAME`)
) COMMENT='应用配置表';


DROP TABLE IF EXISTS `WECHAT_API_TOKEN_CONFIG` ;

CREATE TABLE `WECHAT_API_TOKEN_CONFIG` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `APP_ID` varchar(100) NOT NULL COMMENT '公众号APP_ID',
  `APP_SECRET` varchar(100) NOT NULL COMMENT '公众号APP_SECRET',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='微信公众号接口access_token获取参数配置表';


DROP TABLE IF EXISTS `WECHAT_API_TOKEN` ;

CREATE TABLE `WECHAT_API_TOKEN` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `APP_ID` varchar(100) NOT NULL COMMENT '公众号APP_ID',
  `ACCESS_TOKEN` text COMMENT '公众号ACCESS_TOKEN',
  `UPDATED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否最新记录 0-否 1-是',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `EXPIRE_TIME` timestamp COMMENT '过期时间',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='微信公众号接口access_token记录表';


DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_CONFIG` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_CONFIG` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `APP_SECRET` varchar(100) NOT NULL COMMENT '第三方平台APP_SECRET',
  `TOKEN` varchar(100) NOT NULL COMMENT '第三方平台接收消息的校验TOKEN',
  `AES_KEY` varchar(43) NOT NULL COMMENT '第三方平台接收消息的AES加密Key',
  `REDIRECT_URL` text COMMENT '第三方平台授权回调URL',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='微信第三方平台配置表';


DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_TICKET` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_TICKET` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `TICKET` varchar(100) NOT NULL COMMENT '第三方平台component_verify_ticket',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='微信第三方平台component_verify_ticket记录表';


DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_TOKEN` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_TOKEN` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `COMPONENT_ACCESS_TOKEN` text COMMENT '第三方平台component_access_token',
  `UPDATED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否最新记录 0-否 1-是',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `EXPIRE_TIME` timestamp COMMENT '过期时间',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='第三方平台component_access_token记录表';


DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_AUTHORIZER` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_AUTHORIZER` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `AUTHORIZER_APP_ID` varchar(100) NOT NULL COMMENT '授权方APP_ID',
  `AUTHORIZATION_STATE` tinyint(3) NOT NULL DEFAULT '1' COMMENT '授权状态 0-未授权 1-已授权',
  `AUTHORIZATION_CODE` text COMMENT '授权码(code)',
  `PRE_AUTH_CODE` text COMMENT '预授权码',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`CODE_NAME`, `AUTHORIZER_APP_ID`)
) COMMENT='第三方平台授权方状态表';


DROP TABLE IF EXISTS `WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN` ;

CREATE TABLE `WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `APP_ID` varchar(100) NOT NULL COMMENT '第三方平台APP_ID',
  `AUTHORIZER_APP_ID` varchar(100) NOT NULL COMMENT '授权方APP_ID',
  `AUTHORIZER_ACCESS_TOKEN` text COMMENT '授权方接口调用凭据',
  `AUTHORIZER_REFRESH_TOKEN` text COMMENT '接口调用凭据刷新令牌',
  `UPDATED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '是否最新记录 0-否 1-是',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `EXPIRE_TIME` timestamp COMMENT '过期时间',
  PRIMARY KEY (`CODE_NAME`, `AUTHORIZER_APP_ID`)
) COMMENT='第三方平台授权方状态表';


DROP TABLE IF EXISTS `WECHAT_CORP_TOKEN_CONFIG` ;

CREATE TABLE `WECHAT_CORP_TOKEN_CONFIG` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `CORP_ID` varchar(100) NOT NULL COMMENT '企业ID',
  `CORP_SECRET` varchar(100) NOT NULL COMMENT '应用的凭证密钥',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='企业微信接口access_token获取参数配置表';


DROP TABLE IF EXISTS `WECHAT_CORP_TOKEN` ;

CREATE TABLE `WECHAT_CORP_TOKEN` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `CORP_ID` varchar(100) NOT NULL COMMENT '企业ID',
  `ACCESS_TOKEN` text COMMENT '企业微信ACCESS_TOKEN',
  `UPDATE_TIME` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `EXPIRE_TIME` timestamp COMMENT '过期时间',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='企业微信接口access_token记录表';


DROP TABLE IF EXISTS `WECHAT_CORP_THIRD_PLATFORM_CONFIG` ;

CREATE TABLE `WECHAT_CORP_THIRD_PLATFORM_CONFIG` (
  `CODE_NAME` varchar(42) NOT NULL COMMENT '代号',
  `SUITE_ID` varchar(100) NOT NULL COMMENT '企业微信第三方应用SUITE_ID',
  `SUITE_SECRET` varchar(100) NOT NULL COMMENT '企业微信第三方应用SUITE_SECRET',
  `TOKEN` varchar(100) NOT NULL COMMENT '第三方应用接收消息的校验TOKEN',
  `AES_KEY` varchar(43) NOT NULL COMMENT '第三方应用接收消息的AES加密Key',
  `REDIRECT_URL` text COMMENT '第三方应用授权回调URL',
  `ENABLED` tinyint(3) NOT NULL DEFAULT '1' COMMENT '有效状态 0-无效 1-有效',
  PRIMARY KEY (`CODE_NAME`)
) COMMENT='企业微信第三方应用配置表';
