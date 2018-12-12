package varys

import (
    "fmt"
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatCorpThirdPlatformConfigLifeSpan = time.Minute * 60             // config cache 60 min default
var wechatCorpThirdPlatformCryptorLifeSpan = time.Minute * 60            // cryptor cache 60 min default
var wechatCorpThirdPlatformTokenMaxLifeSpan = time.Minute * 5            // stable token cache 5 min max
var wechatCorpThirdPlatformTokenExpireCriticalSpan = time.Second * 1     // token about to expire critical time span
var wechatCorpThirdPlatformPermanentCodeLifeSpan = time.Minute * 60      // permanent_code cache 60 min default
var wechatCorpThirdPlatformCorpTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpThirdPlatformCorpTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

var wechatCorpThirdPlatformConfigCache *gcache.CacheTable
var wechatCorpThirdPlatformCryptorCache *gcache.CacheTable
var wechatCorpThirdPlatformTokenCache *gcache.CacheTable
var wechatCorpThirdPlatformPermanentCodeCache *gcache.CacheTable
var wechatCorpThirdPlatformCorpTokenCache *gcache.CacheTable

func wechatCorpThirdPlatformAuthorizerTokenInitialize(configMap map[string]string) {
    urlConfigLoader(configMap["wechatCorpThirdPlatformTokenURL"],
        func(configURL string) {
            wechatCorpThirdPlatformTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatCorpThirdPlatformConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpThirdPlatformConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpThirdPlatformCryptorLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpThirdPlatformCryptorLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpThirdPlatformTokenMaxLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpThirdPlatformTokenMaxLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpThirdPlatformTokenExpireCriticalSpan"],
        func(configVal time.Duration) {
            wechatCorpThirdPlatformTokenExpireCriticalSpan = configVal * time.Second
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpThirdPlatformPermanentCodeLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpThirdPlatformPermanentCodeLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpThirdPlatformCorpTokenMaxLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpThirdPlatformCorpTokenMaxLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpThirdPlatformCorpTokenExpireCriticalSpan"],
        func(configVal time.Duration) {
            wechatCorpThirdPlatformCorpTokenExpireCriticalSpan = configVal * time.Second
        })

    wechatCorpThirdPlatformConfigCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformConfig")
    wechatCorpThirdPlatformConfigCache.SetDataLoader(wechatCorpThirdPlatformConfigLoader)
    wechatCorpThirdPlatformCryptorCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformCryptor")
    wechatCorpThirdPlatformCryptorCache.SetDataLoader(wechatCorpThirdPlatformCryptorLoader)
    wechatCorpThirdPlatformTokenCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformCryptor")
    wechatCorpThirdPlatformTokenCache.SetDataLoader(wechatCorpThirdPlatformTokenLoader)
    wechatCorpThirdPlatformPermanentCodeCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformPermanentCode")
    wechatCorpThirdPlatformPermanentCodeCache.SetDataLoader(wechatCorpThirdPlatformPermanentCodeLoader)
    wechatCorpThirdPlatformCorpTokenCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformCorpToken")
    wechatCorpThirdPlatformCorpTokenCache.SetDataLoader(wechatCorpThirdPlatformCorpTokenLoader)
}

type WechatCorpThirdPlatformConfig struct {
    SuiteId     string
    SuiteSecret string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatCorpThirdPlatformConfigLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return configLoader(
        "WechatCorpThirdPlatformConfig",
        queryWechatCorpThirdPlatformConfigSQL,
        wechatCorpThirdPlatformConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatCorpThirdPlatformConfig)
            config.SuiteId = resultItem["SUITE_ID"]
            config.SuiteSecret = resultItem["SUITE_SECRET"]
            config.Token = resultItem["TOKEN"]
            config.AesKey = resultItem["AES_KEY"]
            config.RedirectURL = resultItem["REDIRECT_URL"]
            if 0 == len(config.SuiteId) || 0 == len(config.SuiteSecret) ||
                0 == len(config.Token) || 0 == len(config.AesKey) {
                return nil
            }
            return config
        },
        codeName, args...)
}

func wechatCorpThirdPlatformCryptorLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    cache, err := wechatCorpThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, &UnexpectedError{Message:
        "Require WechatCorpThirdPlatformConfig with key: " + codeName.(string)} // require config
    }
    config := cache.Data().(*WechatCorpThirdPlatformConfig)
    log.Trace("Query WechatCorpThirdPlatformConfig Cache:(%s) %s", codeName, Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.SuiteId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    log.Info("Load WechatCorpThirdPlatformCryptor Cache:(%s) %s", codeName, cryptor)
    return gcache.NewCacheItem(codeName, wechatCorpThirdPlatformCryptorLifeSpan, cryptor), nil
}

