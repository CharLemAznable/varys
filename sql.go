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
   AND C.APP_ID = ?
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
 WHERE APP_ID = ?
   AND UPDATED = 1
`

const replaceWechatAPITokenSQL = `
REPLACE INTO WECHAT_API_TOKEN
      (APP_ID   ,ACCESS_TOKEN
      ,UPDATED  ,EXPIRE_TIME)
VALUES(?        ,?
      ,1        ,DATE_ADD(NOW(), INTERVAL ? SECOND))
`

const queryWechatThirdPlatformConfigSQL = `
SELECT C.APP_ID ,C.APP_SECRET ,C.TOKEN ,C.AES_KEY ,C.REDIRECT_URL
  FROM WECHAT_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED = 1
   AND C.APP_ID = ?
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
 WHERE APP_ID = ?
   AND UPDATED = 1
`

const replaceWechatThirdPlatformTokenSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_TOKEN
      (APP_ID   ,COMPONENT_ACCESS_TOKEN
      ,UPDATED  ,EXPIRE_TIME)
VALUES(?        ,?
      ,1        ,DATE_ADD(NOW(), INTERVAL ? SECOND))
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
 WHERE APP_ID = ?
   AND UPDATED = 1
`

const replaceWechatThirdPlatformPreAuthCodeSQL = `
REPLACE INTO WECHAT_THIRD_PLATFORM_PRE_AUTH_CODE
      (APP_ID   ,PRE_AUTH_CODE
      ,UPDATED  ,EXPIRE_TIME)
VALUES(?        ,?
      ,1        ,DATE_ADD(NOW(), INTERVAL ? SECOND))
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
