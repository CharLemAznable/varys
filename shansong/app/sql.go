package app

const queryConfigSQL = `
select c.app_id             as "AppId"
      ,c.app_secret         as "AppSecret"
      ,c.redirect_url       as "RedirectURL"
      ,c.callback_url       as "CallbackURL"
  from shansong_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const createTokenSQL = `
replace into shansong_app_token
      (code_name
      ,app_id
      ,merchant_code
      ,code
      ,access_token
      ,updated
      ,expire_time
      ,refresh_token)
select c.code_name
      ,c.app_id
      ,:MerchantCode
      ,:Code
      ,:AccessToken
      ,1
      ,date_add(now(), interval :ExpireIn second)
      ,:RefreshToken
  from shansong_app_config c
 where c.enabled      = 1
   and c.code_name    = :CodeName
`

const queryTokenSQL = `
select t.app_id                         as "AppId"
      ,t.merchant_code                  as "MerchantCode"
      ,t.access_token                   as "AccessToken"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
      ,t.refresh_token                  as "RefreshToken"
  from shansong_app_token t
 where t.code_name      = :CodeName
   and t.merchant_code  = :MerchantCode
`

const updateTokenSQL = `
update shansong_app_token
   set updated        = 0
 where code_name      = :CodeName
   and merchant_code  = :MerchantCode
   and updated        = 1
   and expire_time    < now()
`

const uncompleteTokenSQL = `
update shansong_app_token
   set updated        = 1
 where code_name      = :CodeName
   and merchant_code  = :MerchantCode
   and updated        = 0
`

const completeTokenSQL = `
update shansong_app_token
   set updated        = 1
      ,access_token   = :AccessToken
      ,expire_time    = date_add(now(), interval :ExpiresIn second)
 where code_name      = :CodeName
   and merchant_code  = :MerchantCode
   and updated        = 0
`
