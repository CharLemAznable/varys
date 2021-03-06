package main

const queryToutiaoAppConfigSQL = `
select c.app_id     as "AppId"
      ,c.app_secret as "AppSecret"
  from toutiao_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const queryToutiaoAppTokenSQL = `
select t.app_id                         as "AppId"
      ,t.access_token                   as "AccessToken"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from toutiao_app_token t
 where t.code_name  = :CodeName
`

const createToutiaoAppTokenSQL = `
insert into toutiao_app_token
      (code_name
      ,app_id
      ,updated)
select c.code_name
      ,c.app_id
      ,0
  from toutiao_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const updateToutiaoAppTokenSQL = `
update toutiao_app_token
   set updated      = 0
 where code_name    = :CodeName
   and updated      = 1
   and expire_time  < now()
`

const uncompleteToutiaoAppTokenSQL = `
update toutiao_app_token
   set updated      = 1
 where code_name    = :CodeName
   and updated      = 0
`

const completeToutiaoAppTokenSQL = `
update toutiao_app_token
   set updated      = 1
      ,access_token = :AccessToken
      ,expire_time  = date_add(now(), interval :ExpiresIn second)
 where code_name    = :CodeName
   and updated      = 0
`
