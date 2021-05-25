package app

const queryConfigSQL = `
select c.app_id     as "AppId"
      ,c.app_secret as "AppSecret"
  from wechat_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const queryTokenSQL = `
select t.app_id                         as "AppId"
      ,t.access_token                   as "AccessToken"
      ,t.jsapi_ticket                   as "JsapiTicket"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_app_token t
 where t.code_name  = :CodeName
`

const createTokenSQL = `
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

const updateTokenSQL = `
update wechat_app_token
   set updated      = 0
 where code_name    = :CodeName
   and updated      = 1
   and expire_time  < now()
`

const uncompleteTokenSQL = `
update wechat_app_token
   set updated      = 1
 where code_name    = :CodeName
   and updated      = 0
`

const completeTokenSQL = `
update wechat_app_token
   set updated      = 1
      ,access_token = :AccessToken
      ,jsapi_ticket = :JsapiTicket
      ,expire_time  = date_add(now(), interval :ExpiresIn second)
 where code_name    = :CodeName
   and updated      = 0
`
