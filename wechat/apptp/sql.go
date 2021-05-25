package apptp

const updateTicketSQL = `
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

const queryTicketSQL = `
select t.ticket
  from wechat_app_third_platform_ticket t
 where t.code_name  = :CodeName
`

const queryConfigSQL = `
select c.app_id             as "AppId"
      ,c.app_secret         as "AppSecret"
      ,c.token              as "Token"
      ,c.aes_key            as "AesKey"
      ,c.redirect_url       as "RedirectURL"
      ,c.auth_forward_url   as "AuthForwardUrl"
      ,c.msg_forward_url    as "MsgForwardUrl"
  from wechat_app_third_platform_config c
 where c.enabled            = 1
   and c.code_name          = :CodeName
`

const queryTokenSQL = `
select t.app_id                         as "AppId"
      ,t.access_token                   as "AccessToken"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_app_third_platform_token t
 where t.code_name  = :CodeName
`

const createTokenSQL = `
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

const updateTokenSQL = `
update wechat_app_third_platform_token
   set updated      = 0
 where code_name    = :CodeName
   and updated      = 1
   and expire_time  < now()
`

const uncompleteTokenSQL = `
update wechat_app_third_platform_token
   set updated      = 1
 where code_name    = :CodeName
   and updated      = 0
`

const completeTokenSQL = `
update wechat_app_third_platform_token
   set updated      = 1
      ,access_token = :AccessToken
      ,expire_time  = date_add(now(), interval :ExpiresIn second)
 where code_name    = :CodeName
   and updated      = 0
`
