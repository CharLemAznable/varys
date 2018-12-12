package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatThirdPlatformConfigLifeSpan = time.Minute * 60             // config cache 60 min default
var wechatThirdPlatformCryptorLifeSpan = time.Minute * 60            // cryptor cache 60 min default
var wechatThirdPlatformTokenLifeSpan = time.Minute * 5               // stable component token cache 5 min default
var wechatThirdPlatformTokenTempLifeSpan = time.Minute * 1           // temporary component token cache 1 min default
var wechatThirdPlatformAuthorizerTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatThirdPlatformAuthorizerTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

var wechatThirdPlatformConfigCache *gcache.CacheTable
var wechatThirdPlatformCryptorCache *gcache.CacheTable
var wechatThirdPlatformTokenCache *gcache.CacheTable
var wechatThirdPlatformAuthorizerTokenCache *gcache.CacheTable

func wechatThirdPlatformAuthorizerTokenInitialize(configMap map[string]string) {
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
        configMap["wechatThirdPlatformConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformCryptorLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformCryptorLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformTokenTempLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformAuthorizerTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformAuthorizerTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatThirdPlatformAuthorizerTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatThirdPlatformAuthorizerTokenTempLifeSpan = configVal * time.Minute
        })

    wechatThirdPlatformConfigCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformConfig")
    wechatThirdPlatformConfigCache.SetDataLoader(wechatThirdPlatformConfigLoader)
    wechatThirdPlatformCryptorCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformCryptor")
    wechatThirdPlatformCryptorCache.SetDataLoader(wechatThirdPlatformCryptorLoader)
    wechatThirdPlatformTokenCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformToken")
    wechatThirdPlatformTokenCache.SetDataLoader(wechatThirdPlatformTokenLoader)
    wechatThirdPlatformAuthorizerTokenCache = gcache.CacheExpireAfterWrite("wechatThirdPlatformAuthorizerToken")
    wechatThirdPlatformAuthorizerTokenCache.SetDataLoader(wechatThirdPlatformAuthorizerTokenLoader)
}

type WechatThirdPlatformConfig struct {
    AppId       string
    AppSecret   string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatThirdPlatformConfigLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
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
            config.RedirectURL = resultItem["REDIRECT_URL"]
            if 0 == len(config.AppId) || 0 == len(config.AppSecret) ||
                0 == len(config.Token) || 0 == len(config.AesKey) {
                return nil
            }
            return config
        },
        codeName, args...)
}

func wechatThirdPlatformCryptorLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    cache, err := wechatThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, &UnexpectedError{Message:
        "Require WechatThirdPlatformConfig with key: " + codeName.(string)} // require config
    }
    config := cache.Data().(*WechatThirdPlatformConfig)
    log.Trace("Query WechatThirdPlatformConfig Cache:(%s) %s", codeName, Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    log.Info("Load WechatThirdPlatformCryptor Cache:(%s) %s", codeName, cryptor)
    return gcache.NewCacheItem(codeName, wechatThirdPlatformCryptorLifeSpan, cryptor), nil
}

type WechatThirdPlatformToken struct {
    AppId                string
    ComponentAccessToken string
}

func wechatThirdPlatformTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatThirdPlatformToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.ComponentAccessToken = resultItem["COMPONENT_ACCESS_TOKEN"]
    return tokenItem
}

func wechatThirdPlatformTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["COMPONENT_ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

func wechatThirdPlatformTokenLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoader(
        "WechatThirdPlatformToken",
        queryWechatThirdPlatformTokenSQL,
        createWechatThirdPlatformTokenUpdating,
        updateWechatThirdPlatformTokenUpdating,
        uncompleteWechatThirdPlatformTokenSQL,
        completeWechatThirdPlatformTokenSQL,
        wechatThirdPlatformTokenLifeSpan,
        wechatThirdPlatformTokenTempLifeSpan,
        wechatThirdPlatformTokenBuilder,
        wechatThirdPlatformTokenRequestor,
        wechatThirdPlatformTokenCompleteParamBuilder,
        codeName, args...)
}

type WechatThirdPlatformAuthorizerTokenKey struct {
    CodeName        string
    AuthorizerAppId string
}

type WechatThirdPlatformAuthorizerToken struct {
    AppId                 string
    AuthorizerAppId       string
    AuthorizerAccessToken string
}

func wechatThirdPlatformAuthorizerTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatThirdPlatformAuthorizerToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AuthorizerAppId = resultItem["AUTHORIZER_APPID"]
    tokenItem.AuthorizerAccessToken = resultItem["AUTHORIZER_ACCESS_TOKEN"]
    return tokenItem
}

func wechatThirdPlatformAuthorizerTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, codeName interface{}, authorizerAppId interface{}) []interface{} {
    expiresIn, _ := IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{
        resultItem["AUTHORIZER_ACCESS_TOKEN"],
        resultItem["AUTHORIZER_REFRESH_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), codeName, authorizerAppId}
}

func wechatThirdPlatformAuthorizerTokenLoader(key interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    tokenKey, ok := key.(WechatThirdPlatformAuthorizerTokenKey)
    if !ok {
        return nil, &UnexpectedError{Message:
        "WechatThirdPlatformAuthorizerTokenKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatThirdPlatformAuthorizerTokenSQL).
        Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
        "Unauthorized authorizer: " + Json(key)}).(error) // requires that the token already exists
    }
    log.Trace("Query WechatThirdPlatformAuthorizerToken:(%s) %s", Json(key), resultMap)

    resultItem := resultMap[0]
    authorizerRefreshToken := resultItem["AUTHORIZER_REFRESH_TOKEN"] // requires that the refresh token exists
    updated := resultItem["UPDATED"]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err || 0 == len(authorizerRefreshToken) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message: "Refresh token is Empty"}).(error)
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        log.Info("Try to request and update WechatThirdPlatformAuthorizerToken:(%s)", Json(key))
        count, err := db.New().Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
            Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
        if nil == err && count > 0 {
            resultItem, err := wechatThirdPlatformRefreshAuthRequestor(
                tokenKey.CodeName, tokenKey.AuthorizerAppId, authorizerRefreshToken)
            if nil != err {
                db.New().Sql(uncompleteWechatThirdPlatformAuthorizerTokenSQL).
                    Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
                return nil, err
            }
            count, err := db.New().Sql(completeWechatThirdPlatformAuthorizerTokenSQL).
                Params(wechatThirdPlatformAuthorizerTokenCompleteParamBuilder(
                    resultItem, wechatThirdPlatformAuthorizerTokenLifeSpan,
                    tokenKey.CodeName, tokenKey.AuthorizerAppId)...).Execute()
            if nil != err || count < 1 {
                return nil, DefaultIfNil(err, &UnexpectedError{Message:
                "Replace WechatThirdPlatformAuthorizerToken Failed"}).(error)
            }

            tokenItem := wechatThirdPlatformAuthorizerTokenBuilder(resultItem)
            log.Info("Request WechatThirdPlatformAuthorizerToken:(%s) %s", Json(key), Json(tokenItem))
            return gcache.NewCacheItem(key, wechatThirdPlatformAuthorizerTokenLifeSpan, tokenItem), nil
        }
        log.Warn("Give up request and update WechatThirdPlatformAuthorizerToken:(%s), use Query result Temporarily", Json(key))
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := Condition(isExpired, wechatThirdPlatformAuthorizerTokenTempLifeSpan,
        wechatThirdPlatformAuthorizerTokenLifeSpan).(time.Duration)
    tokenItem := wechatThirdPlatformAuthorizerTokenBuilder(resultItem)
    log.Info("Load WechatThirdPlatformAuthorizerToken Cache:(%s) %s, cache %3.1f min", Json(key), Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(key, ls, tokenItem), nil
}

func wechatThirdPlatformAuthorizerTokenCreator(codeName, authorizerAppId, authorizationCode interface{}) {
    count, err := db.New().Sql(createWechatThirdPlatformAuthorizerTokenUpdating).
        Params(codeName, authorizerAppId).Execute()
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        count, err = db.New().Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
            Params(codeName, authorizerAppId).Execute()
        for nil != err || count < 1 { // 尝试更新记录失败, 则休眠(1 sec)后再次尝试更新记录
            time.Sleep(1 * time.Second)
            count, err = db.New().Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
                Params(codeName, authorizerAppId).Execute()
        }
    }

    // 锁定成功, 开始更新
    resultItem, err := wechatThirdPlatformQueryAuthRequestor(codeName, authorizationCode)
    if nil != err {
        db.New().Sql(uncompleteWechatThirdPlatformAuthorizerTokenSQL).
            Params(codeName, authorizerAppId).Execute()
        log.Warn("Request WechatThirdPlatformAuthorizerToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, err.Error())
    }
    count, err = db.New().Sql(completeWechatThirdPlatformAuthorizerTokenSQL).
        Params(wechatThirdPlatformAuthorizerTokenCompleteParamBuilder(
            resultItem, wechatThirdPlatformAuthorizerTokenLifeSpan,
            codeName, authorizerAppId)...).Execute()
    if nil != err || count < 1 {
        log.Warn("Record new WechatThirdPlatformAuthorizerToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, DefaultIfNil(err, &UnexpectedError{Message:
            "Replace WechatThirdPlatformAuthorizerToken Failed"}).(error).Error())
    }

    tokenItem := wechatThirdPlatformAuthorizerTokenBuilder(resultItem)
    log.Info("Request WechatThirdPlatformAuthorizerToken:(%s, %s) %s",
        codeName, authorizerAppId, Json(tokenItem))
    wechatThirdPlatformAuthorizerTokenCache.Add(WechatThirdPlatformAuthorizerTokenKey{
        CodeName: codeName.(string), AuthorizerAppId: authorizerAppId.(string)},
        wechatThirdPlatformAuthorizerTokenLifeSpan, tokenItem)
}