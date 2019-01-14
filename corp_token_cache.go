package varys

import (
    . "github.com/CharLemAznable/gokits"
    "time"
)

var wechatCorpConfigLifeSpan = time.Minute * 60         // config cache 60 min default
var wechatCorpTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

var wechatCorpConfigCache *CacheTable
var wechatCorpTokenCache *CacheTable

func wechatCorpTokenInitialize(configMap map[string]string) {
    urlConfigLoader(configMap["wechatCorpTokenURL"],
        func(configURL string) {
            wechatCorpTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatCorpConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpConfigLifeSpan = configVal * time.Minute
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

    wechatCorpConfigCache = CacheExpireAfterWrite("wechatCorpConfig")
    wechatCorpConfigCache.SetDataLoader(wechatCorpTokenConfigLoader)
    wechatCorpTokenCache = CacheExpireAfterWrite("wechatCorpToken")
    wechatCorpTokenCache.SetDataLoader(wechatCorpTokenLoader)
}

type WechatCorpConfig struct {
    CorpId     string
    CorpSecret string
}

func wechatCorpTokenConfigLoader(codeName interface{}, args ...interface{}) (*CacheItem, error) {
    return configLoader(
        "WechatCorpConfig",
        queryWechatCorpConfigSQL,
        wechatCorpConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatCorpConfig)
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

func wechatCorpTokenLoader(codeName interface{}, args ...interface{}) (*CacheItem, error) {
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
