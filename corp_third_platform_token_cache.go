package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatCorpThirdPlatformConfigLifeSpan = time.Minute * 60         // config cache 60 min default
var wechatCorpThirdPlatformCryptorLifeSpan = time.Minute * 60        // cryptor cache 60 min default
var wechatCorpThirdPlatformTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpThirdPlatformTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

var wechatCorpThirdPlatformConfigCache *gcache.CacheTable
var wechatCorpThirdPlatformCryptorCache *gcache.CacheTable
var wechatCorpThirdPlatformTokenCache *gcache.CacheTable

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

    wechatCorpThirdPlatformConfigCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformConfig")
    wechatCorpThirdPlatformConfigCache.SetDataLoader(wechatCorpThirdPlatformConfigLoader)
    wechatCorpThirdPlatformCryptorCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformCryptor")
    wechatCorpThirdPlatformCryptorCache.SetDataLoader(wechatCorpThirdPlatformCryptorLoader)
    wechatCorpThirdPlatformTokenCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformCryptor")
    wechatCorpThirdPlatformTokenCache.SetDataLoader(wechatCorpThirdPlatformTokenLoader)
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