type WechatCorpThirdPlatformToken struct {
    SuiteId          string
    SuiteAccessToken string
}

func wechatCorpThirdPlatformTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatCorpThirdPlatformToken)
    tokenItem.SuiteId = resultItem["SUITE_ID"]
    tokenItem.SuiteAccessToken = resultItem["SUITE_ACCESS_TOKEN"]
    return tokenItem
}

func wechatCorpThirdPlatformTokenSQLParamBuilder(resultItem map[string]string, codeName interface{}) []interface{} {
    expireTime, _ := Int64FromStr(resultItem["EXPIRE_TIME"])
    return []interface{}{resultItem["SUITE_ACCESS_TOKEN"], expireTime, codeName}
}

func wechatCorpThirdPlatformTokenLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoaderStrict(
        "WechatCorpThirdPlatformToken",
        queryWechatCorpThirdPlatformTokenSQL,
        createWechatCorpThirdPlatformTokenUpdating,
        updateWechatCorpThirdPlatformTokenUpdating,
        wechatCorpThirdPlatformTokenMaxLifeSpan,
        wechatCorpThirdPlatformTokenExpireCriticalSpan,
        wechatCorpThirdPlatformTokenBuilder,
        wechatCorpThirdPlatformTokenRequestor,
        wechatCorpThirdPlatformTokenSQLParamBuilder,
        codeName, args...)
}

func wechatCorpThirdPlatformAuthorizeCreator(codeName, authCode interface{}) {
    resultItem, err := wechatCorpThirdPlatformPermanentCodeRequestor(codeName, authCode)
    if nil != err {
        log.Warn("Request WechatCorpThirdPlatformPermanentCode Failed:(%s, authCode:%s) %s",
            codeName, authCode, err.Error())
        return
    }

    enableWechatCorpThirdPlatformAuthorizer(codeName.(string), resultItem["CORP_ID"], resultItem["PERMANENT_CODE"])

    // TODO insert/upate access_token, cache it
}

type WechatCorpThirdPlatformAuthorizerKey struct {
    CodeName string
    CorpId   string
}

type WechatCorpThirdPlatformPermanentCode struct {
    SuiteId       string
    CorpId        string
    PermanentCode string
}

func wechatCorpThirdPlatformPermanentCodeLoader(key interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    corpKey, ok := key.(WechatCorpThirdPlatformAuthorizerKey)
    if !ok {
        return nil, &UnexpectedError{Message:
        "WechatCorpThirdPlatformAuthorizerKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatCorpThirdPlatformPermanentCodeSQL).
        Params(corpKey.CodeName, corpKey.CorpId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
        "Unauthorized corp: " + Json(key)}).(error) // requires that the token already exists
    }
    log.Trace("Query WechatCorpThirdPlatformPermanentCode:(%s) %s", Json(key), resultMap)

    resultItem := resultMap[0]
    codeItem := new(WechatCorpThirdPlatformPermanentCode)
    codeItem.SuiteId = resultItem["SUITE_ID"]
    codeItem.CorpId = resultItem["CORP_ID"]
    codeItem.PermanentCode = resultItem["PERMANENT_CODE"]
    log.Info("Load WechatCorpThirdPlatformPermanentCode Cache:(%s) %s", Json(key), Json(codeItem))
    return gcache.NewCacheItem(key, wechatCorpThirdPlatformPermanentCodeLifeSpan, codeItem), nil
}

type WechatCorpThirdPlatformCorpToken struct {
    SuiteId         string
    CorpId          string
    CorpAccessToken string
}

func wechatCorpThirdPlatformCorpTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatCorpThirdPlatformCorpToken)
    tokenItem.SuiteId = resultItem["SUITE_ID"]
    tokenItem.CorpId = resultItem["CORP_ID"]
    tokenItem.CorpAccessToken = resultItem["CORP_ACCESS_TOKEN"]
    return tokenItem
}

