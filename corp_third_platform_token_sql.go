package varys

const queryWechatCorpThirdPlatformConfigSQL = `
SELECT C.SUITE_ID ,C.SUITE_SECRET ,C.TOKEN ,C.AES_KEY ,C.REDIRECT_URL
  FROM WECHAT_CORP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const replaceWechatCorpThirdPlatformTicketSQL = `
REPLACE INTO WECHAT_CORP_THIRD_PLATFORM_TICKET
      (CODE_NAME    ,SUITE_ID   ,TICKET)
SELECT C.CODE_NAME  ,C.SUITE_ID ,?
  FROM WECHAT_CORP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const queryWechatCorpThirdPlatformTicketSQL = `
SELECT T.TICKET
  FROM WECHAT_CORP_THIRD_PLATFORM_TICKET T
 WHERE T.CODE_NAME = ?
`

const queryWechatCorpThirdPlatformTokenSQL = `
SELECT T.SUITE_ID ,T.SUITE_ACCESS_TOKEN 
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_CORP_THIRD_PLATFORM_TOKEN T
 WHERE T.CODE_NAME = ?
`

const createWechatCorpThirdPlatformTokenUpdating = `
INSERT INTO WECHAT_CORP_THIRD_PLATFORM_TOKEN
      (CODE_NAME            ,SUITE_ID
      ,SUITE_ACCESS_TOKEN   ,EXPIRE_TIME)
SELECT C.CODE_NAME          ,C.SUITE_ID
      ,?                    ,FROM_UNIXTIME(?)
  FROM WECHAT_CORP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const updateWechatCorpThirdPlatformTokenUpdating = `
UPDATE WECHAT_CORP_THIRD_PLATFORM_TOKEN
   SET SUITE_ACCESS_TOKEN   = ?
      ,EXPIRE_TIME          = FROM_UNIXTIME(?)
 WHERE CODE_NAME            = ?
   AND EXPIRE_TIME          < NOW()
`

const enableWechatCorpThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_CORP_THIRD_PLATFORM_AUTHORIZER
      (CODE_NAME    ,SUITE_ID
      ,CORP_ID      ,STATE      ,PERMANENT_CODE)
SELECT C.CODE_NAME  ,C.SUITE_ID
      ,?            ,1          ,?
  FROM WECHAT_CORP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const disableWechatCorpThirdPlatformAuthorizerSQL = `
REPLACE INTO WECHAT_CORP_THIRD_PLATFORM_AUTHORIZER
      (CODE_NAME    ,SUITE_ID
      ,CORP_ID      ,STATE)
SELECT C.CODE_NAME  ,C.SUITE_ID
      ,?            ,0
  FROM WECHAT_CORP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const createWechatCorpThirdPlatformCorpTokenSQL = `
INSERT INTO WECHAT_CORP_THIRD_PLATFORM_CORP_TOKEN
      (CODE_NAME    ,SUITE_ID
      ,CORP_ID      ,CORP_ACCESS_TOKEN  ,EXPIRE_TIME)
SELECT C.CODE_NAME  ,C.SUITE_ID
      ,?            ,?                  ,FROM_UNIXTIME(?)
  FROM WECHAT_CORP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`

const queryWechatCorpThirdPlatformPermanentCodeSQL = `
SELECT A.SUITE_ID ,A.CORP_ID ,A.PERMANENT_CODE
  FROM WECHAT_CORP_THIRD_PLATFORM_AUTHORIZER A
 WHERE A.STATE      = 1
   AND A.CODE_NAME  = ?
   AND A.CORP_ID    = ?
`

const queryWechatCorpThirdPlatformCorpTokenSQL = `
SELECT T.SUITE_ID ,T.CORP_ID ,T.CORP_ACCESS_TOKEN
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM WECHAT_CORP_THIRD_PLATFORM_CORP_TOKEN T
      ,WECHAT_CORP_THIRD_PLATFORM_AUTHORIZER A
 WHERE T.CODE_NAME  = ?
   AND T.CORP_ID    = ?
   AND A.STATE      = 1
   AND A.CODE_NAME  = T.CODE_NAME
   AND A.CORP_ID    = T.CORP_ID
`

const updateWechatCorpThirdPlatformCorpTokenSQL = `
UPDATE WECHAT_CORP_THIRD_PLATFORM_CORP_TOKEN
   SET CORP_ACCESS_TOKEN = ?
      ,EXPIRE_TIME       = FROM_UNIXTIME(?)
 WHERE CODE_NAME         = ?
   AND CORP_ID           = ?
   AND EXPIRE_TIME       < NOW()
`
