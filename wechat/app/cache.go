package app

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "github.com/CharLemAznable/varys/wechat/jsapi"
    "github.com/kataras/golog"
    "time"
)

var configCache *gokits.CacheTable
var tokenCache *gokits.CacheTable

func cacheInitialize() {
    configCache = gokits.CacheExpireAfterWrite("wechat.app.config")
    configCache.SetDataLoader(configLoader)
    tokenCache = gokits.CacheExpireAfterWrite("wechat.app.token")
    tokenCache.SetDataLoader(tokenLoader)
}

type WechatAppConfig struct {
    AppId     string
    AppSecret string
}

func configLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return ConfigLoader(
        "Wechat App",
        &WechatAppConfig{},
        queryConfigSQL,
        configLifeSpan,
        codeName, args...)
}

type WechatAppToken struct {
    AppId       string `json:"appId"`
    AccessToken string `json:"token"`
    JsapiTicket string `json:"ticket"`
}

type Response struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func tokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatAppConfig)

    result, err := gokits.NewHttpReq(tokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request Wechat App Token Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.AccessToken {
        return nil, errors.New("Request Wechat App Token Failed: " + result)
    }

    jsapiTicket := jsapi.TicketRequestor(codeName.(string), response.AccessToken)

    return map[string]string{
        "AppId":       config.AppId,
        "AccessToken": response.AccessToken,
        "JsapiTicket": jsapiTicket,
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

func tokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return TokenLoader(
        "Wechat App",
        &QueryWechatAppToken{},
        queryTokenSQL,
        func(queryDest UpdatedRecord) interface{} {
            query := queryDest.(*QueryWechatAppToken)
            return &WechatAppToken{
                AppId:       query.AppId,
                AccessToken: query.AccessToken,
                JsapiTicket: query.JsapiTicket,
            }
        },
        createTokenSQL,
        updateTokenSQL,
        tokenRequestor,
        uncompleteTokenSQL,
        completeTokenSQL,
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
        tokenLifeSpan,
        tokenTempLifeSpan,
        codeName, args...)
}
