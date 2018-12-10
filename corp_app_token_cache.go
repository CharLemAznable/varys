package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "time"
)

var wechatCorpTokenConfigLifeSpan = time.Minute * 60    // config cache 60 min default
var wechatCorpTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

var wechatCorpTokenConfigCache *gcache.CacheTable
var wechatCorpTokenCache *gcache.CacheTable

func wechatCorpTokenInitialize(configMap map[string]string) {
    urlConfigLoader(configMap["wechatCorpTokenURL"],
        func(configURL string) {
            wechatCorpTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatCorpTokenConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpTokenConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpTokenMaxLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpTokenMaxLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpTokenExpireCriticalSpan"],
        func(configVal time.Duration) {
            wechatCorpTokenExpireCriticalSpan = configVal * time.Second
        })

    wechatCorpTokenConfigCache = gcache.CacheExpireAfterWrite("wechatCorpTokenConfig")
    wechatCorpTokenConfigCache.SetDataLoader(wechatCorpTokenConfigLoader)
    wechatCorpTokenCache = gcache.CacheExpireAfterWrite("wechatCorpToken")
    wechatCorpTokenCache.SetDataLoader(wechatCorpTokenLoader)
}

type WechatCorpTokenConfig struct {
    CorpId     string
    CorpSecret string
}

func wechatCorpTokenConfigLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return configLoader(
        "WechatCorpTokenConfig",
        queryWechatCorpTokenConfigSQL,
        wechatCorpTokenConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatCorpTokenConfig)
            config.CorpId = resultItem["CORP_ID"]
            config.CorpSecret = resultItem["CORP_SECRET"]
            if 0 == len(config.CorpId) || 0 == len(config.CorpSecret) {
                return nil
            }
            return config
        },
        codeName, args...)
}

type WechatCorpToken struct {
    CorpId      string
    AccessToken string
}

func wechatCorpTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatCorpToken)
    tokenItem.CorpId = resultItem["CORP_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
}

func wechatCorpTokenSQLParamBuilder(resultItem map[string]string, codeName interface{}) []interface{} {
    expireTime, _ := Int64FromStr(resultItem["EXPIRE_TIME"])
    return []interface{}{resultItem["ACCESS_TOKEN"], expireTime, codeName}
}

func wechatCorpTokenLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    resultMap, err := db.Sql(queryWechatCorpTokenSQL).Params(codeName).Query()
    if nil != err || 1 != len(resultMap) {
        log.Info("Try to request WechatCorpToken:(%s)", codeName)

        resultItem, err := wechatCorpTokenRequestor(codeName)
        if nil != err {
            log.Warn("Request WechatCorpToken Failed:(%s) %s", codeName, err.Error())
            return nil, err
        }

        count, err := db.Sql(createWechatCorpTokenSQL).Params(
            wechatCorpTokenSQLParamBuilder(resultItem, codeName)...).Execute()
        if nil != err || count < 1 { // 插入记录失败, 则查询记录并缓存
            return queryNewestWechatCorpToken(codeName)
        }

        // 创建成功, 缓存最大缓存时长
        tokenItem := wechatCorpTokenBuilder(resultItem)
        log.Info("Request WechatCorpToken:(%s) %s", codeName, Json(tokenItem))
        return gcache.NewCacheItem(codeName, wechatCorpTokenMaxLifeSpan, tokenItem), nil
    }
    log.Trace("Query WechatCorpToken:(%s) %s", codeName, resultMap)

    resultItem := resultMap[0]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
    if nil != err {
        return nil, err
    }
    effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
    // 即将过期 -> 触发更新
    if effectiveSpan <= wechatCorpTokenExpireCriticalSpan {
        log.Info("Try to request and update WechatCorpToken:(%s)", codeName)

        // 休眠后再请求最新的access_token
        time.Sleep(wechatCorpTokenExpireCriticalSpan)
        resultItem, err := wechatCorpTokenRequestor(codeName)
        if nil != err {
            log.Warn("Request WechatCorpToken Failed:(%s) %s", codeName, err.Error())
            return nil, err
        }

        count, err := db.Sql(updateWechatCorpTokenSQL).Params(
            wechatCorpTokenSQLParamBuilder(resultItem, codeName)...).Execute()
        if nil != err || count < 1 { // 更新记录失败, 则查询记录并缓存
            return queryNewestWechatCorpToken(codeName)
        }

        // 更新成功, 缓存最大缓存时长
        tokenItem := wechatCorpTokenBuilder(resultItem)
        log.Info("Request and Update WechatCorpToken:(%s) %s", codeName, Json(tokenItem))
        return gcache.NewCacheItem(codeName, wechatCorpTokenMaxLifeSpan, tokenItem), nil
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长, 即加快缓存更新频率
    ls := Condition(effectiveSpan > wechatCorpTokenMaxLifeSpan,
        wechatCorpTokenMaxLifeSpan, effectiveSpan).(time.Duration)
    tokenItem := wechatCorpTokenBuilder(resultItem)
    log.Info("Load WechatCorpToken Cache:(%s) %s, cache %3.1f min", codeName, Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(codeName, ls, tokenItem), nil
}

func queryNewestWechatCorpToken(codeName interface{}) (*gcache.CacheItem, error) {
    resultMap, err := db.Sql(queryWechatCorpTokenSQL).Params(codeName).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, DefaultIfNil(err, &UnexpectedError{
            Message: "Record WechatCorpToken Failed"}).(error)
    }

    resultItem := resultMap[0]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
    if nil != err {
        return nil, err
    }
    effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
    if effectiveSpan <= wechatCorpTokenExpireCriticalSpan {
        return nil, &UnexpectedError{Message: "Record WechatCorpToken Failed"}
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长, 即加快缓存更新频率
    ls := Condition(effectiveSpan > wechatCorpTokenMaxLifeSpan,
        wechatCorpTokenMaxLifeSpan, effectiveSpan).(time.Duration)
    tokenItem := wechatCorpTokenBuilder(resultItem)
    log.Info("Load WechatCorpToken Cache:(%s) %s, cache %3.1f min", codeName, Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(codeName, ls, tokenItem), nil
}
