package main

import (
    "github.com/CharLemAznable/gokits"
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
        queryWechatAppConfigSQL,
        wechatAppConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatAppConfig)
            config.AppId = resultItem["APP_ID"]
            config.AppSecret = resultItem["APP_SECRET"]
            if 0 == len(config.AppId) || 0 == len(config.AppSecret) {
                return nil
            }
            return config
        },
        codeName, args...)
}

type WechatAppToken struct {
    AppId       string
    AccessToken string
}

func wechatAppTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatAppToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
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
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    gokits.LOG.Trace("Request WechatAppToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatAppTokenResponse)).(*WechatAppTokenResponse)
    if nil == response || 0 == len(response.AccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatAppToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":       config.AppId,
        "ACCESS_TOKEN": response.AccessToken,
        "EXPIRES_IN":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

func wechatAppTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := gokits.IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

func wechatAppTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoader(
        "WechatAppToken",
        queryWechatAppTokenSQL,
        createWechatAppTokenSQL,
        updateWechatAppTokenSQL,
        uncompleteWechatAppTokenSQL,
        completeWechatAppTokenSQL,
        wechatAppTokenLifeSpan,
        wechatAppTokenTempLifeSpan,
        wechatAppTokenBuilder,
        wechatAppTokenRequestor,
        wechatAppTokenCompleteParamBuilder,
        codeName, args...)
}
