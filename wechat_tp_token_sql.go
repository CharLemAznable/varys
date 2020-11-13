package main

const updateWechatTpTicketSQL = `
replace into wechat_app_third_platform_ticket
      (code_name
      ,app_id
      ,ticket)
select c.code_name
      ,c.app_id
      ,:Ticket
  from wechat_app_third_platform_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const queryWechatTpTicketSQL = `
select t.ticket as "Ticket"
  from wechat_app_third_platform_ticket t
 where t.code_name  = :CodeName
`

const queryWechatTpConfigSQL = `
select c.app_id         as "AppId"
      ,c.app_secret     as "AppSecret"
      ,c.token          as "Token"
      ,c.aes_key        as "AesKey"
      ,c.redirect_url   as "RedirectURL"
  from wechat_app_third_platform_config c
 where c.enabled        = 1
   and c.code_name      = :CodeName
`

const queryWechatTpTokenSQL = `
select t.app_id                         as "AppId"
      ,t.access_token                   as "AccessToken"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_app_third_platform_token t
 where t.code_name  = :CodeName
`

const createWechatTpTokenSQL = `
insert into wechat_app_third_platform_token
      (code_name
      ,app_id
      ,updated)
select c.code_name
      ,c.app_id
      ,0
  from wechat_app_third_platform_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const updateWechatTpTokenSQL = `
update wechat_app_third_platform_token
   set updated      = 0
 where code_name    = :CodeName
   and updated      = 1
   and expire_time  < now()
`

const uncompleteWechatTpTokenSQL = `
update wechat_app_third_platform_token
   set updated      = 1
 where code_name    = :CodeName
   and updated      = 0
`

const completeWechatTpTokenSQL = `
update wechat_app_third_platform_token
   set updated      = 1
      ,access_token = :AccessToken
      ,expire_time  = date_add(now(), interval :ExpiresIn second)
 where code_name    = :CodeName
   and updated      = 0
`
