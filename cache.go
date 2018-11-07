package main

import (
    "github.com/CharLemAznable/gcache"
    "log"
    "time"
)

var wechatAPITokenConfigLifeSpan = time.Minute * 60 // config cache 60 min default
var wechatAPITokenLifeSpan = time.Minute * 5        // stable token cache 5 min default
var wechatAPITokenTempLifeSpan = time.Minute * 1    // temporary token cache 1 min default

var wechatAPITokenConfigCache *gcache.CacheTable
var wechatAPITokenCache *gcache.CacheTable

func wechatAPITokenConfigLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
    resultMap, err := db.Sql(queryWechatAPITokenConfigSQL).Params(appId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil // require config
    }
    log.Printf("Query WechatAPITokenConfig: %s", resultMap)

    resultItem := resultMap[0]
    config := new(WechatAPITokenConfig)
    config.AppId = resultItem["APP_ID"]
    config.AppSecret = resultItem["APP_SECRET"]
    if 0 == len(config.AppId) || 0 == len(config.AppSecret) {
        return nil
    }
    log.Printf("Load WechatAPITokenConfig Cache: %s", Json(config))
    return gcache.NewCacheItem(appId, wechatAPITokenConfigLifeSpan, config)
}

func wechatAPITokenLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
    resultMap, err := db.Sql(queryWechatAPITokenSQL).Params(appId).Query()
    if nil != err || 1 != len(resultMap) {
        token, err := requestWechatAPITokenAndReplaceToDB(appId.(string))
        if nil != err {
            return nil
        }
        log.Printf("Request WechatAPIToken: %s", Json(token))
        return gcache.NewCacheItem(appId, wechatAPITokenLifeSpan, token)
    }
    log.Printf("Query WechatAPIToken: %s", resultMap)

    resultItem := resultMap[0]
    updated := resultItem["UPDATED"]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err {
        return nil
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        log.Println("Try to request and update WechatAPIToken")
        count, err := updatingWechatAPIToken(appId.(string))
        if nil != err {
            return nil
        }
        if count > 0 {
            token, err := requestWechatAPITokenAndReplaceToDB(appId.(string))
            if nil != err {
                return nil
            }
            log.Printf("Request WechatAPIToken: %s", Json(token))
            return gcache.NewCacheItem(appId, wechatAPITokenLifeSpan, token)
        }
        log.Println("Give up request and update WechatAPIToken, use Query result Temporarily")
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    lifeSpan := Condition(isExpired, wechatAPITokenTempLifeSpan, wechatAPITokenLifeSpan).(time.Duration)
    token := new(WechatAPIToken)
    token.AppId = appId.(string)
    token.AccessToken = resultItem["ACCESS_TOKEN"]
    log.Printf("Load WechatAPIToken Cache: %s, cache %3.1f min", Json(token), lifeSpan.Minutes())
    return gcache.NewCacheItem(appId, lifeSpan, token)
}

func updatingWechatAPIToken(appId string) (int64, error) {
    count, err := db.Sql(updateWechatAPITokenUpdating).Params(appId).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func requestWechatAPITokenAndReplaceToDB(appId string) (*WechatAPIToken, error) {
    response, err := requestWechatAPIToken(appId)
    if nil != err {
        return nil, err
    }
    log.Printf("Request WechatAPIToken Response: %s", Json(response))
    // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    expireTimeInc := response.ExpiresIn - int(wechatAPITokenLifeSpan.Seconds()*1.1)
    count, err := db.Sql(replaceWechatAPITokenSQL).
        Params(appId, response.AccessToken, expireTimeInc).Execute()
    if nil != err {
        return nil, err
    }
    if count < 1 {
        return nil, &UnexpectedError{Message: "ReplaceWechatAPIToken Failed"}
    }

    token := new(WechatAPIToken)
    token.AppId = appId
    token.AccessToken = response.AccessToken
    return token, nil
}
