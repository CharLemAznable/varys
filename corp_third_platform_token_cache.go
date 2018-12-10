package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatCorpThirdPlatformConfigLifeSpan = time.Minute * 60  // config cache 60 min default
var wechatCorpThirdPlatformCryptorLifeSpan = time.Minute * 60 // cryptor cache 60 min default

var wechatCorpThirdPlatformConfigCache *gcache.CacheTable
var wechatCorpThirdPlatformCryptorCache *gcache.CacheTable

func wechatCorpThirdPlatformAuthorizerTokenInitialize(configMap map[string]string) {

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

    wechatCorpThirdPlatformConfigCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformConfig")
    wechatCorpThirdPlatformConfigCache.SetDataLoader(wechatCorpThirdPlatformConfigLoader)
    wechatCorpThirdPlatformCryptorCache = gcache.CacheExpireAfterWrite("wechatCorpThirdPlatformCryptor")
    wechatCorpThirdPlatformCryptorCache.SetDataLoader(wechatCorpThirdPlatformCryptorLoader)
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
