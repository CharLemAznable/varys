## varys

[![Build Status](https://travis-ci.org/CharLemAznable/varys.svg?branch=master)](https://travis-ci.org/CharLemAznable/varys)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/CharLemAznable/varys)
[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)
![GitHub code size](https://img.shields.io/github/languages/code-size/CharLemAznable/varys)

AccessToken 中控服务器

统一DB存储AccessToken, 支持分布式部署服务访问和更新.

#### 配置文件

```config.toml``` [示例](https://github.com/CharLemAznable/varys/blob/master/config.toml)

```toml
Port = 4236
ContextPath = ""
LogLevel = "info"
ClusterNodeAddresses = [ "http://localhost:4236" ]

DriverName = "mysql"
DataSourceName = "admin:test123@tcp(127.0.0.1:3306)/rock?charset=utf8"
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

包含微信公众号/小程序配置缓存和```access_token```缓存, 其中:

1) 公众号/小程序配置缓存默认1小时
2) access_token缓存默认5分钟, 当access_token即将过期并被其他分布式节点更新时缓存1分钟

  [wechat_app_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/wechat/app/cache.go)

包含微信第三方平台配置缓存和报文解密器缓存, 其中

1) 第三方平台配置缓存默认1小时
2) 第三方平台报文解密器缓存默认1小时

包含微信第三方平台```component_access_token```/授权用户```authorizer_access_token```缓存, 其中

1) component_access_token缓存默认5分钟, 当component_access_token即将过期并被其他分布式节点更新时缓存1分钟
2) authorizer_access_token缓存默认5分钟, 当authorizer_access_token即将过期并被其他分布式节点更新时缓存1分钟

  [wechat_tp_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/wechat/apptp/cache.go)
  [wechat_tp_auth_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/wechat/apptp/auth_cache.go)

包含企业微信配置缓存和```access_token```缓存, 其中:

1) 企业微信配置缓存默认1小时
2) access_token缓存最大5分钟, 当access_token即将过期时, 缓存时间最大至其有效期结束

  [wechat_corp_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/wechat/corp/cache.go)

包含企业微信第三方应用配置缓存和报文解密器缓存, 其中

1) 企业微信第三方应用配置缓存默认1小时
2) 企业微信第三方应用报文解密器缓存默认1小时

包含企业微信第三方应用```suite_access_token```/```access_token```缓存, 其中

1) suite_access_token缓存最大5分钟, 当suite_access_token即将过期时, 缓存时间最大至其有效期结束
2) access_token缓存最大5分钟, 当access_token即将过期时, 缓存时间最大至其有效期结束

  [wechat_corp_tp_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/wechat/corptp/cache.go)
  [wechat_corp_tp_auth_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/wechat/corptp/auth_cache.go)

包含字节小程序配置缓存和```access_token```缓存, 其中:

1) 小程序配置缓存默认1小时
2) access_token缓存默认5分钟, 当access_token即将过期并被其他分布式节点更新时缓存1分钟

  [toutiao_app_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/toutiao/app/cache.go)

包含蜂鸟应用配置缓存和授权商户```access_token```缓存，其中:

1) 应用配置缓存默认1小时
2) 商户access_token缓存默认8分钟

  [fengniao_app_token_cache.go](https://github.com/CharLemAznable/varys/blob/master/fengniao/app/cache.go)

#### 访问路径

默认服务地址:
```http
http://localhost:4236
```

微信公众号/小程序:
```http
/query-wechat-app-token/{codeName:string}

获取指定codeName对应的公众号/小程序当前的access_token和jsapi_ticket
返回数据:
成功: {"appId": #appId#, "token": #access_token#, "ticket": #jsapi_ticket#}
错误: {"error": #ErrorMessage#}
```
```http
/proxy-wechat-app/{codeName:string}/...

代理指定codeName对应的公众号/小程序微信接口, 自动添加access_token参数
```
```http
/proxy-wechat-app-mp-login/{codeName:string}?js_code=JSCODE

代理指定codeName对应的小程序登录凭证校验

通过 wx.login 接口获得临时登录凭证 code 后调用此接口，获取微信提供的用户身份标识
```
详见: [微信开放文档 auth.code2Session](https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html)
```http
/query-wechat-app-js-config/{codeName:string}?url=URL

获取指定codeName对应的公众号使用JS-SDK的注入配置信息
返回数据:
成功: {"appId": #appId#, "timestamp": #timestamp#, "nonceStr": #nonceStr#, "signature": #signature#}
错误: {"error": #ErrorMessage#}
```
详见: [微信开放文档 JS-SDK说明文档 通过config接口注入权限验证配置](https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html#4)

微信第三方平台:
```http
/accept-wechat-tp-info/{codeName:string}

第三方平台在微信配置的授权事件接收URL
用于接收component_verify_ticket以及公众号对第三方平台进行授权、取消授权、更新授权的推送通知，以及快速创建小程序的审核结果通知
返回数据: "success"
```
```http
/accept-wechat-tp-msg/{codeName:string}

第三方平台在微信配置的消息与事件接收URL
用于代收用户发送给公众号/小程序的消息，以及小程序改名的审核结果通知
返回数据: "success"
```
```http
/query-wechat-tp-token/{codeName:string}

获取指定codeName对应的第三方平台当前的component_access_token
返回数据:
成功: {"appId": #appId#, "token": #component_access_token#}
错误: {"error": #ErrorMessage#}
```
```http
/proxy-wechat-tp/{codeName:string}/...

代理指定codeName对应的第三方平台微信接口, 自动添加component_access_token参数
```

微信第三方平台授权方:
```http
/wechat-tp-authorize-scan/{codeName:string}

第三方平台扫码授权入口页面, 跳转到微信的扫码授权页面
用于引导公众号和小程序管理员向第三方平台授权
跳转页面地址:
https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=#appId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/app-authorize-redirect/{codeName:string}#
```
```http
/wechat-tp-authorize-link/{codeName:string}

第三方平台移动端链接授权入口页面, 跳转到微信的链接授权页面
用于引导公众号和小程序管理员向第三方平台授权
跳转页面地址:
https://mp.weixin.qq.com/safe/bindcomponent?action=bindcomponent&no_scan=1&component_appid=#appId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/app-authorize-redirect/{codeName:string}##wechat_redirect
```
```http
/wechat-tp-authorize-redirect/{codeName:string}

第三方平台授权回调地址
跳转页面地址:
如果第三方平台配置了WECHAT_APP_THIRD_PLATFORM_CONFIG.REDIRECT_URL, 则跳转到此地址
```
```http
/query-wechat-tp-auth-token/{codeName:string}/{authorizerAppId:string}

获取指定codeName对应的第三方平台所代理的authorizerAppId对应的公众号当前的authorizer_access_token和jsapi_ticket
返回数据:
成功: {"appId": #appId#, "authorizerAppId": #authorizerAppId#, "token": #authorizer_access_token#, "ticket": #jsapi_ticket#}
错误: {"error": #ErrorMessage#}
```
```http
/proxy-wechat-tp-auth/{codeName:string}/{authorizerAppId:string}/...

代理指定codeName对应的第三方平台所代理的authorizerAppId对应的公众号/小程序微信接口, 自动添加access_token参数
```
```http
/proxy-wechat-tp-auth-mp-login/{codeName:string}/{authorizerAppId:string}?js_code=JSCODE

代理指定codeName对应的第三方平台所代理的authorizerAppId对应的小程序登录凭证校验

通过 wx.login 接口获得临时登录凭证 code 后调用此接口，获取微信提供的用户身份标识
```
详见: [微信开放文档 第三方平台代小程序实现业务 微信登录](https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/Mini_Programs/WeChat_login.html)
```http
/query-wechat-tp-auth-js-config/{codeName:string}/{authorizerAppId:string}?url=URL

获取指定codeName对应的第三方平台所代理的authorizerAppId对应的公众号使用JS-SDK的注入配置信息
返回数据:
成功: {"appId": #appId#, "timestamp": #timestamp#, "nonceStr": #nonceStr#, "signature": #signature#}
错误: {"error": #ErrorMessage#}
```
详见: [微信开放文档 JS-SDK说明文档 通过config接口注入权限验证配置](https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html#4)
详见: [微信开放文档 第三方平台代公众号实现业务 代公众号使用JS-SDK说明](https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/Official_Accounts/js_sdk_instructions.html)

企业微信:
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

企业微信第三方平台:
```http
/accept-wechat-corp-tp-info/{codeName:string}

企业第三方应用在微信配置的授权事件接收URL
用于接收suite_ticket以及企业微信对第三方应用进行授权、取消授权、更新授权的推送通知
返回数据: "success"
```

企业微信第三方平台授权方:
```http
/wechat-corp-tp-authorize-component/{codeName:string}?state={state:string}

企业第三方应用授权入口页面, 跳转到微信的授权页面
用于引导企业微信管理员向第三方应用授权
跳转页面地址:
https://open.work.weixin.qq.com/3rdapp/install?suite_id=#suiteId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/corp-authorize-redirect/{codeName:string}#&state=#state#
```
```http
/wechat-corp-tp-authorize-redirect/{codeName:string}

企业第三方应用授权回调地址
跳转页面地址:
如果第三方平台配置了WECHAT_CORP_THIRD_PLATFORM_CONFIG.REDIRECT_URL, 则跳转到此地址
```
```http
/query-wechat-corp-tp-auth-token/{codeName:string}/{corpId:string}

获取指定codeName对应的企业第三方应用所代理的corpId对应的企业微信当前的access_token
返回数据:
成功: {"suiteId": #suiteId#, "corpId": #corpId#, "token": #access_token#}
错误: {"error": #ErrorMessage#}
```

头条APP:
```http
/query-toutiao-app-token/{codeName:string}

获取指定codeName对应的字节小程序当前的access_token
返回数据:
成功: {"appId": #appId#, "token": #access_token#}
错误: {"error": #ErrorMessage#}
```

蜂鸟:
```http
//fengniao-app-auth-callback/{codeName:string}

配置蜂鸟商户授权回调地址: 开发者中心 -> 应用管理 -> 查看应用详情
```
```http
/query-fengniao-app-token/{codeName:string}/{merchantId:string}

获取指定codeName对应的蜂鸟应用获取授权的商户当前的access_token
返回数据:
成功: {"appId": #appId#, "merchant_id": #merchantId#, "access_token": #accessToken#}
错误: {"error": #ErrorMessage#}
```
```http
/proxy-fengniao-app/{codeName:string}/{merchantId:string}/...

代理指定codeName对应的蜂鸟应用获取授权的商户接口, 包装原请求体为business_data并签名
```
```http
/fengniao-app-callback/{codeName:string}

配置蜂鸟应用消息推送回调地址: 开发者中心 -> 应用管理 -> 查看应用详情 -> 编辑
验证蜂鸟回调消息签名, 并将回调请求中的business_data提取转发到蜂鸟应用配置fengniao_app_config.callback_url地址(POST JSON)
```

#### Golang Kits

  [varys-go-driver](https://github.com/CharLemAznable/varys-go-driver)

#### Java Kits

  [varys-java-driver](https://github.com/CharLemAznable/varys-java-driver)
