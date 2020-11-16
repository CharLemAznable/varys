package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "time"
)

var wechatCorpConfigCache *gokits.CacheTable
var wechatCorpTokenCache *gokits.CacheTable

func wechatCorpTokenInitialize() {
    wechatCorpConfigCache = gokits.CacheExpireAfterWrite("WechatCorpConfig")
    wechatCorpConfigCache.SetDataLoader(wechatCorpTokenConfigLoader)
    wechatCorpTokenCache = gokits.CacheExpireAfterWrite("WechatCorpToken")
    wechatCorpTokenCache.SetDataLoader(wechatCorpTokenLoader)
}

type WechatCorpConfig struct {
    CorpId     string
    CorpSecret string
}

func wechatCorpTokenConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return configLoader(
        "WechatCorpConfig",
        &WechatCorpConfig{},
        queryWechatCorpConfigSQL,
        wechatCorpConfigLifeSpan,
        codeName, args...)
}

type WechatCorpToken struct {
    CorpId      string `json:"corpId"`
    AccessToken string `json:"token"`
}

type WechatCorpTokenResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatCorpConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatCorpConfig)

    result, err := gokits.NewHttpReq(wechatCorpTokenURL).Params(
        "corpid", config.CorpId, "corpsecret", config.CorpSecret).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request WechatCorpToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpTokenResponse)).(*WechatCorpTokenResponse)
    if nil == response || 0 != response.Errcode || 0 == len(response.AccessToken) {
        return nil, errors.New("Request Corp access_token Failed: " + result)
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "CorpId":      config.CorpId,
        "AccessToken": response.AccessToken,
        "ExpireTime":  gokits.StrFromInt64(expireTime)}, nil
}

type QueryWechatCorpToken struct {
    WechatCorpToken
    ExpireTime int64
}

func (q *QueryWechatCorpToken) GetExpireTime() int64 {
    return q.ExpireTime
}

func wechatCorpTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoaderStrict(
        "WechatCorpToken",
        &QueryWechatCorpToken{},
        queryWechatCorpTokenSQL,
        func(queryDest ExpireTimeRecord) interface{} {
            query := queryDest.(*QueryWechatCorpToken)
            return &WechatCorpToken{
                CorpId:      query.CorpId,
                AccessToken: query.AccessToken,
            }
        },
        wechatCorpTokenRequestor,
        createWechatCorpTokenSQL,
        updateWechatCorpTokenSQL,
        func(response map[string]string) map[string]interface{} {
            expireTime, _ := gokits.Int64FromStr(response["ExpireTime"])
            return map[string]interface{}{
                "AccessToken": response["AccessToken"],
                "ExpireTime":  expireTime}
        },
        wechatCorpTokenExpireCriticalSpan,
        func(response map[string]string) interface{} {
            return &WechatCorpToken{
                CorpId:      response["CorpId"],
                AccessToken: response["AccessToken"],
            }
        },
        wechatCorpTokenMaxLifeSpan,
        codeName, args...)
}
