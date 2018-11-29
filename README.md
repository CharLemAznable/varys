## varys

  AccessToken 中控服务器
  
  统一DB存储AccessToken, 支持分布式部署服务访问和更新.
  
  #### 数据库
  
  建表SQL:
   
  [varys.sql](https://github.com/CharLemAznable/varys/blob/master/varys.sql)
  
  #### 本地缓存
  
  包含微信公众号配置缓存和access_token缓存, 其中:
  
  1) 公众号配置缓存默认1小时
  2) access_token缓存默认5分钟, 当access_token即将过期并被其他分布式节点更新时缓存1分钟
  
  包含微信第三方平台配置缓存和报文解密器缓存, 其中
  
  1) 第三方平台配置缓存默认1小时
  2) 第三方平台报文解密器缓存默认1小时
  
  包含微信第三方平台component_access_token/pre_auth_code/authorizer_access_token缓存, 其中
  
  1) component_access_token缓存默认5分钟, 当component_access_token即将过期并被其他分布式节点更新时缓存1分钟
  2) pre_auth_code缓存默认3分钟, 当pre_auth_code即将过期并被其他分布式节点更新时缓存1分钟
  3) authorizer_access_token缓存默认5分钟, 当authorizer_access_token即将过期并被其他分布式节点更新时缓存1分钟
  
  [cache.go](https://github.com/CharLemAznable/varys/blob/master/cache.go)
  
  #### 访问路径
  
  默认服务地址:
  ```
http://localhost:4236/varys
  ```
  ```
/query-wechat-api-token/{codeName:string}

获取指定codeName对应的公众号当前的access_token
返回数据:
成功: {"appId": #appId#, "token": #access_token#}
错误: {"appId": #appId#, "error": #ErrorMessage#}
  ```
  ```
/query-wechat-authorizer-token/{codeName:string}/{authorizerAppId:string}

获取指定codeName对应的第三方平台所代理的authorizerAppId对应的公众号当前的authorizer_access_token
返回数据:
成功: {"appId": #appId#, "authorizerAppId": #authorizerAppId#, "token": #authorizer_access_token#}
错误: {"appId": #appId#, "authorizerAppId": #authorizerAppId#, "error": #ErrorMessage#}
  ```
  ```
/accept-authorization/{codeName:string}

第三方平台在微信配置的授权事件接收URL
用于接收component_verify_ticket以及公众号对第三方平台进行授权、取消授权、更新授权的推送通知
返回数据: "success"
  ```
  ```
/authorize-component-scan/{codeName:string}

第三方平台扫码授权入口页面, 跳转到微信的扫码授权页面
用于引导公众号和小程序管理员向第三方平台授权
跳转页面地址:
https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=#appId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/authorize-redirect/{codeName:string}#
  ```
  ```
/authorize-component-link/{codeName:string}

第三方平台移动端链接授权入口页面, 跳转到微信的链接授权页面
用于引导公众号和小程序管理员向第三方平台授权
跳转页面地址:
https://mp.weixin.qq.com/safe/bindcomponent?action=bindcomponent&no_scan=1&component_appid=#appId#&pre_auth_code=#pre_auth_code#&redirect_uri=#url_to_/authorize-redirect/{codeName:string}##wechat_redirect
  ```
  ```
/authorize-redirect/{codeName:string}

第三方平台授权回调地址
跳转页面地址:
如果第三方平台配置了WECHAT_THIRD_PLATFORM_CONFIG.REDIRECT_URL, 则跳转到此地址
  ```
  
  [varys.go](https://github.com/CharLemAznable/varys/blob/master/varys.go)
