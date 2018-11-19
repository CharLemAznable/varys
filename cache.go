package varys

import (
    "github.com/CharLemAznable/gcache"
    "github.com/CharLemAznable/wechataes"
    "log"
    "time"
)

var wechatAPITokenConfigLifeSpan = time.Minute * 60              // config cache 60 min default
var wechatAPITokenLifeSpan = time.Minute * 5                     // stable token cache 5 min default
var wechatAPITokenTempLifeSpan = time.Minute * 1                 // temporary token cache 1 min default
var wechatThirdPlatformConfigLifeSpan = time.Minute * 60         // config cache 60 min default
var wechatThirdPlatformCryptorLifeSpan = time.Minute * 60        // cryptor cache 60 min default
var wechatThirdPlatformTokenLifeSpan = time.Minute * 5           // stable component token cache 5 min default
var wechatThirdPlatformTokenTempLifeSpan = time.Minute * 1       // temporary component token cache 1 min default
var WechatThirdPlatformPreAuthCodeLifeSpan = time.Minute * 3     // stable pre-auth code cache 3 min default
var WechatThirdPlatformPreAuthCodeTempLifeSpan = time.Minute * 1 // temporary pre-auth code cache 1 min default

var wechatAPITokenConfigCache *gcache.CacheTable
var wechatAPITokenCache *gcache.CacheTable
var wechatThirdPlatformConfigCache *gcache.CacheTable
var wechatThirdPlatformCryptorCache *gcache.CacheTable
var wechatThirdPlatformTokenCache *gcache.CacheTable
var wechatThirdPlatformPreAuthCodeCache *gcache.CacheTable

// common loader

func configLoader(
    name string,
    sql string,
    lifeSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    key interface{},
    args ...interface{}) *gcache.CacheItem {

    resultMap, err := db.Sql(sql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        return nil // require config
    }
    log.Printf("Query %s: %s", name, resultMap)

    config := builder(resultMap[0])
    if nil == config {
        return nil
    }
    log.Printf("Load %s Cache: %s", name, Json(config))
    return gcache.NewCacheItem(key, lifeSpan, config)
}

func tokenLoader(
    name string,
    querySql string,
    createSql string,
    updateSql string,
    replaceSql string,
    lifeSpan time.Duration,
    lifeSpanTemp time.Duration,
    builder func(key interface{}, token string) interface{},
    requestor func(key interface{}) (string, int, error),
    key interface{},
    args ...interface{}) *gcache.CacheItem {

    resultMap, err := db.Sql(querySql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        log.Printf("Try to request %s", name)
        count, err := db.Sql(createSql).Params(key).Execute()
        if nil == err && count > 0 {
            token, err := requestReplacer(
                name, replaceSql, lifeSpan, requestor, key, args...)
            if nil != err {
                return nil
            }
            tokenItem := builder(key, token)
            log.Printf("Request %s: %s", name, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem)
        }
        log.Printf("Give up request %s, wait for next cache Query", name)
        return nil
    }
    log.Printf("Query %s: %s", name, resultMap)

    resultItem := resultMap[0]
    updated := resultItem["UPDATED"]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err {
        return nil
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        log.Printf("Try to request and update %s", name)
        count, err := db.Sql(updateSql).Params(key).Execute()
        if nil == err && count > 0 {
            token, err := requestReplacer(
                name, replaceSql, lifeSpan, requestor, key, args...)
            if nil != err {
                return nil
            }
            tokenItem := builder(key, token)
            log.Printf("Request %s: %s", name, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem)
        }
        log.Printf("Give up request and update %s, use Query result Temporarily", name)
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := Condition(isExpired, lifeSpanTemp, lifeSpan).(time.Duration)
    tokenItem := builder(key, resultItem["TOKEN"])
    log.Printf("Load %s Cache: %s, cache %3.1f min", name, Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(key, ls, tokenItem)
}

func requestReplacer(
    name string,
    replaceSql string,
    lifeSpan time.Duration,
    requestor func(key interface{}) (string, int, error),
    key interface{},
    args ...interface{}) (string, error) {

    token, expiresIn, err := requestor(key)
    if nil != err {
        return "", err
    }
    // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    expireTimeInc := expiresIn - int(lifeSpan.Seconds()*1.1)
    count, err := db.Sql(replaceSql).Params(key, token, expireTimeInc).Execute()
    if nil != err {
        return "", err
    }
    if count < 1 {
        return "", &UnexpectedError{Message: "Replace " + name + " Failed"}
    }

    return token, nil
}

// Wechat access_token cache loader

type WechatAPITokenConfig struct {
    AppId     string
    AppSecret string
}

func wechatAPITokenConfigLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
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
        appId, args...)
}

