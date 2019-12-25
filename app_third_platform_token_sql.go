package main

const queryWechatAppThirdPlatformConfigSQL = `
SELECT C.APP_ID ,C.APP_SECRET ,C.TOKEN ,C.AES_KEY ,C.REDIRECT_URL
  FROM WECHAT_APP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED    = 1
   AND C.CODE_NAME  = ?
`

const updateWechatAppThirdPlatformTicketSQL = `
REPLACE INTO WECHAT_APP_THIRD_PLATFORM_TICKET
      (CODE_NAME    ,APP_ID     ,TICKET)
SELECT C.CODE_NAME  ,C.APP_ID   ,?
  FROM WECHAT_APP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED    = 1
   AND C.CODE_NAME  = ?
`

const queryWechatAppThirdPlatformTicketSQL = `
SELECT T.TICKET
  FROM WECHAT_APP_THIRD_PLATFORM_TICKET T
 WHERE T.CODE_NAME  = ?
`

const queryWechatAppThirdPlatformTokenSQL = `
SELECT T.APP_ID ,T.ACCESS_TOKEN ,T.UPDATED 
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_APP_THIRD_PLATFORM_TOKEN T
 WHERE T.CODE_NAME  = ?
`

const createWechatAppThirdPlatformTokenSQL = `
INSERT INTO WECHAT_APP_THIRD_PLATFORM_TOKEN
      (CODE_NAME    ,APP_ID     ,UPDATED)
SELECT C.CODE_NAME  ,C.APP_ID   ,0
  FROM WECHAT_APP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED    = 1
   AND C.CODE_NAME  = ?
`

const updateWechatAppThirdPlatformTokenSQL = `
UPDATE WECHAT_APP_THIRD_PLATFORM_TOKEN
   SET UPDATED      = 0
 WHERE CODE_NAME    = ?
   AND UPDATED      = 1
   AND EXPIRE_TIME  < NOW()
`

const uncompleteWechatAppThirdPlatformTokenSQL = `
UPDATE WECHAT_APP_THIRD_PLATFORM_TOKEN
   SET UPDATED      = 1
 WHERE CODE_NAME    = ?
   AND UPDATED      = 0
`

const completeWechatAppThirdPlatformTokenSQL = `
UPDATE WECHAT_APP_THIRD_PLATFORM_TOKEN
   SET UPDATED      = 1
      ,ACCESS_TOKEN = ?
      ,EXPIRE_TIME  = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE CODE_NAME    = ?
   AND UPDATED      = 0
`

const enableWechatAppThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_APP_THIRD_PLATFORM_AUTHORIZER
      (CODE_NAME            ,APP_ID
      ,AUTHORIZER_APP_ID    ,AUTHORIZATION_STATE
      ,AUTHORIZATION_CODE   ,PRE_AUTH_CODE)
SELECT C.CODE_NAME          ,C.APP_ID
      ,?                    ,1
      ,?                    ,?
  FROM WECHAT_APP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED    = 1
   AND C.CODE_NAME  = ?
`

const disableWechatAppThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_APP_THIRD_PLATFORM_AUTHORIZER
      (CODE_NAME            ,APP_ID
      ,AUTHORIZER_APP_ID    ,AUTHORIZATION_STATE)
SELECT C.CODE_NAME          ,C.APP_ID
      ,?                    ,0
  FROM WECHAT_APP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED    = 1
   AND C.CODE_NAME  = ?
`

const queryWechatAppThirdPlatformAuthorizerTokenSQL = `
SELECT T.APP_ID
      ,T.AUTHORIZER_APP_ID
      ,T.AUTHORIZER_ACCESS_TOKEN
      ,T.AUTHORIZER_REFRESH_TOKEN
      ,T.UPDATED
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_APP_THIRD_PLATFORM_AUTHORIZER_TOKEN T
      ,WECHAT_APP_THIRD_PLATFORM_AUTHORIZER A
 WHERE T.CODE_NAME              = ?
   AND T.AUTHORIZER_APP_ID      = ?
   AND A.CODE_NAME              = T.CODE_NAME
   AND A.AUTHORIZER_APP_ID      = T.AUTHORIZER_APP_ID
   AND A.AUTHORIZATION_STATE    = 1
`

const updateWechatAppThirdPlatformAuthorizerTokenSQL = `
UPDATE WECHAT_APP_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED              = 0
 WHERE CODE_NAME            = ?
   AND AUTHORIZER_APP_ID    = ?
   AND UPDATED              = 1
   AND EXPIRE_TIME          < NOW()
`

const uncompleteWechatAppThirdPlatformAuthorizerTokenSQL = `
UPDATE WECHAT_APP_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED              = 1
 WHERE CODE_NAME            = ?
   AND AUTHORIZER_APP_ID    = ?
   AND UPDATED              = 0
`

const completeWechatAppThirdPlatformAuthorizerTokenSQL = `
UPDATE WECHAT_APP_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED                  = 1
      ,AUTHORIZER_ACCESS_TOKEN  = ?
      ,AUTHORIZER_REFRESH_TOKEN = ?
      ,EXPIRE_TIME              = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE CODE_NAME                = ?
   AND AUTHORIZER_APP_ID        = ?
   AND UPDATED                  = 0
`

const createWechatAppThirdPlatformAuthorizerTokenSQL = `
INSERT INTO WECHAT_APP_THIRD_PLATFORM_AUTHORIZER_TOKEN
      (CODE_NAME    ,APP_ID     ,AUTHORIZER_APP_ID      ,UPDATED)
SELECT A.CODE_NAME  ,A.APP_ID   ,A.AUTHORIZER_APP_ID    ,0
  FROM WECHAT_APP_THIRD_PLATFORM_AUTHORIZER A
 WHERE A.AUTHORIZATION_STATE = 1
   AND A.CODE_NAME           = ?
   AND A.AUTHORIZER_APP_ID   = ?
`

const updateWechatAppThirdPlatformAuthorizerTokenForceSQL = `
UPDATE WECHAT_APP_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED              = 0
 WHERE CODE_NAME            = ?
   AND AUTHORIZER_APP_ID    = ?
   AND UPDATED              = 1
`
