package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatAPITokenConfigLifeSpan = time.Minute * 60                  // config cache 60 min default
var wechatAPITokenLifeSpan = time.Minute * 5                         // stable token cache 5 min default
var wechatAPITokenTempLifeSpan = time.Minute * 1                     // temporary token cache 1 min default
var wechatThirdPlatformConfigLifeSpan = time.Minute * 60             // config cache 60 min default
var wechatThirdPlatformCryptorLifeSpan = time.Minute * 60            // cryptor cache 60 min default
var wechatThirdPlatformTokenLifeSpan = time.Minute * 5               // stable component token cache 5 min default
var wechatThirdPlatformTokenTempLifeSpan = time.Minute * 1           // temporary component token cache 1 min default
var wechatThirdPlatformAuthorizerTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatThirdPlatformAuthorizerTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default
var wechatCorpTokenConfigLifeSpan = time.Minute * 60                 // config cache 60 min default
var wechatCorpTokenMaxLifeSpan = time.Minute * 5                     // stable token cache 5 min max
var wechatCorpTokenExpireCriticalSpan = time.Second * 1              // token about to expire critical time span

var wechatAPITokenConfigCache *gcache.CacheTable
var wechatAPITokenCache *gcache.CacheTable
var wechatThirdPlatformConfigCache *gcache.CacheTable
var wechatThirdPlatformCryptorCache *gcache.CacheTable
var wechatThirdPlatformTokenCache *gcache.CacheTable
var wechatThirdPlatformAuthorizerTokenCache *gcache.CacheTable
var wechatCorpTokenConfigCache *gcache.CacheTable
var wechatCorpTokenCache *gcache.CacheTable

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
    log.Trace("Query %s:(%s) %s", name, key, resultMap)

    config := builder(resultMap[0])
    log.Info("Load %s Cache:(%s) %s", name, key, Json(config))
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
    builder func(resultItem map[string]string) interface{},
    requestor func(key interface{}) (map[string]string, error),
    completeParamBuilder func(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{},
    key interface{},
    args ...interface{}) (*gcache.CacheItem, error) {

    resultMap, err := db.Sql(querySql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        log.Info("Try to request %s:(%s)", name, key)
        count, err := db.Sql(createSql).Params(key).Execute()
        if nil == err && count > 0 {
            tokenItem, err := requestUpdater(name, uncompleteSql, completeSql, lifeSpan,
                builder, requestor, completeParamBuilder, key, args...)
            if nil != err {
                return nil, err
            }
            log.Info("Request %s:(%s) %s", name, key, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem), nil
        }
        log.Warn("Give up request %s:(%s), wait for next cache Query", name, key)
        return nil, &UnexpectedError{Message: "Query " + name + " Later"}
    }
    log.Trace("Query %s:(%s) %s", name, key, resultMap)

    resultItem := resultMap[0]
    updated := resultItem["UPDATED"]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err {
        return nil, err
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        log.Info("Try to request and update %s:(%s)", name, key)
        count, err := db.Sql(updateSql).Params(key).Execute()
        if nil == err && count > 0 {
            tokenItem, err := requestUpdater(name, uncompleteSql, completeSql, lifeSpan,
                builder, requestor, completeParamBuilder, key, args...)
            if nil != err {
                return nil, err
            }
            log.Info("Request %s:(%s) %s", name, key, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem), nil
        }
        log.Warn("Give up request and update %s:(%s), use Query result Temporarily", name, key)
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := Condition(isExpired, lifeSpanTemp, lifeSpan).(time.Duration)
    tokenItem := builder(resultItem)
    log.Info("Load %s Cache:(%s) %s, cache %3.1f min", name, key, Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(key, ls, tokenItem), nil
}

func requestUpdater(
    name string,
    uncompleteSql string,
    completeSql string,
    lifeSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    requestor func(key interface{}) (map[string]string, error),
    completeParamBuilder func(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{},
    key interface{},
    args ...interface{}) (interface{}, error) {

    resultItem, err := requestor(key)
    if nil != err {
        db.Sql(uncompleteSql).Params(key).Execute()
        return nil, err
    }
    // // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    // expireTimeInc := expiresIn - int(lifeSpan.Seconds()*1.1)
    // count, err := db.Sql(completeSql).Params(token, expireTimeInc, key).Execute()
    count, err := db.Sql(completeSql).Params(completeParamBuilder(resultItem, lifeSpan, key)...).Execute()
    if nil != err {
        return nil, err
    }
    if count < 1 {
        return nil, &UnexpectedError{Message: "Record new " + name + " Failed"}
    }

    return builder(resultItem), nil
}

// Wechat access_token cache loader

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

// Wechat third-platform component_access_token cache loader

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

    resultMap, err := db.Sql(queryWechatThirdPlatformAuthorizerTokenSQL).
        Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
        "Unauthorized authorizer app_id"}).(error) // requires that the token already exists
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
        count, err := db.Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
            Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
        if nil == err && count > 0 {
            resultItem, err := wechatThirdPlatformRefreshAuthRequestor(
                tokenKey.CodeName, tokenKey.AuthorizerAppId, authorizerRefreshToken)
            if nil != err {
                db.Sql(uncompleteWechatThirdPlatformAuthorizerTokenSQL).
                    Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
                return nil, err
            }
            count, err := db.Sql(completeWechatThirdPlatformAuthorizerTokenSQL).
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
    count, err := db.Sql(createWechatThirdPlatformAuthorizerTokenUpdating).
        Params(codeName, authorizerAppId).Execute()
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        count, err = db.Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
            Params(codeName, authorizerAppId).Execute()
        for nil != err || count < 1 { // 尝试更新记录失败, 则休眠(1 sec)后再次尝试更新记录
            time.Sleep(1 * time.Second)
            count, err = db.Sql(updateWechatThirdPlatformAuthorizerTokenUpdating).
                Params(codeName, authorizerAppId).Execute()
        }
    }

    // 锁定成功, 开始更新
    resultItem, err := wechatThirdPlatformQueryAuthRequestor(codeName, authorizationCode)
    if nil != err {
        db.Sql(uncompleteWechatThirdPlatformAuthorizerTokenSQL).
            Params(codeName, authorizerAppId).Execute()
        log.Warn("Request WechatThirdPlatformAuthorizerToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, err.Error())
    }
    count, err = db.Sql(completeWechatThirdPlatformAuthorizerTokenSQL).
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

// Wechat Corp access_token cache loader

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
    resultMap, err := db.Sql(queryWechatCorpTokenSQL).Params(codeName).Query()
    if nil != err || 1 != len(resultMap) {
        log.Info("Try to request WechatCorpToken:(%s)", codeName)

        resultItem, err := wechatCorpTokenRequestor(codeName)
        if nil != err {
            log.Warn("Request WechatCorpToken Failed:(%s) %s", codeName, err.Error())
            return nil, err
        }

        count, err := db.Sql(createWechatCorpTokenSQL).Params(
            wechatCorpTokenSQLParamBuilder(resultItem, codeName)...).Execute()
        if nil != err || count < 1 { // 插入记录失败, 则查询记录并缓存
            return queryNewestWechatCorpToken(codeName)
        }

        // 创建成功, 缓存最大缓存时长
        tokenItem := wechatCorpTokenBuilder(resultItem)
        log.Info("Request WechatCorpToken:(%s) %s", codeName, Json(tokenItem))
        return gcache.NewCacheItem(codeName, wechatCorpTokenMaxLifeSpan, tokenItem), nil
    }
    log.Trace("Query WechatCorpToken:(%s) %s", codeName, resultMap)

    resultItem := resultMap[0]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
    if nil != err {
        return nil, err
    }
    effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
    // 即将过期 -> 触发更新
    if effectiveSpan <= wechatCorpTokenExpireCriticalSpan {
        log.Info("Try to request and update WechatCorpToken:(%s)", codeName)

        // 休眠后再请求最新的access_token
        time.Sleep(wechatCorpTokenExpireCriticalSpan)
        resultItem, err := wechatCorpTokenRequestor(codeName)
        if nil != err {
            log.Warn("Request WechatCorpToken Failed:(%s) %s", codeName, err.Error())
            return nil, err
        }

        count, err := db.Sql(updateWechatCorpTokenSQL).Params(
            wechatCorpTokenSQLParamBuilder(resultItem, codeName)...).Execute()
        if nil != err || count < 1 { // 更新记录失败, 则查询记录并缓存
            return queryNewestWechatCorpToken(codeName)
        }

        // 更新成功, 缓存最大缓存时长
        tokenItem := wechatCorpTokenBuilder(resultItem)
        log.Info("Request and Update WechatCorpToken:(%s) %s", codeName, Json(tokenItem))
        return gcache.NewCacheItem(codeName, wechatCorpTokenMaxLifeSpan, tokenItem), nil
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长的一半, 即加快缓存更新频率
    ls := Condition(effectiveSpan > wechatCorpTokenMaxLifeSpan,
        wechatCorpTokenMaxLifeSpan, effectiveSpan/2).(time.Duration)
    tokenItem := wechatCorpTokenBuilder(resultItem)
    log.Info("Load WechatCorpToken Cache:(%s) %s, cache %3.1f min", codeName, Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(codeName, ls, tokenItem), nil
}

func queryNewestWechatCorpToken(codeName interface{}) (*gcache.CacheItem, error) {
    resultMap, err := db.Sql(queryWechatCorpTokenSQL).Params(codeName).Query()
    if nil != err {
        return nil, err
    }

    resultItem := resultMap[0]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
    if nil != err {
        return nil, err
    }
    effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
    if effectiveSpan <= wechatCorpTokenExpireCriticalSpan {
        return nil, &UnexpectedError{Message: "Record WechatCorpToken Failed"}
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长的一半, 即加快缓存更新频率
    ls := Condition(effectiveSpan > wechatCorpTokenMaxLifeSpan,
        wechatCorpTokenMaxLifeSpan, effectiveSpan/2).(time.Duration)
    tokenItem := wechatCorpTokenBuilder(resultItem)
    log.Info("Load WechatCorpToken Cache:(%s) %s, cache %3.1f min", codeName, Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(codeName, ls, tokenItem), nil
}
