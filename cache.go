package varys

import (
    "github.com/CharLemAznable/gcache"
    "github.com/CharLemAznable/wechataes"
    "log"
    "time"
)

var wechatAPITokenConfigLifeSpan = time.Minute * 60                  // config cache 60 min default
var wechatAPITokenLifeSpan = time.Minute * 5                         // stable token cache 5 min default
var wechatAPITokenTempLifeSpan = time.Minute * 1                     // temporary token cache 1 min default
var wechatThirdPlatformConfigLifeSpan = time.Minute * 60             // config cache 60 min default
var wechatThirdPlatformCryptorLifeSpan = time.Minute * 60            // cryptor cache 60 min default
var wechatThirdPlatformTokenLifeSpan = time.Minute * 5               // stable component token cache 5 min default
var wechatThirdPlatformTokenTempLifeSpan = time.Minute * 1           // temporary component token cache 1 min default
var WechatThirdPlatformPreAuthCodeLifeSpan = time.Minute * 3         // stable pre-auth code cache 3 min default
var WechatThirdPlatformPreAuthCodeTempLifeSpan = time.Minute * 1     // temporary pre-auth code cache 1 min default
var wechatThirdPlatformAuthorizerTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatThirdPlatformAuthorizerTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

var wechatAPITokenConfigCache *gcache.CacheTable
var wechatAPITokenCache *gcache.CacheTable
var wechatThirdPlatformConfigCache *gcache.CacheTable
var wechatThirdPlatformCryptorCache *gcache.CacheTable
var wechatThirdPlatformTokenCache *gcache.CacheTable
var wechatThirdPlatformPreAuthCodeCache *gcache.CacheTable
var wechatThirdPlatformAuthorizerTokenCache *gcache.CacheTable

// common loader

func configLoader(
    name string,
    sql string,
    lifeSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    key interface{},
    args ...interface{}) (*gcache.CacheItem, error) {

    resultMap, err := db.Sql(sql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, &UnexpectedError{Message:
        "Require " + name + " Config with key: " + key.(string)} // require config
    }
    log.Printf("Query %s: %s", name, resultMap)

    config := builder(resultMap[0])
    log.Printf("Load %s Cache: %s", name, Json(config))
    return gcache.NewCacheItem(key, lifeSpan, config), nil
}