func wechatCorpThirdPlatformCorpTokenLoader(key interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    corpKey, ok := key.(WechatCorpThirdPlatformAuthorizerKey)
    if !ok {
        return nil, &UnexpectedError{Message:
        "WechatCorpThirdPlatformAuthorizerKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatCorpThirdPlatformCorpTokenSQL).
        Params(corpKey.CodeName, corpKey.CorpId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
        "Unauthorized corp: " + Json(key)}).(error) // requires that the token already exists
    }
    log.Trace("Query WechatCorpThirdPlatformCorpToken:(%s) %s", Json(key), resultMap)

    resultItem := resultMap[0]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
    if nil != err {
        return nil, err
    }
    effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
    // 即将过期 -> 触发更新
    if effectiveSpan <= wechatCorpThirdPlatformCorpTokenExpireCriticalSpan {
        time.Sleep(wechatCorpThirdPlatformCorpTokenExpireCriticalSpan) // 休眠后再请求最新的access_token
        log.Info("Try to request and update WechatCorpThirdPlatformCorpToken:(%s)", Json(key))

        resultItem, err := wechatCorpThirdPlatformCorpTokenRequestor(corpKey.CodeName, corpKey.CorpId)
        if nil != err {
            log.Warn("Request WechatCorpThirdPlatformCorpToken Failed:(%s) %s", key, err.Error())
            return nil, err
        }

        expireTime, _ := Int64FromStr(resultItem["EXPIRE_TIME"])
        count, err := db.New().Sql(updateWechatCorpThirdPlatformCorpTokenSQL).
            Params(resultItem["CORP_ACCESS_TOKEN"], expireTime, corpKey.CodeName, corpKey.CorpId).Execute()
        if nil != err || count < 1 { // 记录入库失败, 则查询记录并返回
            resultMap, err := db.New().Sql(queryWechatCorpThirdPlatformCorpTokenSQL).
                Params(corpKey.CodeName, corpKey.CorpId).Query()
            if nil != err || 1 != len(resultMap) {
                return nil, DefaultIfNil(err, &UnexpectedError{Message:
                    fmt.Sprintf("Query WechatCorpThirdPlatformCorpToken:(%s) Failed", Json(key))}).(error)
            }

            resultItem := resultMap[0]
            expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
            if nil != err {
                return nil, err
            }
            effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
            if effectiveSpan <= wechatCorpThirdPlatformCorpTokenExpireCriticalSpan {
                return nil, &UnexpectedError{Message:
                fmt.Sprintf("Query WechatCorpThirdPlatformCorpToken:(%s) expireTime Failed", Json(key))}
            }

            // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长
            ls := Condition(effectiveSpan > wechatCorpThirdPlatformCorpTokenMaxLifeSpan,
                wechatCorpThirdPlatformCorpTokenMaxLifeSpan, effectiveSpan).(time.Duration)
            tokenItem := wechatCorpThirdPlatformCorpTokenBuilder(resultItem)
            log.Info("Load WechatCorpThirdPlatformCorpToken Cache:(%s) %s, cache %3.1f min",
                Json(key), Json(tokenItem), ls.Minutes())
            return gcache.NewCacheItem(key, ls, tokenItem), nil
        }

        // 记录入库成功, 缓存最大缓存时长
        tokenItem := wechatCorpThirdPlatformCorpTokenBuilder(resultItem)
        log.Info("Load WechatCorpThirdPlatformCorpToken Cache:(%s) %s, cache %3.1f min",
            Json(key), Json(tokenItem), wechatCorpThirdPlatformCorpTokenMaxLifeSpan.Minutes())
        return gcache.NewCacheItem(key, wechatCorpThirdPlatformCorpTokenMaxLifeSpan, tokenItem), nil
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长
    ls := Condition(effectiveSpan > wechatCorpThirdPlatformCorpTokenMaxLifeSpan,
        wechatCorpThirdPlatformCorpTokenMaxLifeSpan, effectiveSpan).(time.Duration)
    tokenItem := wechatCorpThirdPlatformCorpTokenBuilder(resultItem)
    log.Info("Load WechatCorpThirdPlatformCorpToken Cache:(%s) %s, cache %3.1f min",
        Json(key), Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(key, ls, tokenItem), nil
}