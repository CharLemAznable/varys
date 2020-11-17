package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "github.com/kataras/golog"
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
        &WechatTpConfig{},
        queryWechatTpConfigSQL,
        wechatTpConfigLifeSpan,
        codeName, args...)
}

func wechatTpCryptorLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    cache, err := wechatTpConfigCache.Value(codeName)
    if nil != err {
        return nil, errors.New("Require WechatTpConfig with key: " + codeName.(string)) // require config
    }
    config := cache.Data().(*WechatTpConfig)
    golog.Debugf("Query WechatTpConfig Cache:(%s) %+v", codeName, config)

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    golog.Infof("Load WechatTpCryptor Cache:(%s) %+v", codeName, cryptor)
    return gokits.NewCacheItem(codeName, wechatTpCryptorLifeSpan, cryptor), nil
}

type WechatTpToken struct {
    AppId       string `json:"appId"`
    AccessToken string `json:"token"`
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

    var ticket string
    err = db.NamedGet(&ticket, queryWechatTpTicketSQL,
        map[string]interface{}{"CodeName": codeName})
    if nil != err {
        return nil, err
    }

    result, err := gokits.NewHttpReq(wechatTpTokenURL).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request WechatTpToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpTokenResponse)).(*WechatTpTokenResponse)
    if nil == response || 0 == len(response.ComponentAccessToken) {
        return nil, errors.New("Request WechatTpToken Failed: " + result)
    }
    return map[string]string{
        "AppId":       config.AppId,
        "AccessToken": response.ComponentAccessToken,
        "ExpiresIn":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

type QueryWechatTpToken struct {
    WechatTpToken
    Updated    string
    ExpireTime int64
}

func (q *QueryWechatTpToken) GetUpdated() string {
    return q.Updated
}

func (q *QueryWechatTpToken) GetExpireTime() int64 {
    return q.ExpireTime
}

// 获取第三方平台component_access_token
func wechatTpTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoader(
        "WechatTpToken",
        &QueryWechatTpToken{},
        queryWechatTpTokenSQL,
        func(queryDest UpdatedRecord) interface{} {
            query := queryDest.(*QueryWechatTpToken)
            return &WechatTpToken{
                AppId:       query.AppId,
                AccessToken: query.AccessToken,
            }
        },
        createWechatTpTokenSQL,
        updateWechatTpTokenSQL,
        wechatTpTokenRequestor,
        uncompleteWechatTpTokenSQL,
        completeWechatTpTokenSQL,
        func(response map[string]string, lifeSpan time.Duration) map[string]interface{} {
            expiresIn, _ := gokits.IntFromStr(response["ExpiresIn"])
            return map[string]interface{}{
                "AccessToken": response["AccessToken"],
                // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
                "ExpiresIn": expiresIn - int(lifeSpan.Seconds()*1.1),
            }
        },
        func(response map[string]string) interface{} {
            return &WechatTpToken{
                AppId:       response["AppId"],
                AccessToken: response["AccessToken"],
            }
        },
        wechatTpTokenLifeSpan,
        wechatTpTokenTempLifeSpan,
        codeName, args...)
}
