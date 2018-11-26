package varys

// global configuration

const queryConfigurationSQL = `
SELECT C.CONFIG_NAME ,C.CONFIG_VALUE
  FROM APP_CONFIG C
 WHERE C.ENABLED = 1
`

// wechat API access_token

const queryWechatAPITokenConfigSQL = `
SELECT C.APP_ID ,C.APP_SECRET
  FROM WECHAT_API_TOKEN_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const queryWechatAPITokenSQL = `
SELECT T.APP_ID ,T.ACCESS_TOKEN ,T.UPDATED
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_API_TOKEN T
 WHERE T.CODE_NAME = ?
`

const createWechatAPITokenUpdating = `
INSERT INTO WECHAT_API_TOKEN
      (CODE_NAME    ,APP_ID     ,UPDATED)
SELECT C.CODE_NAME  ,C.APP_ID   ,0
  FROM WECHAT_API_TOKEN_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const updateWechatAPITokenUpdating = `
UPDATE WECHAT_API_TOKEN
   SET UPDATED   = 0
 WHERE CODE_NAME = ?
   AND UPDATED   = 1
`

const uncompleteWechatAPITokenSQL = `
UPDATE WECHAT_API_TOKEN
   SET UPDATED   = 1
 WHERE CODE_NAME = ?
   AND UPDATED   = 0
`

const completeWechatAPITokenSQL = `
UPDATE WECHAT_API_TOKEN
   SET UPDATED      = 1
      ,ACCESS_TOKEN = ?
      ,EXPIRE_TIME  = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE CODE_NAME    = ?
   AND UPDATED      = 0
`

// wechat third platform authorizer access_token

const queryWechatThirdPlatformConfigSQL = `
SELECT C.APP_ID ,C.APP_SECRET ,C.TOKEN ,C.AES_KEY ,C.REDIRECT_URL
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

// component_verify_ticket

const replaceWechatThirdPlatformTicketSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_TICKET
      (CODE_NAME    ,APP_ID     ,TICKET)
SELECT C.CODE_NAME  ,C.APP_ID   ,?
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const queryWechatThirdPlatformTicketSQL = `
SELECT T.TICKET
  FROM WECHAT_THIRD_PLATFORM_TICKET T
 WHERE T.CODE_NAME = ?
`

// component_access_token

const queryWechatThirdPlatformTokenSQL = `
SELECT T.APP_ID ,T.COMPONENT_ACCESS_TOKEN ,T.UPDATED 
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_THIRD_PLATFORM_TOKEN T
 WHERE T.CODE_NAME = ?
`

const createWechatThirdPlatformTokenUpdating = `
INSERT INTO WECHAT_THIRD_PLATFORM_TOKEN
      (CODE_NAME    ,APP_ID     ,UPDATED)
SELECT C.CODE_NAME  ,C.APP_ID   ,0
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const updateWechatThirdPlatformTokenUpdating = `
UPDATE WECHAT_THIRD_PLATFORM_TOKEN
   SET UPDATED   = 0
 WHERE CODE_NAME = ?
   AND UPDATED   = 1
`

const uncompleteWechatThirdPlatformTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_TOKEN
   SET UPDATED   = 1
 WHERE CODE_NAME = ?
   AND UPDATED   = 0
`

const completeWechatThirdPlatformTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_TOKEN
   SET UPDATED                = 1
      ,COMPONENT_ACCESS_TOKEN = ?
      ,EXPIRE_TIME            = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE CODE_NAME              = ?
   AND UPDATED                = 0
`

// pre_auth_code

const queryWechatThirdPlatformPreAuthCodeSQL = `
SELECT T.APP_ID ,T.PRE_AUTH_CODE ,T.UPDATED 
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE T
 WHERE T.CODE_NAME = ?
`

const createWechatThirdPlatformPreAuthCodeUpdating = `
INSERT INTO WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
      (CODE_NAME    ,APP_ID     ,UPDATED)
SELECT C.CODE_NAME  ,C.APP_ID   ,0
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const updateWechatThirdPlatformPreAuthCodeUpdating = `
UPDATE WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
   SET UPDATED   = 0
 WHERE CODE_NAME = ?
   AND UPDATED   = 1
`

const uncompleteWechatThirdPlatformPreAuthCodeSQL = `
UPDATE WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
   SET UPDATED   = 1
 WHERE CODE_NAME = ?
   AND UPDATED   = 0
`

const completeWechatThirdPlatformPreAuthCodeSQL = `
UPDATE WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
   SET UPDATED       = 1
      ,PRE_AUTH_CODE = ?
      ,EXPIRE_TIME   = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE CODE_NAME     = ?
   AND UPDATED       = 0
`

// authorization_code

const enableWechatThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_AUTHORIZER
      (CODE_NAME            ,APP_ID                 ,AUTHORIZER_APP_ID
      ,AUTHORIZATION_STATE  ,AUTHORIZATION_CODE     ,PRE_AUTH_CODE)
SELECT C.CODE_NAME          ,C.APP_ID               ,?
      ,1                    ,?                      ,?
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const disableWechatThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_AUTHORIZER
      (CODE_NAME            ,APP_ID
      ,AUTHORIZER_APP_ID    ,AUTHORIZATION_STATE)
SELECT C.CODE_NAME          ,C.APP_ID
      ,?                    ,0
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

// authorizer_access_token

const queryWechatThirdPlatformAuthorizerTokenSQL = `
SELECT T.APP_ID ,T.AUTHORIZER_APP_ID
      ,T.AUTHORIZER_ACCESS_TOKEN
      ,T.AUTHORIZER_REFRESH_TOKEN
      ,T.UPDATED ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN T
      ,WECHAT_THIRD_PLATFORM_AUTHORIZER A
 WHERE T.CODE_NAME           = ?
   AND T.AUTHORIZER_APP_ID   = ?
   AND A.CODE_NAME           = T.CODE_NAME
   AND A.AUTHORIZER_APP_ID   = T.AUTHORIZER_APP_ID
   AND A.AUTHORIZATION_STATE = 1
`

const updateWechatThirdPlatformAuthorizerTokenUpdating = `
UPDATE WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED           = 0
 WHERE CODE_NAME         = ?
   AND AUTHORIZER_APP_ID = ?
   AND UPDATED           = 1
`

const uncompleteWechatThirdPlatformAuthorizerTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED           = 1
 WHERE CODE_NAME         = ?
   AND AUTHORIZER_APP_ID = ?
   AND UPDATED           = 0
`

const completeWechatThirdPlatformAuthorizerTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED                  = 1
      ,AUTHORIZER_ACCESS_TOKEN  = ?
      ,AUTHORIZER_REFRESH_TOKEN = ?
      ,EXPIRE_TIME              = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE CODE_NAME                = ?
   AND AUTHORIZER_APP_ID        = ?
   AND UPDATED                  = 0
`

const createWechatThirdPlatformAuthorizerTokenUpdating = `
INSERT INTO WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
      (CODE_NAME    ,APP_ID     ,AUTHORIZER_APP_ID      ,UPDATED)
SELECT A.CODE_NAME  ,A.APP_ID   ,A.AUTHORIZER_APP_ID    ,0
  FROM WECHAT_THIRD_PLATFORM_AUTHORIZER A
 WHERE A.AUTHORIZATION_STATE = 1
   AND A.CODE_NAME           = ?
   AND A.AUTHORIZER_APP_ID   = ?
`
