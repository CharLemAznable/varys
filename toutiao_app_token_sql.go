package main

const queryToutiaoAppConfigSQL = `
SELECT C.APP_ID ,C.APP_SECRET
  FROM TOUTIAO_APP_CONFIG C
 WHERE C.ENABLED    = 1
   AND C.CODE_NAME  = ?
`

const queryToutiaoAppTokenSQL = `
SELECT T.APP_ID ,T.ACCESS_TOKEN ,T.UPDATED
      ,UNIX_TIMESTAMP(T.EXPIRE_TIME) AS EXPIRE_TIME
  FROM TOUTIAO_APP_TOKEN T
 WHERE T.CODE_NAME  = ?
`

const createToutiaoAppTokenSQL = `
INSERT INTO TOUTIAO_APP_TOKEN
      (CODE_NAME    ,APP_ID     ,UPDATED)
SELECT C.CODE_NAME  ,C.APP_ID   ,0
  FROM TOUTIAO_APP_CONFIG C
 WHERE C.ENABLED    = 1
   AND C.CODE_NAME  = ?
`

const updateToutiaoAppTokenSQL = `
UPDATE TOUTIAO_APP_TOKEN
   SET UPDATED      = 0
 WHERE CODE_NAME    = ?
   AND UPDATED      = 1
   AND EXPIRE_TIME  < NOW()
`

const uncompleteToutiaoAppTokenSQL = `
UPDATE TOUTIAO_APP_TOKEN
   SET UPDATED      = 1
 WHERE CODE_NAME    = ?
   AND UPDATED      = 0
`

const completeToutiaoAppTokenSQL = `
UPDATE TOUTIAO_APP_TOKEN
   SET UPDATED      = 1
      ,ACCESS_TOKEN = ?
      ,EXPIRE_TIME  = DATE_ADD(NOW(), INTERVAL ? SECOND)
 WHERE CODE_NAME    = ?
   AND UPDATED      = 0
`