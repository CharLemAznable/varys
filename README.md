## varys

[![Build Status](https://travis-ci.org/CharLemAznable/varys.svg?branch=master)](https://travis-ci.org/CharLemAznable/varys)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/CharLemAznable/varys)
[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)
![GitHub code size](https://img.shields.io/github/languages/code-size/CharLemAznable/varys)

AccessToken 中控服务器

统一DB存储AccessToken, 支持分布式部署服务访问和更新.

#### 配置文件

1. ```appConfig.toml```

```toml
Port = 4236
ContextPath = ""
ConnectName = "Default"
```

2. ```logback.xml```

```xml
<logging>
    <filter enabled="true">
        <tag>file</tag>
        <type>file</type>
        <level>INFO</level>
        <property name="filename">varys.log</property>
        <property name="format">[%D %T] [%L] (%S) %M</property>
        <property name="rotate">false</property>
        <property name="maxsize">0M</property>
        <property name="maxlines">0K</property>
        <property name="daily">false</property>
    </filter>
</logging>
```

3. ```gql.yaml```

```yaml
Default:
  DriverName:       mysql
  DataSourceName:   admin:test123@tcp(127.0.0.1:3306)/rock?charset=utf8
  MaxOpenConns:     50
  MaxIdleConns:     1
  ConnMaxLifetime:  60
```

#### 部署执行

1. 下载最新的可执行文件压缩包并解压

    下载地址: [varys release](https://github.com/CharLemAznable/varys/releases)

```bash
$ tar -xvJf varys-[version].[arch].[os].tar.xz
```

2. 新建/编辑配置文件, 启动运行

```bash
$ nohup ./varys-[version].[arch].[os].bin &
```

#### 数据库

建表SQL:

  [varys.sql](https://github.com/CharLemAznable/varys/blob/master/varys.sql)

#### 本地缓存

包含微信公众号配置缓存和```access_token```缓存, 其中:

1) 公众号配置缓存默认1小时
2) access_token缓存默认5分钟, 当access_token即将过期并被其他分布式节点更新时缓存1分钟

  [app_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/app_token_cache.go)

包含微信第三方平台配置缓存和报文解密器缓存, 其中

1) 第三方平台配置缓存默认1小时
2) 第三方平台报文解密器缓存默认1小时

包含微信第三方平台```component_access_token```/```authorizer_access_token```缓存, 其中

1) component_access_token缓存默认5分钟, 当component_access_token即将过期并被其他分布式节点更新时缓存1分钟
2) authorizer_access_token缓存默认5分钟, 当authorizer_access_token即将过期并被其他分布式节点更新时缓存1分钟

  [app_third_platform_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/app_third_platform_token_cache.go)

包含企业微信配置缓存和```access_token```缓存, 其中:

1) 企业微信配置缓存默认1小时
2) access_token缓存最大5分钟, 当access_token即将过期时, 缓存时间最大至其有效期结束

  [corp_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/corp_token_cache.go)

包含企业微信第三方应用配置缓存和报文解密器缓存, 其中

1) 企业微信第三方应用配置缓存默认1小时
2) 企业微信第三方应用报文解密器缓存默认1小时

包含企业微信第三方应用```suite_access_token```/```access_token```缓存, 其中

1) suite_access_token缓存最大5分钟, 当suite_access_token即将过期时, 缓存时间最大至其有效期结束
2) access_token缓存最大5分钟, 当access_token即将过期时, 缓存时间最大至其有效期结束

  [corp_third_platform_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/corp_third_platform_token_cache.go)

#### 访问路径

默认服务地址:
```http
http://localhost:4236
```
```http
/query-wechat-app-token/{codeName:string}

获取指定codeName对应的公众号当前的access_token
返回数据:
成功: {"appId": #appId#, "token": #access_token#}
错误: {"error": #ErrorMessage#}
```
```http
/proxy-wechat-app/{codeName:string}/...

代理指定codeName对应的公众号微信接口, 自动添加access_token参数
```
```http
/accept-app-authorization/{codeName:string}

第三方平台在微信配置的授权事件接收URL
用于接收component_verify_ticket以及公众号对第三方平台进行授权、取消授权、更新授权的推送通知
返回数据: "success"
```
```http
/app-authorize-component-scan/{codeName:string}

第三方平台扫码授权入口页面, 跳转到微信的扫码授权页面
用于引导公众号和小程序管理员向第三方平台授权
跳转页面地址:
https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=#appId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/app-authorize-redirect/{codeName:string}#
```
```http
/app-authorize-component-link/{codeName:string}

第三方平台移动端链接授权入口页面, 跳转到微信的链接授权页面
用于引导公众号和小程序管理员向第三方平台授权
跳转页面地址:
https://mp.weixin.qq.com/safe/bindcomponent?action=bindcomponent&no_scan=1&component_appid=#appId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/app-authorize-redirect/{codeName:string}##wechat_redirect
```
```http
/app-authorize-redirect/{codeName:string}

第三方平台授权回调地址
跳转页面地址:
如果第三方平台配置了WECHAT_APP_THIRD_PLATFORM_CONFIG.REDIRECT_URL, 则跳转到此地址
```
```http
/query-wechat-app-authorizer-token/{codeName:string}/{authorizerAppId:string}

获取指定codeName对应的第三方平台所代理的authorizerAppId对应的公众号当前的authorizer_access_token
返回数据:
成功: {"appId": #appId#, "authorizerAppId": #authorizerAppId#, "token": #authorizer_access_token#}
错误: {"error": #ErrorMessage#}
```
```http
/query-wechat-corp-token/{codeName:string}

获取指定codeName对应的企业微信当前的access_token
返回数据:
成功: {"corpId": #corpId#, "token": #access_token#}
错误: {"error": #ErrorMessage#}
```
```http
/proxy-wechat-corp/{codeName:string}/...

代理指定codeName对应的企业微信接口, 自动添加access_token参数
```
```http
/accept-corp-authorization/{codeName:string}

企业第三方应用在微信配置的授权事件接收URL
用于接收suite_ticket以及企业微信对第三方应用进行授权、取消授权、更新授权的推送通知
返回数据: "success"
```
```http
/corp-authorize-component/{codeName:string}?state={state:string}

企业第三方应用授权入口页面, 跳转到微信的授权页面
用于引导企业微信管理员向第三方应用授权
跳转页面地址:
https://open.work.weixin.qq.com/3rdapp/install?suite_id=#suiteId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/corp-authorize-redirect/{codeName:string}#&state=#state#
```
```http
/corp-authorize-redirect/{codeName:string}

企业第三方应用授权回调地址
跳转页面地址:
如果第三方平台配置了WECHAT_CORP_THIRD_PLATFORM_CONFIG.REDIRECT_URL, 则跳转到此地址
```
```http
/query-wechat-corp-authorizer-token/{codeName:string}/{corpId:string}

获取指定codeName对应的企业第三方应用所代理的corpId对应的企业微信当前的access_token
返回数据:
成功: {"suiteId": #suiteId#, "corpId": #corpId#, "token": #access_token#}
错误: {"error": #ErrorMessage#}
```

#### Golang Kits

  [varys-go-driver](https://github.com/CharLemAznable/varys-go-driver)

#### Java Kits

  [varys-java-driver](https://github.com/CharLemAznable/varys-java-driver)
