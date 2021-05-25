package apptp

const enableAuthSQL = `
replace into wechat_app_third_platform_authorizer
      (code_name
      ,app_id
      ,authorizer_app_id
      ,authorization_state
      ,authorization_code
      ,pre_auth_code)
select c.code_name
      ,c.app_id
      ,:AuthorizerAppId
      ,1
      ,:AuthorizationCode
      ,:PreAuthCode
  from wechat_app_third_platform_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const disableAuthSQL = `
replace into wechat_app_third_platform_authorizer
      (code_name
      ,app_id
      ,authorizer_app_id
      ,authorization_state)
select c.code_name
      ,c.app_id
      ,:AuthorizerAppId
      ,0
  from wechat_app_third_platform_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const createAuthTokenSQL = `
insert into wechat_app_third_platform_authorizer_token
      (code_name
      ,app_id
      ,authorizer_app_id
      ,updated)
select a.code_name
      ,a.app_id
      ,a.authorizer_app_id
      ,0
  from wechat_app_third_platform_authorizer a
 where a.authorization_state = 1
   and a.code_name           = :CodeName
   and a.authorizer_app_id   = :AuthorizerAppId
`

const updateAuthTokenForceSQL = `
update wechat_app_third_platform_authorizer_token
   set updated              = 0
 where code_name            = :CodeName
   and authorizer_app_id    = :AuthorizerAppId
   and updated              = 1
`

const uncompleteAuthTokenSQL = `
update wechat_app_third_platform_authorizer_token
   set updated              = 1
 where code_name            = :CodeName
   and authorizer_app_id    = :AuthorizerAppId
   and updated              = 0
`

const completeAuthTokenSQL = `
update wechat_app_third_platform_authorizer_token
   set updated                  = 1
      ,authorizer_access_token  = :AuthorizerAccessToken
      ,authorizer_refresh_token = :AuthorizerRefreshToken
      ,authorizer_jsapi_ticket  = :AuthorizerJsapiTicket
      ,expire_time              = date_add(now(), interval :ExpiresIn second)
 where code_name                = :CodeName
   and authorizer_app_id        = :AuthorizerAppId
   and updated                  = 0
`

const queryAuthTokenSQL = `
select t.app_id                         as "AppId"
      ,t.authorizer_app_id              as "AuthorizerAppId"
      ,t.authorizer_access_token        as "AuthorizerAccessToken"
      ,t.authorizer_refresh_token       as "AuthorizerRefreshToken"
      ,t.authorizer_jsapi_ticket        as "AuthorizerJsapiTicket"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_app_third_platform_authorizer_token t
      ,wechat_app_third_platform_authorizer a
 where t.code_name                      = :CodeName
   and t.authorizer_app_id              = :AuthorizerAppId
   and a.code_name                      = t.code_name
   and a.authorizer_app_id              = t.authorizer_app_id
   and a.authorization_state            = 1
`

const updateAuthTokenSQL = `
update wechat_app_third_platform_authorizer_token
   set updated              = 0
 where code_name            = :CodeName
   and authorizer_app_id    = :AuthorizerAppId
   and updated              = 1
   and expire_time          < now()
`
