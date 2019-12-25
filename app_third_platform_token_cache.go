package main

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatAppThirdPlatformConfigCache *gokits.CacheTable
var wechatAppThirdPlatformCryptorCache *gokits.CacheTable
var wechatAppThirdPlatformTokenCache *gokits.CacheTable
var wechatAppThirdPlatformAuthorizerTokenCache *gokits.CacheTable

func wechatAppThirdPlatformAuthorizerTokenInitialize() {
    wechatAppThirdPlatformConfigCache = gokits.CacheExpireAfterWrite("wechatAppThirdPlatformConfig")
    wechatAppThirdPlatformConfigCache.SetDataLoader(wechatAppThirdPlatformConfigLoader)
    wechatAppThirdPlatformCryptorCache = gokits.CacheExpireAfterWrite("wechatAppThirdPlatformCryptor")
    wechatAppThirdPlatformCryptorCache.SetDataLoader(wechatAppThirdPlatformCryptorLoader)
    wechatAppThirdPlatformTokenCache = gokits.CacheExpireAfterWrite("wechatAppThirdPlatformToken")
    wechatAppThirdPlatformTokenCache.SetDataLoader(wechatAppThirdPlatformTokenLoader)
    wechatAppThirdPlatformAuthorizerTokenCache = gokits.CacheExpireAfterWrite("wechatAppThirdPlatformAuthorizerToken")
    wechatAppThirdPlatformAuthorizerTokenCache.SetDataLoader(wechatAppThirdPlatformAuthorizerTokenLoader)
}

