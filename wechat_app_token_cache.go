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
    JsapiTicket string `json:"ticket"`
}

type WechatAppTokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

type WechatAppTicketResponse struct {
    Errcode   int    `json:"errcode"`
    Errmsg    string `json:"errmsg"`
    Ticket    string `json:"ticket"`
    ExpiresIn int    `json:"expires_in"`
}

func wechatAppTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatAppConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatAppConfig)

    tokenResult, err := gokits.NewHttpReq(wechatAppTokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request WechatAppToken Response:(%s) %s", codeName, tokenResult)
    if nil != err {
        return nil, err
    }

    tokenResponse := gokits.UnJson(tokenResult, new(WechatAppTokenResponse)).(*WechatAppTokenResponse)
    if nil == tokenResponse || "" == tokenResponse.AccessToken {
        return nil, errors.New("Request WechatAppToken Failed: " + tokenResult)
    }

    // request ticket maybe failed, maybe wechat mini app
    ticketResult, err := gokits.NewHttpReq(wechatAppTicketURL).Params(
        "type", "jsapi", "access_token", tokenResponse.AccessToken).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request WechatAppTicket Response:(%s) %s", codeName, ticketResult)
    if nil != err {
        golog.Warnf("Request WechatAppTicket Error: %s", err.Error())
    }

    ticketResponse := new(WechatAppTicketResponse)
    gokits.UnJson(ticketResult, ticketResponse)
    if "" == ticketResponse.Ticket {
        golog.Warnf("Request WechatAppTicket Error: %d - %s", ticketResponse.Errcode, ticketResponse.Errmsg)
        ticketResponse.ExpiresIn = tokenResponse.ExpiresIn
    }

    expiresIn := gokits.Condition(tokenResponse.ExpiresIn < ticketResponse.ExpiresIn,
        tokenResponse.ExpiresIn, ticketResponse.ExpiresIn).(int)
    return map[string]string{
        "AppId":       config.AppId,
        "AccessToken": tokenResponse.AccessToken,
        "JsapiTicket": ticketResponse.Ticket,
        "ExpiresIn":   gokits.StrFromInt(expiresIn)}, nil
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
                JsapiTicket: query.JsapiTicket,
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
                "JsapiTicket": response["JsapiTicket"],
                // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
                "ExpiresIn": expiresIn - int(lifeSpan.Seconds()*1.1),
            }
        },
        func(response map[string]string) interface{} {
            return &WechatAppToken{
                AppId:       response["AppId"],
                AccessToken: response["AccessToken"],
                JsapiTicket: response["JsapiTicket"],
            }
        },
        wechatAppTokenLifeSpan,
        wechatAppTokenTempLifeSpan,
        codeName, args...)
}
