## varys

  AccessToken 中控服务器

  统一DB存储AccessToken, 支持分布式部署服务访问和更新.

  #### 数据库

  建表SQL:

  [varys.sql](https://github.com/CharLemAznable/varys/blob/master/varys.sql)

  #### 本地缓存

  包含微信公众号配置缓存和```access_token```缓存, 其中:

  1) 公众号配置缓存默认1小时
  2) access_token缓存默认5分钟, 当access_token即将过期并被其他分布式节点更新时缓存1分钟

  包含微信第三方平台配置缓存和报文解密器缓存, 其中

  1) 第三方平台配置缓存默认1小时
  2) 第三方平台报文解密器缓存默认1小时

  包含微信第三方平台```component_access_token```/```authorizer_access_token```缓存, 其中

  1) component_access_token缓存默认5分钟, 当component_access_token即将过期并被其他分布式节点更新时缓存1分钟
  2) authorizer_access_token缓存默认5分钟, 当authorizer_access_token即将过期并被其他分布式节点更新时缓存1分钟
  
  包含企业微信配置缓存和```access_token```缓存, 其中:

  1) 企业微信配置缓存默认1小时
  2) access_token缓存最大5分钟, 当access_token即将过期时, 缓存时间最大至其有效期结束

  [cache.go](https://github.com/CharLemAznable/varys/blob/master/cache.go)

  #### 访问路径

  默认服务地址:
```http
http://localhost:4236/varys
```
```http
/query-wechat-app-token/{codeName:string}

获取指定codeName对应的公众号当前的access_token
返回数据:
成功: {"appId": #appId#, "token": #access_token#}
错误: {"error": #ErrorMessage#}
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

  [varys.go](https://github.com/CharLemAznable/varys/blob/master/varys.go)

  #### 打包部署

  新建Go File:
```go
package main

import "github.com/CharLemAznable/varys"

func main() {
    varys.Default().Run()
    // 或自定义路径和端口
    // varys.NewVarys("/varys", ":4236").Run()
}
```
  命令行```build```: (Linux AMD64主机环境)
```bash
$ env GOOS=linux GOARCH=amd64 go build -o varys.linux.bin
```
  同路径下新建日志配置文件```logback.xml```:
```xml
<logging>
    <filter enabled="true">
        <tag>file</tag>
        <type>file</type>
        <level>TRACE</level>
        <property name="filename">varys.log</property>
        <property name="format">[%D %T] [%L] (%S) %M</property>
        <property name="rotate">false</property>
        <property name="maxsize">0M</property>
        <property name="maxlines">0K</property>
        <property name="daily">false</property>
    </filter>
</logging>
```
  同路径下新建数据库连接配置文件```gql.yaml```:
```yaml
Default:
  DriverName:       mysql
  DataSourceName:   username:password@tcp(host:port)/dbname?charset=utf8
  MaxOpenConns:     50
  MaxIdleConns:     1
  ConnMaxLifetime:  60
```
  启动```varys```服务:
```bash
$ nohup ./varys.linux.bin &
```