func tokenLoader(
    name string,
    querySql string,
    createSql string,
    updateSql string,
    uncompleteSql string,
    completeSql string,
    lifeSpan time.Duration,
    lifeSpanTemp time.Duration,
    builder func(key interface{}, token string) interface{},
    requestor func(key interface{}) (string, int, error),
    key interface{},
    args ...interface{}) (*gcache.CacheItem, error) {

    resultMap, err := db.Sql(querySql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        log.Printf("Try to request %s", name)
        count, err := db.Sql(createSql).Params(key).Execute()
        if nil == err && count > 0 {
            token, err := requestUpdater(
                name, uncompleteSql, completeSql, lifeSpan, requestor, key, args...)
            if nil != err {
                return nil, err
            }
            tokenItem := builder(key, token)
            log.Printf("Request %s: %s", name, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem), nil
        }
        log.Printf("Give up request %s, wait for next cache Query", name)
        return nil, &UnexpectedError{Message: "Query " + name + " Later"}
    }
    log.Printf("Query %s: %s", name, resultMap)

    resultItem := resultMap[0]
    updated := resultItem["UPDATED"]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err {
        return nil, err
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        log.Printf("Try to request and update %s", name)
        count, err := db.Sql(updateSql).Params(key).Execute()
        if nil == err && count > 0 {
            token, err := requestUpdater(
                name, uncompleteSql, completeSql, lifeSpan, requestor, key, args...)
            if nil != err {
                return nil, err
            }
            tokenItem := builder(key, token)
            log.Printf("Request %s: %s", name, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem), nil
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
    return gcache.NewCacheItem(key, ls, tokenItem), nil
}

func requestUpdater(
    name string,
    uncompleteSql string,
    completeSql string,
    lifeSpan time.Duration,
    requestor func(key interface{}) (string, int, error),
    key interface{},
    args ...interface{}) (string, error) {

    token, expiresIn, err := requestor(key)
    if nil != err {
        db.Sql(uncompleteSql).Params(key).Execute()
        return "", err
    }
    // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    expireTimeInc := expiresIn - int(lifeSpan.Seconds()*1.1)
    count, err := db.Sql(completeSql).Params(token, expireTimeInc, key).Execute()
    if nil != err {
        return "", err
    }
    if count < 1 {
        return "", &UnexpectedError{Message: "Record new " + name + " Failed"}
    }

    return token, nil
}

// Wechat access_token cache loader

type WechatAPITokenConfig struct {
    AppId     string
    AppSecret string
}

func wechatAPITokenConfigLoader(appId interface{}, args ...interface{}) (*gcache.CacheItem, error) {
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

func wechatAPITokenLoader(appId interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoader(
        "WechatAPIToken",
        queryWechatAPITokenSQL,
        createWechatAPITokenUpdating,
        updateWechatAPITokenUpdating,
        uncompleteWechatAPITokenSQL,
        completeWechatAPITokenSQL,
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
    AppId       string
    AppSecret   string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatThirdPlatformConfigLoader(appId interface{}, args ...interface{}) (*gcache.CacheItem, error) {
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
        appId, args...)
}

func wechatThirdPlatformCryptorLoader(appId interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    cache, err := wechatThirdPlatformConfigCache.Value(appId)
    if nil != err {
        return nil, &UnexpectedError{Message:
        "Require WechatThirdPlatformConfig with key: " + appId.(string)} // require config
    }
    config := cache.Data().(*WechatThirdPlatformConfig)
    log.Printf("Query WechatThirdPlatformConfig Cache: %s", Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    log.Printf("Load WechatThirdPlatformCryptor Cache: %s", Json(cryptor))
    return gcache.NewCacheItem(appId, wechatThirdPlatformCryptorLifeSpan, cryptor), nil
}

type WechatThirdPlatformToken struct {
    AppId                string
    ComponentAccessToken string
}

func wechatThirdPlatformTokenLoader(appId interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoader(
        "WechatThirdPlatformToken",
        queryWechatThirdPlatformTokenSQL,
        createWechatThirdPlatformTokenUpdating,
        updateWechatThirdPlatformTokenUpdating,
        uncompleteWechatThirdPlatformTokenSQL,
        completeWechatThirdPlatformTokenSQL,
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

func wechatThirdPlatformPreAuthCodeLoader(appId interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    return tokenLoader(
        "WechatThirdPlatformPreAuthCode",
        queryWechatThirdPlatformPreAuthCodeSQL,
        createWechatThirdPlatformPreAuthCodeUpdating,
        updateWechatThirdPlatformPreAuthCodeUpdating,
        uncompleteWechatThirdPlatformPreAuthCodeSQL,
        completeWechatThirdPlatformPreAuthCodeSQL,
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

type WechatThirdPlatformAuthorizerTokenKey struct {
    AppId           string
    AuthorizerAppId string
}

type WechatThirdPlatformAuthorizerToken struct {
    AppId                  string
    AuthorizerAppId        string
    AuthorizerAccessToken  string
}

func wechatThirdPlatformAuthorizerTokenLoader(key interface{}, args ...interface{}) (*gcache.CacheItem, error) {
    tokenKey, ok := key.(WechatThirdPlatformAuthorizerTokenKey)
    if !ok {
        return nil, &UnexpectedError{Message:
        "WechatThirdPlatformAuthorizerTokenKey type error"} // key type error
    }

    resultMap, err := db.Sql(queryWechatThirdPlatformAuthorizerTokenSQL).
        Params(tokenKey.AppId, tokenKey.AuthorizerAppId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
        "Unauthorized authorizer app_id"}).(error) // requires that the token already exists
    }
    log.Printf("Query WechatThirdPlatformAuthorizerToken: %s", resultMap)

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
        log.Printf("Try to request and update WechatThirdPlatformAuthorizerToken")
        count, err := db.Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
            Params(tokenKey.AppId, tokenKey.AuthorizerAppId).Execute()
        if nil == err && count > 0 {
            accessToken, refreshToken, expiresIn, err := wechatThirdPlatformRefreshAuthRequestor(
                tokenKey.AppId, tokenKey.AuthorizerAppId, authorizerRefreshToken)
            if nil != err {
                db.Sql(uncompleteWechatThirdPlatformAuthorizerTokenSQL).
                    Params(tokenKey.AppId, tokenKey.AuthorizerAppId).Execute()
                return nil, err
            }
            // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
            expireTimeInc := expiresIn - int(wechatThirdPlatformAuthorizerTokenLifeSpan.Seconds()*1.1)
            count, err := db.Sql(completeWechatThirdPlatformAuthorizerTokenSQL).
                Params(accessToken, refreshToken, expireTimeInc, tokenKey.AppId, tokenKey.AuthorizerAppId).Execute()
            if nil != err || count < 1 {
                return nil, DefaultIfNil(err, &UnexpectedError{Message:
                "Replace WechatThirdPlatformAuthorizerToken Failed"}).(error)
            }

            tokenItem := new(WechatThirdPlatformAuthorizerToken)
            tokenItem.AppId = tokenKey.AppId
            tokenItem.AuthorizerAppId = tokenKey.AuthorizerAppId
            tokenItem.AuthorizerAccessToken = accessToken
            log.Printf("Request WechatThirdPlatformAuthorizerToken: %s", Json(tokenItem))
            return gcache.NewCacheItem(key, wechatThirdPlatformAuthorizerTokenLifeSpan, tokenItem), nil
        }
        log.Printf("Give up request and update WechatThirdPlatformAuthorizerToken, use Query result Temporarily")
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := Condition(isExpired, wechatThirdPlatformAuthorizerTokenTempLifeSpan,
        wechatThirdPlatformAuthorizerTokenLifeSpan).(time.Duration)
    tokenItem := new(WechatThirdPlatformAuthorizerToken)
    tokenItem.AppId = tokenKey.AppId
    tokenItem.AuthorizerAppId = tokenKey.AuthorizerAppId
    tokenItem.AuthorizerAccessToken = resultItem["AUTHORIZER_ACCESS_TOKEN"]
    log.Printf("Load WechatThirdPlatformAuthorizerToken Cache: %s, cache %3.1f min",
        Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(key, ls, tokenItem), nil
}

func wechatThirdPlatformAuthorizerTokenCreator(appId, authorizerAppid, authorizationCode interface{}) {
    count, err := db.Sql(createWechatThirdPlatformAuthorizerTokenUpdating).
        Params(appId, authorizerAppid).Execute()
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        count, err = db.Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
            Params(appId, authorizerAppid).Execute()
        for nil != err || count < 1 { // 尝试更新记录失败, 则休眠1sec后再次尝试更新记录
            time.Sleep(1 * time.Second)
            count, err = db.Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
                Params(appId, authorizerAppid).Execute()
        }
    }

    // 锁定成功, 开始更新
    accessToken, refreshToken, expiresIn, err :=
        wechatThirdPlatformQueryAuthRequestor(appId, authorizationCode)
    if nil != err {
        db.Sql(uncompleteWechatThirdPlatformAuthorizerTokenSQL).
            Params(appId, authorizerAppid).Execute()
        log.Printf("Request WechatThirdPlatformAuthorizerToken Failed: %s", err.Error())
    }
    // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    expireTimeInc := expiresIn - int(wechatThirdPlatformAuthorizerTokenLifeSpan.Seconds()*1.1)
    count, err = db.Sql(completeWechatThirdPlatformAuthorizerTokenSQL).
        Params(appId, authorizerAppid, accessToken, refreshToken, expireTimeInc).Execute()
    if nil != err || count < 1 {
        log.Printf("Record new WechatThirdPlatformAuthorizerToken Failed: %s",
            DefaultIfNil(err, &UnexpectedError{Message:
            "Replace WechatThirdPlatformAuthorizerToken Failed"}).(error).Error())
    }

    tokenItem := new(WechatThirdPlatformAuthorizerToken)
    tokenItem.AppId = appId.(string)
    tokenItem.AuthorizerAppId = authorizerAppid.(string)
    tokenItem.AuthorizerAccessToken = accessToken
    log.Printf("Request WechatThirdPlatformAuthorizerToken: %s", Json(tokenItem))
    wechatThirdPlatformAuthorizerTokenCache.Add(WechatThirdPlatformAuthorizerTokenKey{
        AppId: appId.(string), AuthorizerAppId: authorizerAppid.(string)},
        wechatThirdPlatformAuthorizerTokenLifeSpan, tokenItem)
}
