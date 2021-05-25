package corptp

const updateTicketSQL = `
replace into wechat_corp_third_platform_ticket
      (code_name
      ,suite_id
      ,ticket)
select c.code_name
      ,c.suite_id
      ,:Ticket
  from wechat_corp_third_platform_config c
 where c.enabled   = 1
   and c.code_name = :CodeName
`

const queryTicketSQL = `
select t.ticket
  from wechat_corp_third_platform_ticket t
 where t.code_name = :CodeName
`

const queryConfigSQL = `
select c.suite_id       as "SuiteId"
      ,c.suite_secret   as "SuiteSecret"
      ,c.token          as "Token"
      ,c.aes_key        as "AesKey"
      ,c.redirect_url   as "RedirectURL"
  from wechat_corp_third_platform_config c
 where c.enabled        = 1
   and c.code_name      = :CodeName
`

const queryTokenSQL = `
select t.suite_id                       as "SuiteId"
      ,t.access_token                   as "AccessToken"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_corp_third_platform_token t
 where t.code_name = :CodeName
`

const createTokenSQL = `
insert into wechat_corp_third_platform_token
      (code_name
      ,suite_id
      ,access_token
      ,expire_time)
select c.code_name
      ,c.suite_id
      ,:AccessToken
      ,from_unixtime(:ExpireTime)
  from wechat_corp_third_platform_config c
 where c.enabled   = 1
   and c.code_name = :CodeName
`

const updateTokenSQL = `
update wechat_corp_third_platform_token
   set access_token = :AccessToken
      ,expire_time  = from_unixtime(:ExpireTime)
 where code_name    = :CodeName
   and expire_time  < now()
`