type WechatAppThirdPlatformConfig struct {
    AppId       string
    AppSecret   string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatAppThirdPlatformConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
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

func wechatAppThirdPlatformCryptorLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    cache, err := wechatAppThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, &UnexpectedError{Message: "Require WechatAppThirdPlatformConfig with key: " + codeName.(string)} // require config
    }
    config := cache.Data().(*WechatAppThirdPlatformConfig)
    gokits.LOG.Trace("Query WechatAppThirdPlatformConfig Cache:(%s) %s", codeName, gokits.Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    gokits.LOG.Info("Load WechatAppThirdPlatformCryptor Cache:(%s) %s", codeName, cryptor)
    return gokits.NewCacheItem(codeName, wechatAppThirdPlatformCryptorLifeSpan, cryptor), nil
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

type WechatAppThirdPlatformTokenResponse struct {
    ComponentAccessToken string `json:"component_access_token"`
    ExpiresIn            int    `json:"expires_in"`
}

func wechatAppThirdPlatformTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatAppThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatAppThirdPlatformConfig)

    ticket, err := queryWechatAppThirdPlatformTicket(codeName.(string))
    if nil != err {
        return nil, err
    }

    result, err := gokits.NewHttpReq(wechatAppThirdPlatformTokenURL).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatAppThirdPlatformToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatAppThirdPlatformTokenResponse)).(*WechatAppThirdPlatformTokenResponse)
    if nil == response || 0 == len(response.ComponentAccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatAppThirdPlatformToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":       config.AppId,
        "ACCESS_TOKEN": response.ComponentAccessToken,
        "EXPIRES_IN":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

func wechatAppThirdPlatformTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := gokits.IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

// 获取第三方平台component_access_token
func wechatAppThirdPlatformTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
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
    expiresIn, _ := gokits.IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{
        resultItem["AUTHORIZER_ACCESS_TOKEN"],
        resultItem["AUTHORIZER_REFRESH_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), codeName, authorizerAppId}
}

type WechatAppThirdPlatformRefreshAuthResponse struct {
    AuthorizerAccessToken  string `json:"authorizer_access_token"`
    ExpiresIn              int    `json:"expires_in"`
    AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

func wechatAppThirdPlatformRefreshAuthRequestor(codeName, authorizerAppId, authorizerRefreshToken string) (map[string]string, error) {
    cache, err := wechatAppThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatAppThirdPlatformToken)

    result, err := gokits.NewHttpReq(wechatAppThirdPlatformRefreshAuthURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":          tokenItem.AppId,
            "authorizer_appid":         authorizerAppId,
            "authorizer_refresh_token": authorizerRefreshToken})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Refresh WechatAppThirdPlatformAuthorizerToken Response:(%s, %s) %s", codeName, authorizerAppId, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatAppThirdPlatformRefreshAuthResponse)).(*WechatAppThirdPlatformRefreshAuthResponse)
    if nil == response || 0 == len(response.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message: "Refresh WechatAppThirdPlatformAuthorizerToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         authorizerAppId,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizerRefreshToken,
        "EXPIRES_IN":               gokits.StrFromInt(response.ExpiresIn)}, nil
}

func wechatAppThirdPlatformAuthorizerTokenLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    tokenKey, ok := key.(WechatAppThirdPlatformAuthorizerKey)
    if !ok {
        return nil, &UnexpectedError{Message: "WechatAppThirdPlatformAuthorizerKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatAppThirdPlatformAuthorizerTokenSQL).
        Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Unauthorized authorizer: " + gokits.Json(key)}).(error) // requires that the token already exists
    }
    gokits.LOG.Trace("Query WechatAppThirdPlatformAuthorizerToken:(%s) %s", gokits.Json(key), resultMap)

    resultItem := resultMap[0]
    authorizerRefreshToken := resultItem["AUTHORIZER_REFRESH_TOKEN"] // requires that the refresh token exists
    updated := resultItem["UPDATED"]
    expireTime, err := gokits.Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err || 0 == len(authorizerRefreshToken) {
        return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Refresh token is Empty"}).(error)
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        gokits.LOG.Info("Try to request and update WechatAppThirdPlatformAuthorizerToken:(%s)", gokits.Json(key))
        count, err := db.New().Sql(updateWechatAppThirdPlatformAuthorizerTokenSQL).
            Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
        if nil == err && count > 0 {
            resultItem, err := wechatAppThirdPlatformRefreshAuthRequestor(
                tokenKey.CodeName, tokenKey.AuthorizerAppId, authorizerRefreshToken)
            if nil != err {
                _, _ = db.New().Sql(uncompleteWechatAppThirdPlatformAuthorizerTokenSQL).
                    Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
                return nil, err
            }
            count, err := db.New().Sql(completeWechatAppThirdPlatformAuthorizerTokenSQL).
                Params(wechatAppThirdPlatformAuthorizerTokenCompleteParamBuilder(
                    resultItem, wechatAppThirdPlatformAuthorizerTokenLifeSpan,
                    tokenKey.CodeName, tokenKey.AuthorizerAppId)...).Execute()
            if nil != err || count < 1 {
                return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Replace WechatAppThirdPlatformAuthorizerToken Failed"}).(error)
            }

            tokenItem := wechatAppThirdPlatformAuthorizerTokenBuilder(resultItem)
            gokits.LOG.Info("Request WechatAppThirdPlatformAuthorizerToken:(%s) %s", gokits.Json(key), gokits.Json(tokenItem))
            return gokits.NewCacheItem(key, wechatAppThirdPlatformAuthorizerTokenLifeSpan, tokenItem), nil
        }
        _ = gokits.LOG.Warn("Give up request and update WechatAppThirdPlatformAuthorizerToken:(%s), use Query result Temporarily", gokits.Json(key))
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := gokits.Condition(isExpired, wechatAppThirdPlatformAuthorizerTokenTempLifeSpan,
        wechatAppThirdPlatformAuthorizerTokenLifeSpan).(time.Duration)
    tokenItem := wechatAppThirdPlatformAuthorizerTokenBuilder(resultItem)
    gokits.LOG.Info("Load WechatAppThirdPlatformAuthorizerToken Cache:(%s) %s, cache %3.1f min", gokits.Json(key), gokits.Json(tokenItem), ls.Minutes())
    return gokits.NewCacheItem(key, ls, tokenItem), nil
}

type WechatAppThirdPlatformQueryAuthResponse struct {
    AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

type AuthorizationInfo struct {
    AuthorizerAppid        string     `json:"authorizer_appid"`
    AuthorizerAccessToken  string     `json:"authorizer_access_token"`
    ExpiresIn              int        `json:"expires_in"`
    AuthorizerRefreshToken string     `json:"authorizer_refresh_token"`
    FuncInfo               []FuncInfo `json:"func_info"`
}

type FuncInfo struct {
    FuncscopeCategory FuncscopeCategory `json:"funcscope_category"`
}

type FuncscopeCategory struct {
    Id int `json:"id"`
}

func wechatAppThirdPlatformQueryAuthRequestor(codeName, authorizationCode interface{}) (map[string]string, error) {
    cache, err := wechatAppThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatAppThirdPlatformToken)

    result, err := gokits.NewHttpReq(wechatAppThirdPlatformQueryAuthURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":    tokenItem.AppId,
            "authorization_code": authorizationCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatAppThirdPlatformAuthorizerToken Response:(%s, ) %s", codeName, authorizationCode, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatAppThirdPlatformQueryAuthResponse)).(*WechatAppThirdPlatformQueryAuthResponse)
    if nil == response || 0 == len(response.AuthorizationInfo.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatAppThirdPlatformAuthorizerToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         response.AuthorizationInfo.AuthorizerAppid,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizationInfo.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizationInfo.AuthorizerRefreshToken,
        "EXPIRES_IN":               gokits.StrFromInt(response.AuthorizationInfo.ExpiresIn)}, nil
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
        _, _ = db.New().Sql(uncompleteWechatAppThirdPlatformAuthorizerTokenSQL).
            Params(codeName, authorizerAppId).Execute()
        _ = gokits.LOG.Warn("Request WechatAppThirdPlatformAuthorizerToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, err.Error())
        return
    }
    count, err = db.New().Sql(completeWechatAppThirdPlatformAuthorizerTokenSQL).
        Params(wechatAppThirdPlatformAuthorizerTokenCompleteParamBuilder(
            resultItem, wechatAppThirdPlatformAuthorizerTokenLifeSpan,
            codeName, authorizerAppId)...).Execute()
    if nil != err || count < 1 {
        _ = gokits.LOG.Warn("Record new WechatAppThirdPlatformAuthorizerToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Replace WechatAppThirdPlatformAuthorizerToken Failed"}).(error).Error())
        return
    }

    tokenItem := wechatAppThirdPlatformAuthorizerTokenBuilder(resultItem)
    gokits.LOG.Info("Request WechatAppThirdPlatformAuthorizerToken:(%s, %s) %s",
        codeName, authorizerAppId, gokits.Json(tokenItem))
    wechatAppThirdPlatformAuthorizerTokenCache.Add(WechatAppThirdPlatformAuthorizerKey{
        CodeName: codeName.(string), AuthorizerAppId: authorizerAppId.(string)},
        wechatAppThirdPlatformAuthorizerTokenLifeSpan, tokenItem)
}
