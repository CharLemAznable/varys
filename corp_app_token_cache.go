package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
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
    return tokenLoaderStrict(
        "WechatCorpToken",
        queryWechatCorpTokenSQL,
        createWechatCorpTokenSQL,
        updateWechatCorpTokenSQL,
        wechatCorpTokenMaxLifeSpan,
        wechatCorpTokenExpireCriticalSpan,
        wechatCorpTokenBuilder,
        wechatCorpTokenRequestor,
        wechatCorpTokenSQLParamBuilder,
        codeName, args...)
}