type WechatAPIToken struct {
    AppId       string
    AccessToken string
}

func wechatAPITokenLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
    return tokenLoader(
        "WechatAPIToken",
        queryWechatAPITokenSQL,
        createWechatAPITokenUpdating,
        updateWechatAPITokenUpdating,
        replaceWechatAPITokenSQL,
        wechatAPITokenLifeSpan,
        wechatAPITokenTempLifeSpan,
        func(key interface{}, token string) interface{} {
            tokenItem := new(WechatAPIToken)
            tokenItem.AppId = key.(string)
            tokenItem.AccessToken = token
            return tokenItem
        },
        wechatAPITokenRequestor,
        appId, args...)
}

// Wechat third-platform component_access_token cache loader

type WechatThirdPlatformConfig struct {
    AppId     string
    AppSecret string
    Token     string
    AesKey    string
}

func wechatThirdPlatformConfigLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
    return configLoader(
        "WechatThirdPlatformConfig",
        queryWechatThirdPlatformConfigSQL,
        wechatThirdPlatformConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatThirdPlatformConfig)
            config.AppId = resultItem["APP_ID"]
            config.AppSecret = resultItem["APP_SECRET"]
            config.Token = resultItem["TOKEN"]
            config.AesKey = resultItem["AES_KEY"]
            if 0 == len(config.AppId) || 0 == len(config.AppSecret) ||
                0 == len(config.Token) || 0 == len(config.AesKey) {
                return nil
            }
            return config
        },
        appId, args...)
}

func wechatThirdPlatformCryptorLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
    cache, err := wechatThirdPlatformConfigCache.Value(appId)
    if nil != err {
        return nil // require config
    }
    config := cache.Data().(*WechatThirdPlatformConfig)
    log.Printf("Query WechatThirdPlatformConfig Cache: %s", Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil // require legal config
    }
    log.Printf("Load WechatThirdPlatformCryptor Cache: %s", Json(cryptor))
    return gcache.NewCacheItem(appId, wechatThirdPlatformCryptorLifeSpan, cryptor)
}

type WechatThirdPlatformToken struct {
    AppId                string
    ComponentAccessToken string
}

func wechatThirdPlatformTokenLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
    return tokenLoader(
        "WechatThirdPlatformToken",
        queryWechatThirdPlatformTokenSQL,
        createWechatThirdPlatformTokenUpdating,
        updateWechatThirdPlatformTokenUpdating,
        replaceWechatThirdPlatformTokenSQL,
        wechatThirdPlatformTokenLifeSpan,
        wechatThirdPlatformTokenTempLifeSpan,
        func(key interface{}, token string) interface{} {
            tokenItem := new(WechatThirdPlatformToken)
            tokenItem.AppId = key.(string)
            tokenItem.ComponentAccessToken = token
            return tokenItem
        },
        wechatThirdPlatformTokenRequestor,
        appId, args...)
}

type WechatThirdPlatformPreAuthCode struct {
    AppId       string
    PreAuthCode string
}

func wechatThirdPlatformPreAuthCodeLoader(appId interface{}, args ...interface{}) *gcache.CacheItem {
    return tokenLoader(
        "WechatThirdPlatformPreAuthCode",
        queryWechatThirdPlatformPreAuthCodeSQL,
        createWechatThirdPlatformPreAuthCodeUpdating,
        updateWechatThirdPlatformPreAuthCodeUpdating,
        replaceWechatThirdPlatformPreAuthCodeSQL,
        WechatThirdPlatformPreAuthCodeLifeSpan,
        WechatThirdPlatformPreAuthCodeTempLifeSpan,
        func(key interface{}, token string) interface{} {
            codeItem := new(WechatThirdPlatformPreAuthCode)
            codeItem.AppId = key.(string)
            codeItem.PreAuthCode = token
            return codeItem
        },
        wechatThirdPlatformPreAuthCodeRequestor,
        appId, args...)
}
