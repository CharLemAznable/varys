package main

const queryWechatAppConfigSQL = `
select c.app_id     as "AppId"
      ,c.app_secret as "AppSecret"
  from wechat_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const queryWechatAppTokenSQL = `
select t.app_id                         as "AppId"
      ,t.access_token                   as "AccessToken"
      ,t.jsapi_ticket                   as "JsapiTicket"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_app_token t
 where t.code_name  = :CodeName
`

const createWechatAppTokenSQL = `
insert into wechat_app_token
      (code_name
      ,app_id
      ,updated)
select c.code_name
      ,c.app_id
      ,0
  from wechat_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const updateWechatAppTokenSQL = `
update wechat_app_token
   set updated      = 0
 where code_name    = :CodeName
   and updated      = 1
   and expire_time  < now()
`

const uncompleteWechatAppTokenSQL = `
update wechat_app_token
   set updated      = 1
 where code_name    = :CodeName
   and updated      = 0
`

const completeWechatAppTokenSQL = `
update wechat_app_token
   set updated      = 1
      ,access_token = :AccessToken
      ,jsapi_ticket = :JsapiTicket
      ,expire_time  = date_add(now(), interval :ExpiresIn second)
 where code_name    = :CodeName
   and updated      = 0
`
