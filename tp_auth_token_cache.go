package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var wechatTpAuthTokenCache *gokits.CacheTable

func wechatTpAuthTokenInitialize() {
    wechatTpAuthTokenCache = gokits.CacheExpireAfterWrite("WechatTpAuthTokenCache")
    wechatTpAuthTokenCache.SetDataLoader(wechatTpAuthTokenLoader)
}

type WechatTpAuthKey struct {
    CodeName        string
    AuthorizerAppId string
}

type WechatTpAuthToken struct {
    AppId                 string
    AuthorizerAppId       string
    AuthorizerAccessToken string
}

func wechatTpAuthTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatTpAuthToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AuthorizerAppId = resultItem["AUTHORIZER_APPID"]
    tokenItem.AuthorizerAccessToken = resultItem["AUTHORIZER_ACCESS_TOKEN"]
    return tokenItem
}

func wechatTpAuthTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, codeName interface{}, authorizerAppId interface{}) []interface{} {
    expiresIn, _ := gokits.IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{
        resultItem["AUTHORIZER_ACCESS_TOKEN"],
        resultItem["AUTHORIZER_REFRESH_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), codeName, authorizerAppId}
}

type WechatTpRefreshAuthResponse struct {
    AuthorizerAccessToken  string `json:"authorizer_access_token"`
    ExpiresIn              int    `json:"expires_in"`
    AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

func wechatTpRefreshAuthRequestor(codeName, authorizerAppId, authorizerRefreshToken string) (map[string]string, error) {
    cache, err := wechatTpTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatTpToken)

    result, err := gokits.NewHttpReq(wechatTpRefreshAuthURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":          tokenItem.AppId,
            "authorizer_appid":         authorizerAppId,
            "authorizer_refresh_token": authorizerRefreshToken})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Refresh WechatTpAuthToken Response:(%s, %s) %s", codeName, authorizerAppId, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpRefreshAuthResponse)).(*WechatTpRefreshAuthResponse)
    if nil == response || 0 == len(response.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message: "Refresh WechatTpAuthToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         authorizerAppId,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizerRefreshToken,
        "EXPIRES_IN":               gokits.StrFromInt(response.ExpiresIn)}, nil
}

func wechatTpAuthTokenLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    tokenKey, ok := key.(WechatTpAuthKey)
    if !ok {
        return nil, &UnexpectedError{Message: "WechatTpAuthKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatTpAuthTokenSQL).
        Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Unauthorized authorizer: " + gokits.Json(key)}).(error) // requires that the token already exists
    }
    gokits.LOG.Trace("Query WechatTpAuthToken:(%s) %s", gokits.Json(key), resultMap)

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
        gokits.LOG.Info("Try to request and update WechatTpAuthToken:(%s)", gokits.Json(key))
        count, err := db.New().Sql(updateWechatTpAuthTokenSQL).
            Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
        if nil == err && count > 0 {
            resultItem, err := wechatTpRefreshAuthRequestor(
                tokenKey.CodeName, tokenKey.AuthorizerAppId, authorizerRefreshToken)
            if nil != err {
                _, _ = db.New().Sql(uncompleteWechatTpAuthTokenSQL).
                    Params(tokenKey.CodeName, tokenKey.AuthorizerAppId).Execute()
                return nil, err
            }
            count, err := db.New().Sql(completeWechatTpAuthTokenSQL).
                Params(wechatTpAuthTokenCompleteParamBuilder(
                    resultItem, wechatTpAuthTokenLifeSpan,
                    tokenKey.CodeName, tokenKey.AuthorizerAppId)...).Execute()
            if nil != err || count < 1 {
                return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Replace WechatTpAuthToken Failed"}).(error)
            }

            tokenItem := wechatTpAuthTokenBuilder(resultItem)
            gokits.LOG.Info("Request WechatTpAuthToken:(%s) %s", gokits.Json(key), gokits.Json(tokenItem))
            return gokits.NewCacheItem(key, wechatTpAuthTokenLifeSpan, tokenItem), nil
        }
        _ = gokits.LOG.Warn("Give up request and update WechatTpAuthToken:(%s), use Query result Temporarily", gokits.Json(key))
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := gokits.Condition(isExpired, wechatTpAuthTokenTempLifeSpan,
        wechatTpAuthTokenLifeSpan).(time.Duration)
    tokenItem := wechatTpAuthTokenBuilder(resultItem)
    gokits.LOG.Info("Load WechatTpAuthToken Cache:(%s) %s, cache %3.1f min", gokits.Json(key), gokits.Json(tokenItem), ls.Minutes())
    return gokits.NewCacheItem(key, ls, tokenItem), nil
}
