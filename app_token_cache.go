package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    "time"
)

var wechatAPITokenConfigLifeSpan = time.Minute * 60 // config cache 60 min default
var wechatAPITokenLifeSpan = time.Minute * 5        // stable token cache 5 min default
var wechatAPITokenTempLifeSpan = time.Minute * 1    // temporary token cache 1 min default

var wechatAPITokenConfigCache *gcache.CacheTable
var wechatAPITokenCache *gcache.CacheTable

func wechatAPITokenInitialize(configMap map[string]string) {
    urlConfigLoader(configMap["wechatAPITokenURL"],
        func(configURL string) {
            wechatAPITokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatAPITokenConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatAPITokenConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAPITokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAPITokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAPITokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAPITokenTempLifeSpan = configVal * time.Minute
        })

    wechatAPITokenConfigCache = gcache.CacheExpireAfterWrite("WechatAPITokenConfig")
    wechatAPITokenConfigCache.SetDataLoader(wechatAPITokenConfigLoader)
    wechatAPITokenCache = gcache.CacheExpireAfterWrite("wechatAPIToken")
    wechatAPITokenCache.SetDataLoader(wechatAPITokenLoader)
}

type WechatAPITokenConfig struct {
    AppId     string
    AppSecret string
}

func wechatAPITokenConfigLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return configLoader(
        "WechatAPITokenConfig",
        queryWechatAPITokenConfigSQL,
        wechatAPITokenConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatAPITokenConfig)
            config.AppId = resultItem["APP_ID"]
            config.AppSecret = resultItem["APP_SECRET"]
            if 0 == len(config.AppId) || 0 == len(config.AppSecret) {
                return nil
            }
            return config
        },
        codeName, args...)
}

type WechatAPIToken struct {
    AppId       string
    AccessToken string
}

func wechatAPITokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatAPIToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
}

func wechatAPITokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

func wechatAPITokenLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoader(
        "WechatAPIToken",
        queryWechatAPITokenSQL,
        createWechatAPITokenUpdating,
        updateWechatAPITokenUpdating,
        uncompleteWechatAPITokenSQL,
        completeWechatAPITokenSQL,
        wechatAPITokenLifeSpan,
        wechatAPITokenTempLifeSpan,
        wechatAPITokenBuilder,
        wechatAPITokenRequestor,
        wechatAPITokenCompleteParamBuilder,
        codeName, args...)
}
