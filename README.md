## varys

  微信API AccessToken 中控服务器
  
  统一DB存储微信AccessToken, 支持分布式部署服务访问和更新.
  
  #### 数据库
  
  建表SQL:
   
  [varys.sql](https://github.com/CharLemAznable/varys/blob/master/varys.sql)
  
  #### 本地缓存
  
  包含微信公众号配置缓存和AccessToken缓存, 其中:
  
  1) 公众号配置缓存默认1小时
  2) AccessToken缓存默认5分钟, 当AccessToken即将过期并被其他分布式节点更新时缓存1分钟
  
  [cache.go](https://github.com/CharLemAznable/varys/blob/master/cache.go)
  
  #### 访问路径
  
  默认服务地址:
  ```
http://localhost:4236/varys
  ```
  ```
/query-wechat-api-token/{appId:string}

获取指定appId对应的公众号当前的AccessToken
返回数据:
成功: {"appId": #appId#, "token": #AccessToken#}
错误: {"appId": #appId#, "error": #ErrorMessage#}
  ```
  
  [main.go](https://github.com/CharLemAznable/varys/blob/master/main.go)
