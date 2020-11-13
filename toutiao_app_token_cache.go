package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
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
        &ToutiaoAppConfig{},
        queryToutiaoAppConfigSQL,
        toutiaoAppConfigLifeSpan,
        codeName, args...)
}

type ToutiaoAppToken struct {
    AppId       string `json:"appId"`
    AccessToken string `json:"token"`
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
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request ToutiaoAppToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(ToutiaoAppTokenResponse)).(*ToutiaoAppTokenResponse)
    if nil == response || 0 == len(response.AccessToken) {
        return nil, errors.New("Request ToutiaoAppToken Failed: " + result)
    }
    return map[string]string{
        "AppId":       config.AppId,
        "AccessToken": response.AccessToken,
        "ExpiresIn":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

func toutiaoAppTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoader(
        "ToutiaoAppToken",
        queryToutiaoAppTokenSQL,
        func(query map[string]string) interface{} {
            return &ToutiaoAppToken{
                AppId:       query["AppId"],
                AccessToken: query["AccessToken"],
            }
        },
        createToutiaoAppTokenSQL,
        updateToutiaoAppTokenSQL,
        toutiaoAppTokenRequestor,
        uncompleteToutiaoAppTokenSQL,
        completeToutiaoAppTokenSQL,
        func(response map[string]string, lifeSpan time.Duration) map[string]interface{} {
            expiresIn, _ := gokits.IntFromStr(response["ExpiresIn"])
            return map[string]interface{}{
                "AccessToken": response["AccessToken"],
                // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
                "ExpiresIn": expiresIn - int(lifeSpan.Seconds()*1.1),
            }
        },
        func(response map[string]string) interface{} {
            return &ToutiaoAppToken{
                AppId:       response["AppId"],
                AccessToken: response["AccessToken"],
            }
        },
        toutiaoAppTokenLifeSpan,
        toutiaoAppTokenTempLifeSpan,
        codeName, args...)
}
