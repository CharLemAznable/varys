package corptp

const enableAuthSQL = `
replace into wechat_corp_third_platform_authorizer
      (code_name
      ,suite_id
      ,corp_id
      ,state
      ,permanent_code)
select c.code_name
      ,c.suite_id
      ,:CorpId
      ,1
      ,:PermanentCode
  from wechat_corp_third_platform_config c
 where c.enabled   = 1
   and c.code_name = :CodeName
`

const disableAuthSQL = `
replace into wechat_corp_third_platform_authorizer
      (code_name
      ,suite_id
      ,corp_id
      ,state)
select c.code_name
      ,c.suite_id
      ,:CorpId
      ,0
  from wechat_corp_third_platform_config c
 where c.enabled   = 1
   and c.code_name = :CodeName
`

const queryPermanentCodeSQL = `
select a.suite_id       as "SuiteId"
      ,a.corp_id        as "CorpId"
      ,a.permanent_code as "PermanentCode"
  from wechat_corp_third_platform_authorizer a
 where a.state      = 1
   and a.code_name  = :CodeName
   and a.corp_id    = :CorpId
`

const createAuthTokenSQL = `
insert into wechat_corp_third_platform_corp_token
      (code_name
      ,suite_id
      ,corp_id
      ,corp_access_token
      ,expire_time)
select c.code_name
      ,c.suite_id
      ,:CorpId
      ,:AccessToken
      ,from_unixtime(:ExpireTime)
  from wechat_corp_third_platform_config c
 where c.enabled   = 1
   and c.code_name = :CodeName
`

const updateAuthTokenSQL = `
update wechat_corp_third_platform_corp_token
   set corp_access_token = :AccessToken
      ,expire_time       = from_unixtime(:ExpireTime)
 where code_name         = :CodeName
   and corp_id           = :CorpId
   and expire_time       < now()
`

const queryAuthTokenSQL = `
select t.suite_id                       as "SuiteId"
      ,t.corp_id                        as "CorpId"
      ,t.corp_access_token              as "CorpAccessToken"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_corp_third_platform_corp_token t
      ,wechat_corp_third_platform_authorizer a
 where t.code_name  = :CodeName
   and t.corp_id    = :CorpId
   and a.state      = 1
   and a.code_name  = t.code_name
   and a.corp_id    = t.corp_id
`
