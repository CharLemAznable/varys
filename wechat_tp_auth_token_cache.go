package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
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
    AppId                 string `json:"appId"`
    AuthorizerAppId       string `json:"authorizerAppId"`
    AuthorizerAccessToken string `json:"token"`
    AuthorizerJsapiTicket string `json:"ticket"`
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
    golog.Debugf("Refresh WechatTpAuthToken Response:(%s, %s) %s", codeName, authorizerAppId, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpRefreshAuthResponse)).(*WechatTpRefreshAuthResponse)
    if nil == response || "" == response.AuthorizerAccessToken {
        return nil, errors.New("Refresh WechatTpAuthToken Failed: " + result)
    }

    jsapiTicket := wechatJsapiTicketRequestor(codeName, response.AuthorizerAccessToken)

    return map[string]string{
        "AppId":                  tokenItem.AppId,
        "AuthorizerAppId":        authorizerAppId,
        "AuthorizerAccessToken":  response.AuthorizerAccessToken,
        "AuthorizerRefreshToken": response.AuthorizerRefreshToken,
        "AuthorizerJsapiTicket":  jsapiTicket,
        "ExpiresIn":              gokits.StrFromInt(response.ExpiresIn)}, nil
}

type QueryWechatTpAuthToken struct {
    WechatTpAuthToken
    AuthorizerRefreshToken string
    Updated                string
    ExpireTime             int64
}

func wechatTpAuthTokenLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    tokenKey, ok := key.(WechatTpAuthKey)
    if !ok {
        return nil, errors.New("WechatTpAuthKey type error") // key type error
    }

    query := &QueryWechatTpAuthToken{}
    err := db.NamedGet(query, queryWechatTpAuthTokenSQL,
        map[string]interface{}{"CodeName": tokenKey.CodeName,
            "AuthorizerAppId": tokenKey.AuthorizerAppId})
    if nil != err {
        return nil, err // requires that the token already exists
    }
    golog.Debugf("Query WechatTpAuthToken:(%+v) %+v", key, query)

    authorizerRefreshToken := query.AuthorizerRefreshToken // requires that the refresh token exists
    if "" == authorizerRefreshToken {
        return nil, errors.New("AuthorizerRefreshToken is Empty")
    }
    isExpired := time.Now().Unix() > query.ExpireTime
    isUpdated := "1" == query.Updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        golog.Debugf("Try to request and update WechatTpAuthToken:(%+v)", key)
        count, err := db.NamedExecX(updateWechatTpAuthTokenSQL,
            map[string]interface{}{"CodeName": tokenKey.CodeName,
                "AuthorizerAppId": tokenKey.AuthorizerAppId})
        if nil == err && count > 0 {
            response, err := wechatTpRefreshAuthRequestor(
                tokenKey.CodeName, tokenKey.AuthorizerAppId, authorizerRefreshToken)
            if nil != err {
                _, _ = db.NamedExec(uncompleteWechatTpAuthTokenSQL,
                    map[string]interface{}{"CodeName": tokenKey.CodeName,
                        "AuthorizerAppId": tokenKey.AuthorizerAppId})
                return nil, err
            }
            completeArg := wechatTpAuthCompleteArg(response, wechatTpAuthTokenLifeSpan)
            completeArg["CodeName"] = tokenKey.CodeName
            _, err = db.NamedExec(completeWechatTpAuthTokenSQL, completeArg)
            if nil != err {
                return nil, err
            }

            token := wechatTpAuthTokenBuilder(response)
            golog.Infof("Request WechatTpAuthToken:(%+v) %+v", key, token)
            return gokits.NewCacheItem(key, wechatTpAuthTokenLifeSpan, token), nil
        }
        golog.Warnf("Give up request and update WechatTpAuthToken:(%+v), use Query result Temporarily", key)
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := gokits.Condition(isExpired,
        wechatTpAuthTokenTempLifeSpan,
        wechatTpAuthTokenLifeSpan).(time.Duration)
    token := &WechatTpAuthToken{AppId: query.AppId,
        AuthorizerAppId:       query.AuthorizerAppId,
        AuthorizerAccessToken: query.AuthorizerAccessToken,
        AuthorizerJsapiTicket: query.AuthorizerJsapiTicket}
    golog.Infof("Load WechatTpAuthToken Cache:(%+v) %+v, cache %3.1f min", key, token, ls.Minutes())
    return gokits.NewCacheItem(key, ls, token), nil
}
