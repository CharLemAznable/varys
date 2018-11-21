package varys

const queryConfigurationSQL = `
SELECT C.CONFIG_NAME ,C.CONFIG_VALUE
  FROM APP_CONFIG C
 WHERE C.ENABLED = 1
`

const queryWechatAPITokenConfigSQL = `
SELECT C.APP_ID ,C.APP_SECRET
  FROM WECHAT_API_TOKEN_CONFIG C
 WHERE C.ENABLED = 1
   AND C.APP_ID  = ?
`

const queryWechatAPITokenSQL = `
SELECT T.APP_ID ,T.ACCESS_TOKEN AS TOKEN ,T.UPDATED 
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_API_TOKEN T
 WHERE T.APP_ID = ?
`

const createWechatAPITokenUpdating = `
INSERT INTO WECHAT_API_TOKEN
      (APP_ID   ,UPDATED)
VALUES(?        ,0)
`

const updateWechatAPITokenUpdating = `
UPDATE WECHAT_API_TOKEN
   SET UPDATED = 0
 WHERE APP_ID  = ?
   AND UPDATED = 1
`

const uncompleteWechatAPITokenSQL = `
UPDATE WECHAT_API_TOKEN
   SET UPDATED      = 1
 WHERE APP_ID       = ?
   AND UPDATED      = 0
`

const completeWechatAPITokenSQL = `
UPDATE WECHAT_API_TOKEN
   SET UPDATED      = 1
      ,ACCESS_TOKEN = ?
      ,EXPIRE_TIME  = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE APP_ID       = ?
   AND UPDATED      = 0
`

const queryWechatThirdPlatformConfigSQL = `
SELECT C.APP_ID ,C.APP_SECRET ,C.TOKEN ,C.AES_KEY ,C.REDIRECT_URL
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED = 1
   AND C.APP_ID  = ?
`

const replaceWechatThirdPlatformTicketSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_TICKET
      (APP_ID   ,TICKET)
VALUES(?        ,?)
`

const queryWechatThirdPlatformTicketSQL = `
SELECT T.TICKET
  FROM WECHAT_THIRD_PLATFORM_TICKET T
 WHERE T.APP_ID = ?
`

const queryWechatThirdPlatformTokenSQL = `
SELECT T.APP_ID ,T.COMPONENT_ACCESS_TOKEN AS TOKEN ,T.UPDATED 
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_THIRD_PLATFORM_TOKEN T
 WHERE T.APP_ID = ?
`

const createWechatThirdPlatformTokenUpdating = `
INSERT INTO WECHAT_THIRD_PLATFORM_TOKEN
      (APP_ID   ,UPDATED)
VALUES(?        ,0)
`

const updateWechatThirdPlatformTokenUpdating = `
UPDATE WECHAT_THIRD_PLATFORM_TOKEN
   SET UPDATED = 0
 WHERE APP_ID  = ?
   AND UPDATED = 1
`

const uncompleteWechatThirdPlatformTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_TOKEN
   SET UPDATED                = 1
 WHERE APP_ID                 = ?
   AND UPDATED                = 0
`

const completeWechatThirdPlatformTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_TOKEN
   SET UPDATED                = 1
      ,COMPONENT_ACCESS_TOKEN = ?
      ,EXPIRE_TIME            = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE APP_ID                 = ?
   AND UPDATED                = 0
`

const queryWechatThirdPlatformPreAuthCodeSQL = `
SELECT T.APP_ID ,T.PRE_AUTH_CODE AS TOKEN ,T.UPDATED 
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE T
 WHERE T.APP_ID = ?
`

const createWechatThirdPlatformPreAuthCodeUpdating = `
INSERT INTO WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
      (APP_ID   ,UPDATED)
VALUES(?        ,0)
`

const updateWechatThirdPlatformPreAuthCodeUpdating = `
UPDATE WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
   SET UPDATED = 0
 WHERE APP_ID  = ?
   AND UPDATED = 1
`

const uncompleteWechatThirdPlatformPreAuthCodeSQL = `
UPDATE WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
   SET UPDATED       = 1
 WHERE APP_ID        = ?
   AND UPDATED       = 0
`

const completeWechatThirdPlatformPreAuthCodeSQL = `
UPDATE WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
   SET UPDATED       = 1
      ,PRE_AUTH_CODE = ?
      ,EXPIRE_TIME   = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE APP_ID        = ?
   AND UPDATED       = 0
`

const enableWechatThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_AUTHORIZER
      (APP_ID               ,AUTHORIZER_APP_ID  ,AUTHORIZATION_STATE  
      ,AUTHORIZATION_CODE   ,PRE_AUTH_CODE)
VALUES(?                    ,?                  ,1
      ,?                    ,?)
`

const disableWechatThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_AUTHORIZER
      (APP_ID               ,AUTHORIZER_APP_ID  ,AUTHORIZATION_STATE)
VALUES(?                    ,?                  ,0)
`

const queryWechatThirdPlatformAuthorizerTokenSQL = `
SELECT T.APP_ID ,T.AUTHORIZER_APP_ID
      ,T.AUTHORIZER_ACCESS_TOKEN
      ,T.AUTHORIZER_REFRESH_TOKEN
      ,T.UPDATED ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN T
      ,WECHAT_THIRD_PLATFORM_AUTHORIZER A
 WHERE T.APP_ID              = ?
   AND T.AUTHORIZER_APP_ID   = ?
   AND A.APP_ID              = T.APP_ID
   AND A.AUTHORIZER_APP_ID   = T.AUTHORIZER_APP_ID
   AND A.AUTHORIZATION_STATE = 1
`

const updateWechatThirdPlatformAuthorizerTokenUpdating = `
UPDATE WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED           = 0
 WHERE APP_ID            = ?
   AND AUTHORIZER_APP_ID = ?
   AND UPDATED           = 1
`

const uncompleteWechatThirdPlatformAuthorizerTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED                  = 1
 WHERE APP_ID                   = ?
   AND AUTHORIZER_APP_ID        = ?
   AND UPDATED                  = 0
`

const completeWechatThirdPlatformAuthorizerTokenSQL = `
UPDATE WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
   SET UPDATED                  = 1
      ,AUTHORIZER_ACCESS_TOKEN  = ?
      ,AUTHORIZER_REFRESH_TOKEN = ?
      ,EXPIRE_TIME              = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE APP_ID                   = ?
   AND AUTHORIZER_APP_ID        = ?
   AND UPDATED                  = 0
`

const createWechatThirdPlatformAuthorizerTokenUpdating = `
INSERT INTO WECHAT_THIRD_PLATFORM_AUTHORIZER_TOKEN
      (APP_ID   ,AUTHORIZER_APP_ID  ,UPDATED)
VALUES(?        ,?                  ,0)
`
