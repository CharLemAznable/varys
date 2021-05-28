package app

const queryConfigSQL = `
select c.dev_id             as "DevId"
      ,c.app_id             as "AppId"
      ,c.app_secret         as "AppSecret"
      ,c.callback_url       as "CallbackURL"
  from fengniao_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const createTokenSQL = `
replace into fengniao_app_token
      (code_name
      ,app_id
      ,merchant_id
      ,code
      ,access_token
      ,refresh_token
      ,expire_time
      ,re_expire_time)
select c.code_name
      ,c.app_id
      ,:MerchantId
      ,:Code
      ,:AccessToken
      ,:RefreshToken
      ,date_add(now(), interval :ExpireIn second)
      ,date_add(now(), interval :ReExpireIn second)
  from fengniao_app_config c
 where c.enabled      = 1
   and c.code_name    = :CodeName
`

const queryTokenSQL = `
select t.app_id                         as "AppId"
      ,t.merchant_id                    as "MerchantId"
      ,t.access_token                   as "AccessToken"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
      ,t.refresh_token                  as "RefreshToken"
      ,unix_timestamp(t.re_expire_time) as "ReExpireTime"
  from fengniao_app_token t
 where t.code_name    = :CodeName
   and t.merchant_id  = :MerchantId
`

const updateTokenSQL = `
update fengniao_app_token
   set access_token   = :AccessToken
      ,refresh_token  = :RefreshToken
      ,expire_time    = date_add(now(), interval :ExpireIn second)
      ,re_expire_time = date_add(now(), interval :ReExpireIn second)
 where code_name      = :CodeName
   and merchant_id    = :MerchantId
`
