package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    "time"
)

var wechatAppTokenConfigLifeSpan = time.Minute * 60 // config cache 60 min default
var wechatAppTokenLifeSpan = time.Minute * 5        // stable token cache 5 min default
var wechatAppTokenTempLifeSpan = time.Minute * 1    // temporary token cache 1 min default

var wechatAppTokenConfigCache *gcache.CacheTable
var wechatAppTokenCache *gcache.CacheTable

func wechatAppTokenInitialize(configMap map[string]string) {
    urlConfigLoader(configMap["wechatAppTokenURL"],
        func(configURL string) {
            wechatAppTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatAppTokenConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatAppTokenConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAppTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAppTokenTempLifeSpan = configVal * time.Minute
        })

    wechatAppTokenConfigCache = gcache.CacheExpireAfterWrite("WechatAppTokenConfig")
    wechatAppTokenConfigCache.SetDataLoader(wechatAppTokenConfigLoader)
    wechatAppTokenCache = gcache.CacheExpireAfterWrite("wechatAppToken")
    wechatAppTokenCache.SetDataLoader(wechatAppTokenLoader)
}

type WechatAppTokenConfig struct {
    AppId     string
    AppSecret string
}

func wechatAppTokenConfigLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return configLoader(
        "WechatAppTokenConfig",
        queryWechatAppTokenConfigSQL,
        wechatAppTokenConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatAppTokenConfig)
            config.AppId = resultItem["APP_ID"]
            config.AppSecret = resultItem["APP_SECRET"]
            if 0 == len(config.AppId) || 0 == len(config.AppSecret) {
                return nil
            }
            return config
        },
        codeName, args...)
}

type WechatAppToken struct {
    AppId       string
    AccessToken string
}

func wechatAppTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatAppToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
}

func wechatAppTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

func wechatAppTokenLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoader(
        "WechatAppToken",
        queryWechatAppTokenSQL,
        createWechatAppTokenSQL,
        updateWechatAppTokenSQL,
        uncompleteWechatAppTokenSQL,
        completeWechatAppTokenSQL,
        wechatAppTokenLifeSpan,
        wechatAppTokenTempLifeSpan,
        wechatAppTokenBuilder,
        wechatAppTokenRequestor,
        wechatAppTokenCompleteParamBuilder,
        codeName, args...)
}
