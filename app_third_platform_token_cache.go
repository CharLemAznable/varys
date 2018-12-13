package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatAppThirdPlatformConfigLifeSpan = time.Minute * 60             // config cache 60 min default
var wechatAppThirdPlatformCryptorLifeSpan = time.Minute * 60            // cryptor cache 60 min default
var wechatAppThirdPlatformTokenLifeSpan = time.Minute * 5               // stable component token cache 5 min default
var wechatAppThirdPlatformTokenTempLifeSpan = time.Minute * 1           // temporary component token cache 1 min default
var wechatAppThirdPlatformAuthorizerTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatAppThirdPlatformAuthorizerTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

var wechatAppThirdPlatformConfigCache *gcache.CacheTable
var wechatAppThirdPlatformCryptorCache *gcache.CacheTable
var wechatAppThirdPlatformTokenCache *gcache.CacheTable
var wechatAppThirdPlatformAuthorizerTokenCache *gcache.CacheTable

func wechatAppThirdPlatformAuthorizerTokenInitialize(configMap map[string]string) {
    urlConfigLoader(configMap["wechatAppThirdPlatformTokenURL"],
        func(configURL string) {
            wechatAppThirdPlatformTokenURL = configURL
        })
    urlConfigLoader(configMap["wechatAppThirdPlatformPreAuthCodeURL"],
        func(configURL string) {
            wechatAppThirdPlatformPreAuthCodeURL = configURL
        })
    urlConfigLoader(configMap["wechatAppThirdPlatformQueryAuthURL"],
        func(configURL string) {
            wechatAppThirdPlatformQueryAuthURL = configURL
        })
    urlConfigLoader(configMap["wechatAppThirdPlatformRefreshAuthURL"],
        func(configURL string) {
            wechatAppThirdPlatformRefreshAuthURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformCryptorLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformCryptorLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformTokenTempLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformAuthorizerTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformAuthorizerTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformAuthorizerTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformAuthorizerTokenTempLifeSpan = configVal * time.Minute
        })

    wechatAppThirdPlatformConfigCache = gcache.CacheExpireAfterWrite("wechatAppThirdPlatformConfig")
    wechatAppThirdPlatformConfigCache.SetDataLoader(wechatAppThirdPlatformConfigLoader)
    wechatAppThirdPlatformCryptorCache = gcache.CacheExpireAfterWrite("wechatAppThirdPlatformCryptor")
    wechatAppThirdPlatformCryptorCache.SetDataLoader(wechatAppThirdPlatformCryptorLoader)
    wechatAppThirdPlatformTokenCache = gcache.CacheExpireAfterWrite("wechatAppThirdPlatformToken")
    wechatAppThirdPlatformTokenCache.SetDataLoader(wechatAppThirdPlatformTokenLoader)
    wechatAppThirdPlatformAuthorizerTokenCache = gcache.CacheExpireAfterWrite("wechatAppThirdPlatformAuthorizerToken")
    wechatAppThirdPlatformAuthorizerTokenCache.SetDataLoader(wechatAppThirdPlatformAuthorizerTokenLoader)
}

type WechatAppThirdPlatformConfig struct {
    AppId       string
    AppSecret   string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatAppThirdPlatformConfigLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return configLoader(
        "WechatAppThirdPlatformConfig",
        queryWechatAppThirdPlatformConfigSQL,
        wechatAppThirdPlatformConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatAppThirdPlatformConfig)
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

func wechatAppThirdPlatformCryptorLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    cache, err := wechatAppThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, &UnexpectedError{Message:
        "Require WechatAppThirdPlatformConfig with key: " + codeName.(string)} // require config
    }
    config := cache.Data().(*WechatAppThirdPlatformConfig)
    log.Trace("Query WechatAppThirdPlatformConfig Cache:(%s) %s", codeName, Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    log.Info("Load WechatAppThirdPlatformCryptor Cache:(%s) %s", codeName, cryptor)
    return gcache.NewCacheItem(codeName, wechatAppThirdPlatformCryptorLifeSpan, cryptor), nil
}

type WechatAppThirdPlatformToken struct {
    AppId       string
    AccessToken string
}

func wechatAppThirdPlatformTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatAppThirdPlatformToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
}

func wechatAppThirdPlatformTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

func wechatAppThirdPlatformTokenLoader(codeName interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoader(
        "WechatAppThirdPlatformToken",
        queryWechatAppThirdPlatformTokenSQL,
        createWechatAppThirdPlatformTokenSQL,
        updateWechatAppThirdPlatformTokenSQL,
        uncompleteWechatAppThirdPlatformTokenSQL,
        completeWechatAppThirdPlatformTokenSQL,
        wechatAppThirdPlatformTokenLifeSpan,
        wechatAppThirdPlatformTokenTempLifeSpan,
        wechatAppThirdPlatformTokenBuilder,
        wechatAppThirdPlatformTokenRequestor,
        wechatAppThirdPlatformTokenCompleteParamBuilder,
        codeName, args...)
}

type WechatAppThirdPlatformAuthorizerKey struct {
    CodeName        string
    AuthorizerAppId string
}

type WechatAppThirdPlatformAuthorizerToken struct {
    AppId                 string
    AuthorizerAppId       string
    AuthorizerAccessToken string
}

func wechatAppThirdPlatformAuthorizerTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatAppThirdPlatformAuthorizerToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AuthorizerAppId = resultItem["AUTHORIZER_APPID"]
    tokenItem.AuthorizerAccessToken = resultItem["AUTHORIZER_ACCESS_TOKEN"]
    return tokenItem
}

func wechatAppThirdPlatformAuthorizerTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, codeName interface{}, authorizerAppId interface{}) []interface{} {
    expiresIn, _ := IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{
        resultItem["AUTHORIZER_ACCESS_TOKEN"],
        resultItem["AUTHORIZER_REFRESH_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), codeName, authorizerAppId}
}

func wechatAppThirdPlatformAuthorizerTokenLoader(key interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    tokenKey, ok := key.(WechatAppThirdPlatformAuthorizerKey)
    if !ok {
        return nil, &UnexpectedError{Message:
        "WechatAppThirdPlatformAuthorizerKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatAppThirdPlatformAuthorizerTokenSQL).
        Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
        "Unauthorized authorizer: " + Json(key)}).(error) // requires that the token already exists
    }
    log.Trace("Query WechatAppThirdPlatformAuthorizerToken:(%s) %s", Json(key), resultMap)

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
        log.Info("Try to request and update WechatAppThirdPlatformAuthorizerToken:(%s)", Json(key))
        count, err := db.New().Sql(updateWechatAppThirdPlatformAuthorizerTokenSQL).
            Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
        if nil == err && count > 0 {
            resultItem, err := wechatAppThirdPlatformRefreshAuthRequestor(
                tokenKey.CodeName, tokenKey.AuthorizerAppId, authorizerRefreshToken)
            if nil != err {
                db.New().Sql(uncompleteWechatAppThirdPlatformAuthorizerTokenSQL).
                    Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
                return nil, err
            }
            count, err := db.New().Sql(completeWechatAppThirdPlatformAuthorizerTokenSQL).
                Params(wechatAppThirdPlatformAuthorizerTokenCompleteParamBuilder(
                    resultItem, wechatAppThirdPlatformAuthorizerTokenLifeSpan,
                    tokenKey.CodeName, tokenKey.AuthorizerAppId)...).Execute()
            if nil != err || count < 1 {
                return nil, DefaultIfNil(err, &UnexpectedError{Message:
                "Replace WechatAppThirdPlatformAuthorizerToken Failed"}).(error)
            }

            tokenItem := wechatAppThirdPlatformAuthorizerTokenBuilder(resultItem)
            log.Info("Request WechatAppThirdPlatformAuthorizerToken:(%s) %s", Json(key), Json(tokenItem))
            return gcache.NewCacheItem(key, wechatAppThirdPlatformAuthorizerTokenLifeSpan, tokenItem), nil
        }
        log.Warn("Give up request and update WechatAppThirdPlatformAuthorizerToken:(%s), use Query result Temporarily", Json(key))
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := Condition(isExpired, wechatAppThirdPlatformAuthorizerTokenTempLifeSpan,
        wechatAppThirdPlatformAuthorizerTokenLifeSpan).(time.Duration)
    tokenItem := wechatAppThirdPlatformAuthorizerTokenBuilder(resultItem)
    log.Info("Load WechatAppThirdPlatformAuthorizerToken Cache:(%s) %s, cache %3.1f min", Json(key), Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(key, ls, tokenItem), nil
}

func wechatAppThirdPlatformAuthorizerTokenCreator(codeName, authorizerAppId, authorizationCode interface{}) {
    count, err := db.New().Sql(createWechatAppThirdPlatformAuthorizerTokenSQL).
        Params(codeName, authorizerAppId).Execute()
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        // 强制更新记录, 不论token是否已过期
        count, err = db.New().Sql(updateWechatAppThirdPlatformAuthorizerTokenForceSQL).
            Params(codeName, authorizerAppId).Execute()
        for nil != err || count < 1 { // 尝试更新记录失败, 则休眠(1 sec)后再次尝试更新记录
            time.Sleep(1 * time.Second)
            count, err = db.New().Sql(updateWechatAppThirdPlatformAuthorizerTokenForceSQL).
                Params(codeName, authorizerAppId).Execute()
        }
    }

    // 锁定成功, 开始更新
    resultItem, err := wechatAppThirdPlatformQueryAuthRequestor(codeName, authorizationCode)
    if nil != err {
        db.New().Sql(uncompleteWechatAppThirdPlatformAuthorizerTokenSQL).
            Params(codeName, authorizerAppId).Execute()
        log.Warn("Request WechatAppThirdPlatformAuthorizerToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, err.Error())
        return
    }
    count, err = db.New().Sql(completeWechatAppThirdPlatformAuthorizerTokenSQL).
        Params(wechatAppThirdPlatformAuthorizerTokenCompleteParamBuilder(
            resultItem, wechatAppThirdPlatformAuthorizerTokenLifeSpan,
            codeName, authorizerAppId)...).Execute()
    if nil != err || count < 1 {
        log.Warn("Record new WechatAppThirdPlatformAuthorizerToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, DefaultIfNil(err, &UnexpectedError{Message:
            "Replace WechatAppThirdPlatformAuthorizerToken Failed"}).(error).Error())
        return
    }

    tokenItem := wechatAppThirdPlatformAuthorizerTokenBuilder(resultItem)
    log.Info("Request WechatAppThirdPlatformAuthorizerToken:(%s, %s) %s",
        codeName, authorizerAppId, Json(tokenItem))
    wechatAppThirdPlatformAuthorizerTokenCache.Add(WechatAppThirdPlatformAuthorizerKey{
        CodeName: codeName.(string), AuthorizerAppId: authorizerAppId.(string)},
        wechatAppThirdPlatformAuthorizerTokenLifeSpan, tokenItem)
}
