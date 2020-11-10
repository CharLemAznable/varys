package main

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatTpConfigCache *gokits.CacheTable
var wechatTpCryptorCache *gokits.CacheTable
var wechatTpTokenCache *gokits.CacheTable

func wechatTpTokenInitialize() {
    wechatTpConfigCache = gokits.CacheExpireAfterWrite("WechatTpConfig")
    wechatTpConfigCache.SetDataLoader(wechatTpConfigLoader)
    wechatTpCryptorCache = gokits.CacheExpireAfterWrite("WechatTpCryptor")
    wechatTpCryptorCache.SetDataLoader(wechatTpCryptorLoader)
    wechatTpTokenCache = gokits.CacheExpireAfterWrite("WechatTpToken")
    wechatTpTokenCache.SetDataLoader(wechatTpTokenLoader)
}

type WechatTpConfig struct {
    AppId       string
    AppSecret   string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatTpConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return configLoader(
        "WechatTpConfig",
        queryWechatTpConfigSQL,
        wechatTpConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatTpConfig)
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

func wechatTpCryptorLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    cache, err := wechatTpConfigCache.Value(codeName)
    if nil != err {
        return nil, &UnexpectedError{Message: "Require WechatTpConfig with key: " + codeName.(string)} // require config
    }
    config := cache.Data().(*WechatTpConfig)
    gokits.LOG.Trace("Query WechatTpConfig Cache:(%s) %s", codeName, gokits.Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    gokits.LOG.Info("Load WechatTpCryptor Cache:(%s) %s", codeName, cryptor)
    return gokits.NewCacheItem(codeName, wechatTpCryptorLifeSpan, cryptor), nil
}

type WechatTpToken struct {
    AppId       string
    AccessToken string
}

func wechatTpTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatTpToken)
    tokenItem.AppId = resultItem["APP_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
}

type WechatTpTokenResponse struct {
    ComponentAccessToken string `json:"component_access_token"`
    ExpiresIn            int    `json:"expires_in"`
}

func wechatTpTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatTpConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatTpConfig)

    ticket, err := queryWechatTpTicket(codeName.(string))
    if nil != err {
        return nil, err
    }

    result, err := gokits.NewHttpReq(wechatTpTokenURL).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatTpToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpTokenResponse)).(*WechatTpTokenResponse)
    if nil == response || 0 == len(response.ComponentAccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatTpToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":       config.AppId,
        "ACCESS_TOKEN": response.ComponentAccessToken,
        "EXPIRES_IN":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

func wechatTpTokenCompleteParamBuilder(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{} {
    expiresIn, _ := gokits.IntFromStr(resultItem["EXPIRES_IN"])
    return []interface{}{resultItem["ACCESS_TOKEN"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        expiresIn - int(lifeSpan.Seconds()*1.1), key}
}

// 获取第三方平台component_access_token
func wechatTpTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoader(
        "WechatTpToken",
        queryWechatTpTokenSQL,
        createWechatTpTokenSQL,
        updateWechatTpTokenSQL,
        uncompleteWechatTpTokenSQL,
        completeWechatTpTokenSQL,
        wechatTpTokenLifeSpan,
        wechatTpTokenTempLifeSpan,
        wechatTpTokenBuilder,
        wechatTpTokenRequestor,
        wechatTpTokenCompleteParamBuilder,
        codeName, args...)
}
