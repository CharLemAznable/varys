package varys

import (
    . "github.com/CharLemAznable/gokits"
    "time"
)

var wechatAppConfigLifeSpan = time.Minute * 60   // config cache 60 min default
var wechatAppTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatAppTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

var wechatAppConfigCache *CacheTable
var wechatAppTokenCache *CacheTable

func wechatAppTokenInitialize(configMap map[string]string) {
    urlConfigLoader(configMap["wechatAppTokenURL"],
        func(configURL string) {
            wechatAppTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatAppConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatAppConfigLifeSpan = configVal * time.Minute
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

    wechatAppConfigCache = CacheExpireAfterWrite("WechatAppConfig")
    wechatAppConfigCache.SetDataLoader(wechatAppConfigLoader)
    wechatAppTokenCache = CacheExpireAfterWrite("wechatAppToken")
    wechatAppTokenCache.SetDataLoader(wechatAppTokenLoader)
}

type WechatAppConfig struct {
    AppId     string
    AppSecret string
}

func wechatAppConfigLoader(codeName interface{}, args ...interface{}) (*CacheItem, error) {
    return configLoader(
        "WechatAppConfig",
        queryWechatAppConfigSQL,
        wechatAppConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatAppConfig)
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

func wechatAppTokenLoader(codeName interface{}, args ...interface{}) (*CacheItem, error) {
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
