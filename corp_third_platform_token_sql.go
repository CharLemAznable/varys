package varys

const queryWechatCorpThirdPlatformConfigSQL = `
SELECT C.SUITE_ID ,C.SUITE_SECRET ,C.TOKEN ,C.AES_KEY ,C.REDIRECT_URL
  FROM WECHAT_CORP_THIRD_PLATFORM_CONFIG C
 WHERE C.ENABLED   = 1
   AND C.CODE_NAME = ?
`
