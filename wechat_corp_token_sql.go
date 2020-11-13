package main

const queryWechatCorpConfigSQL = `
select c.corp_id        as "CorpId"
      ,c.corp_secret    as "CorpSecret"
  from wechat_corp_config c
 where c.enabled        = 1
   and c.code_name      = :CodeName
`

const queryWechatCorpTokenSQL = `
select t.corp_id                        as "CorpId"
      ,t.access_token                   as "AccessToken"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from wechat_corp_token t
 where t.code_name = :CodeName
`

const createWechatCorpTokenSQL = `
insert into wechat_corp_token
      (code_name
      ,corp_id
      ,access_token
      ,expire_time)
select c.code_name
      ,c.corp_id
      ,:AccessToken
      ,from_unixtime(:ExpireTime)
  from wechat_corp_config c
 where c.enabled   = 1
   and c.code_name = :CodeName
`

const updateWechatCorpTokenSQL = `
update wechat_corp_token
   set access_token = :AccessToken
      ,expire_time  = from_unixtime(:ExpireTime)
 where code_name    = :CodeName
   and expire_time  < now()
`
