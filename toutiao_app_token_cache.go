package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var toutiaoAppConfigCache *gokits.CacheTable
var toutiaoAppTokenCache *gokits.CacheTable

func toutiaoAppTokenInitialize() {
    toutiaoAppConfigCache = gokits.CacheExpireAfterWrite("ToutiaoAppConfig")
    toutiaoAppConfigCache.SetDataLoader(toutiaoAppConfigLoader)
    toutiaoAppTokenCache = gokits.CacheExpireAfterWrite("ToutiaoAppToken")
    toutiaoAppTokenCache.SetDataLoader(toutiaoAppTokenLoader)
}

type ToutiaoAppConfig struct {
    AppId     string
    AppSecret string
}

func toutiaoAppConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return configLoader(
        "ToutiaoAppConfig",
        queryToutiaoAppConfigSQL,
        toutiaoAppConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(ToutiaoAppConfig)
            config.AppId = resultItem["APP_ID"]
            config.AppSecret = resultItem["APP_SECRET"]
            if 0 == len(config.AppId) || 0 == len(config.AppSecret) {
                return nil
            }
            return config
        },
        codeName, args...)
}

type ToutiaoAppToken struct {
    AppId       string
    AccessToken string
}

func toutiaoAppTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(ToutiaoAppToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
}

type ToutiaoAppTokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func toutiaoAppTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := toutiaoAppConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*ToutiaoAppConfig)

    result, err := gokits.NewHttpReq(toutiaoAppTokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    gokits.LOG.Trace("Request ToutiaoAppToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(ToutiaoAppTokenResponse)).(*ToutiaoAppTokenResponse)
    if nil == response || 0 == len(response.AccessToken) {
        return nil, &UnexpectedError{Message: "Request ToutiaoAppToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":       config.AppId,
        "ACCESS_TOKEN": response.AccessToken,
        "EXPIRES_IN":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

func toutiaoAppTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := gokits.IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

func toutiaoAppTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoader(
        "ToutiaoAppToken",
        queryToutiaoAppTokenSQL,
        createToutiaoAppTokenSQL,
        updateToutiaoAppTokenSQL,
        uncompleteToutiaoAppTokenSQL,
        completeToutiaoAppTokenSQL,
        toutiaoAppTokenLifeSpan,
        toutiaoAppTokenTempLifeSpan,
        toutiaoAppTokenBuilder,
        toutiaoAppTokenRequestor,
        toutiaoAppTokenCompleteParamBuilder,
        codeName, args...)
}
