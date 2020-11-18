package main

const queryFengniaoAppConfigSQL = `
select c.app_id     as "AppId"
      ,c.secret_key as "SecretKey"
  from fengniao_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const queryFengniaoAppTokenSQL = `
select t.app_id                         as "AppId"
      ,t.access_token                   as "AccessToken"
      ,t.updated                        as "Updated"
      ,unix_timestamp(t.expire_time)    as "ExpireTime"
  from fengniao_app_token t
 where t.code_name  = :CodeName
`

const createFengniaoAppTokenSQL = `
insert into fengniao_app_token
      (code_name
      ,app_id
      ,updated)
select c.code_name
      ,c.app_id
      ,0
  from fengniao_app_config c
 where c.enabled    = 1
   and c.code_name  = :CodeName
`

const updateFengniaoAppTokenSQL = `
update fengniao_app_token
   set updated      = 0
 where code_name    = :CodeName
   and updated      = 1
   and expire_time  < now()
`

const uncompleteFengniaoAppTokenSQL = `
update fengniao_app_token
   set updated      = 1
 where code_name    = :CodeName
   and updated      = 0
`

const completeFengniaoAppTokenSQL = `
update fengniao_app_token
   set updated      = 1
      ,access_token = :AccessToken
      ,expire_time  = from_unixtime(:ExpireTime)
 where code_name    = :CodeName
   and updated      = 0
`
