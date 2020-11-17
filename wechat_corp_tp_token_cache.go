package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "github.com/kataras/golog"
    "time"
)

var wechatCorpTpConfigCache *gokits.CacheTable
var wechatCorpTpCryptorCache *gokits.CacheTable
var wechatCorpTpTokenCache *gokits.CacheTable

func wechatCorpTpTokenInitialize() {
    wechatCorpTpConfigCache = gokits.CacheExpireAfterWrite("WechatCorpTpConfigCache")
    wechatCorpTpConfigCache.SetDataLoader(wechatCorpTpConfigLoader)
    wechatCorpTpCryptorCache = gokits.CacheExpireAfterWrite("WechatCorpTpCryptorCache")
    wechatCorpTpCryptorCache.SetDataLoader(wechatCorpTpCryptorLoader)
    wechatCorpTpTokenCache = gokits.CacheExpireAfterWrite("WechatCorpTpTokenCache")
    wechatCorpTpTokenCache.SetDataLoader(wechatCorpTpTokenLoader)
}

type WechatCorpTpConfig struct {
    SuiteId     string
    SuiteSecret string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatCorpTpConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return configLoader(
        "WechatCorpTpConfig",
        &WechatCorpTpConfig{},
        queryWechatCorpTpConfigSQL,
        wechatCorpTpConfigLifeSpan,
        codeName, args...)
}

func wechatCorpTpCryptorLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    cache, err := wechatCorpTpConfigCache.Value(codeName)
    if nil != err {
        return nil, errors.New("Require WechatCorpTpConfig with key: " + codeName.(string)) // require config
    }
    config := cache.Data().(*WechatCorpTpConfig)
    golog.Debugf("Query WechatCorpTpConfig Cache:(%s) %+v", codeName, config)

    cryptor, err := wechataes.NewWechatCryptor(config.SuiteId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    golog.Infof("Load WechatCorpTpCryptor Cache:(%s) %+v", codeName, cryptor)
    return gokits.NewCacheItem(codeName, wechatCorpTpCryptorLifeSpan, cryptor), nil
}

type WechatCorpTpToken struct {
    SuiteId     string
    AccessToken string
}

type WechatCorpTpTokenResponse struct {
    Errcode          int    `json:"errcode"`
    Errmsg           string `json:"errmsg"`
    SuiteAccessToken string `json:"suite_access_token"`
    ExpiresIn        int    `json:"expires_in"`
}

func wechatCorpTpTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatCorpTpConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatCorpTpConfig)

    var ticket string
    err = db.NamedGet(&ticket, queryWechatCorpTpTicketSQL,
        map[string]interface{}{"CodeName": codeName})
    if nil != err {
        return nil, err
    }

    result, err := gokits.NewHttpReq(wechatCorpTpTokenURL).
        RequestBody(gokits.Json(map[string]string{
            "suite_id":     config.SuiteId,
            "suite_secret": config.SuiteSecret,
            "suite_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request WechatCorpTpToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpTpTokenResponse)).(*WechatCorpTpTokenResponse)
    if nil == response || 0 == len(response.SuiteAccessToken) {
        return nil, errors.New("Request WechatCorpTpToken Failed: " + result)
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SuiteId":     config.SuiteId,
        "AccessToken": response.SuiteAccessToken,
        "ExpireTime":  gokits.StrFromInt64(expireTime)}, nil
}

type QueryWechatCorpTpToken struct {
    WechatCorpTpToken
    ExpireTime int64
}

func (q *QueryWechatCorpTpToken) GetExpireTime() int64 {
    return q.ExpireTime
}

func wechatCorpTpTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoaderStrict(
        "WechatCorpTpToken",
        &QueryWechatCorpTpToken{},
        queryWechatCorpTpTokenSQL,
        func(queryDest ExpireTimeRecord) interface{} {
            query := queryDest.(*QueryWechatCorpTpToken)
            return &WechatCorpTpToken{
                SuiteId:     query.SuiteId,
                AccessToken: query.AccessToken,
            }
        },
        wechatCorpTpTokenRequestor,
        createWechatCorpTpTokenSQL,
        updateWechatCorpTpTokenSQL,
        func(response map[string]string) map[string]interface{} {
            expireTime, _ := gokits.Int64FromStr(response["ExpireTime"])
            return map[string]interface{}{
                "AccessToken": response["AccessToken"],
                "ExpireTime":  expireTime}
        },
        wechatCorpTpTokenExpireCriticalSpan,
        func(response map[string]string) interface{} {
            return &WechatCorpTpToken{
                SuiteId:     response["SuiteId"],
                AccessToken: response["AccessToken"],
            }
        },
        wechatCorpTpTokenMaxLifeSpan,
        codeName, args...)
}
