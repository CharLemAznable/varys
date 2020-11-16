package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "time"
)

var wechatAppConfigCache *gokits.CacheTable
var wechatAppTokenCache *gokits.CacheTable

func wechatAppTokenInitialize() {
    wechatAppConfigCache = gokits.CacheExpireAfterWrite("WechatAppConfig")
    wechatAppConfigCache.SetDataLoader(wechatAppConfigLoader)
    wechatAppTokenCache = gokits.CacheExpireAfterWrite("WechatAppToken")
    wechatAppTokenCache.SetDataLoader(wechatAppTokenLoader)
}

type WechatAppConfig struct {
    AppId     string
    AppSecret string
}

func wechatAppConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return configLoader(
        "WechatAppConfig",
        &WechatAppConfig{},
        queryWechatAppConfigSQL,
        wechatAppConfigLifeSpan,
        codeName, args...)
}

type WechatAppToken struct {
    AppId       string `json:"appId"`
    AccessToken string `json:"token"`
}

type WechatAppTokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatAppTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatAppConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatAppConfig)

    result, err := gokits.NewHttpReq(wechatAppTokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request WechatAppToken Response:(%s) %+v", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatAppTokenResponse)).(*WechatAppTokenResponse)
    if nil == response || 0 == len(response.AccessToken) {
        return nil, errors.New("Request WechatAppToken Failed: " + result)
    }
    return map[string]string{
        "AppId":       config.AppId,
        "AccessToken": response.AccessToken,
        "ExpiresIn":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

type QueryWechatAppToken struct {
    WechatAppToken
    Updated    string
    ExpireTime int64
}

func (q *QueryWechatAppToken) GetUpdated() string {
    return q.Updated
}

func (q *QueryWechatAppToken) GetExpireTime() int64 {
    return q.ExpireTime
}

func wechatAppTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoader(
        "WechatAppToken",
        &QueryWechatAppToken{},
        queryWechatAppTokenSQL,
        func(queryDest UpdatedRecord) interface{} {
            query := queryDest.(*QueryWechatAppToken)
            return &WechatAppToken{
                AppId:       query.AppId,
                AccessToken: query.AccessToken,
            }
        },
        createWechatAppTokenSQL,
        updateWechatAppTokenSQL,
        wechatAppTokenRequestor,
        uncompleteWechatAppTokenSQL,
        completeWechatAppTokenSQL,
        func(response map[string]string, lifeSpan time.Duration) map[string]interface{} {
            expiresIn, _ := gokits.IntFromStr(response["ExpiresIn"])
            return map[string]interface{}{
                "AccessToken": response["AccessToken"],
                // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
                "ExpiresIn": expiresIn - int(lifeSpan.Seconds()*1.1),
            }
        },
        func(response map[string]string) interface{} {
            return &WechatAppToken{
                AppId:       response["AppId"],
                AccessToken: response["AccessToken"],
            }
        },
        wechatAppTokenLifeSpan,
        wechatAppTokenTempLifeSpan,
        codeName, args...)
}
