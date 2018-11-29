package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    "github.com/CharLemAznable/gql"
    log "github.com/CharLemAznable/log4go"
    "os"
    "time"
)

var db *gql.Gql

func load() {
    log.LoadConfiguration("logback.xml")

    // init db config
    gql.LoadConfigFile("gql.yaml")
    _db, err := gql.Default()
    if nil != err {
        log.Error("Missing db config: Default in gql.yaml")
        os.Exit(-1)
    }
    db = _db

    // query app config -> map[string]string
    configs, err := db.Sql(queryConfigurationSQL).Query()
    if nil != err {
        log.Error("Query Configuration Err: %s", err.Error())
        os.Exit(-1)
    }
    configMap := make(map[string]string)
    for _, config := range configs {
        name := config["CONFIG_NAME"]
        value := config["CONFIG_VALUE"]
        if 0 != len(name) && 0 != len(value) {
            configMap[name] = value
        }
    }

    // load app config
    urlConfigLoader(configMap["wechatAPITokenURL"],
        func(configURL string) {
            wechatAPITokenURL = configURL
        })
    urlConfigLoader(configMap["wechatThirdPlatformTokenURL"],
        func(configURL string) {
            wechatThirdPlatformTokenURL = configURL
        })
    urlConfigLoader(configMap["wechatThirdPlatformPreAuthCodeURL"],
        func(configURL string) {
            wechatThirdPlatformPreAuthCodeURL = configURL
        })
    urlConfigLoader(configMap["wechatThirdPlatformQueryAuthURL"],
        func(configURL string) {
            wechatThirdPlatformQueryAuthURL = configURL
        })
    urlConfigLoader(configMap["wechatThirdPlatformRefreshAuthURL"],
        func(configURL string) {
            wechatThirdPlatformRefreshAuthURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatAPITokenConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatAPITokenConfigLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatAPITokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAPITokenLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatAPITokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAPITokenTempLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformConfigLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformCryptorLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformCryptorLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformTokenLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformTokenTempLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["WechatThirdPlatformPreAuthCodeLifeSpan"],
        func(configVal time.Duration) {
            WechatThirdPlatformPreAuthCodeLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["WechatThirdPlatformPreAuthCodeTempLifeSpan"],
        func(configVal time.Duration) {
            WechatThirdPlatformPreAuthCodeTempLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformAuthorizerTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformAuthorizerTokenLifeSpan = configVal
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformAuthorizerTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformAuthorizerTokenTempLifeSpan = configVal
        })

    // build cache loader
    wechatAPITokenConfigCache = gcache.CacheExpireAfterWrite("WechatAPITokenConfig")
    wechatAPITokenConfigCache.SetDataLoader(wechatAPITokenConfigLoader)
    wechatAPITokenCache = gcache.CacheExpireAfterWrite("wechatAPIToken")
    wechatAPITokenCache.SetDataLoader(wechatAPITokenLoader)
    wechatThirdPlatformConfigCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformConfig")
    wechatThirdPlatformConfigCache.SetDataLoader(wechatThirdPlatformConfigLoader)
    wechatThirdPlatformCryptorCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformCryptor")
    wechatThirdPlatformCryptorCache.SetDataLoader(wechatThirdPlatformCryptorLoader)
    wechatThirdPlatformTokenCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformToken")
    wechatThirdPlatformTokenCache.SetDataLoader(wechatThirdPlatformTokenLoader)
    wechatThirdPlatformPreAuthCodeCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformPreAuthCode")
    wechatThirdPlatformPreAuthCodeCache.SetDataLoader(wechatThirdPlatformPreAuthCodeLoader)
    wechatThirdPlatformAuthorizerTokenCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformAuthorizerToken")
    wechatThirdPlatformAuthorizerTokenCache.SetDataLoader(wechatThirdPlatformAuthorizerTokenLoader)
}

func urlConfigLoader(configStr string, loader func(configURL string)) {
    If(0 != len(configStr), func() {
        loader(configStr)
    })
}

func lifeSpanConfigLoader(configStr string, loader func(configVal time.Duration)) {
    If(0 != len(configStr), func() {
        lifeSpan, err := Int64FromStr(configStr)
        if nil == err {
            loader(time.Minute * time.Duration(lifeSpan))
        }
    })
}
