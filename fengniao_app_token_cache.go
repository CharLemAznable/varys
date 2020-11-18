package main

import (
    "crypto/md5"
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "math/rand"
    "net/url"
    "time"
)

var fengniaoAppConfigCache *gokits.CacheTable
var fengniaoAppTokenCache *gokits.CacheTable

func fengniaoAppTokenInitialize() {
    rand.Seed(time.Now().UnixNano())

    fengniaoAppConfigCache = gokits.CacheExpireAfterWrite("FengniaoAppConfig")
    fengniaoAppConfigCache.SetDataLoader(fengniaoAppConfigLoader)
    fengniaoAppTokenCache = gokits.CacheExpireAfterWrite("FengniaoAppToken")
    fengniaoAppTokenCache.SetDataLoader(fengniaoAppTokenLoader)
}

func newSalt() int {
    return (rand.Intn(9)+1)*1000 + rand.Intn(1000) // [1,9]*1000+[0,999]
}

type FengniaoAppConfig struct {
    AppId     string
    SecretKey string
}

func fengniaoAppConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return configLoader(
        "FengniaoAppConfig",
        &FengniaoAppConfig{},
        queryFengniaoAppConfigSQL,
        fengniaoAppConfigLifeSpan,
        codeName, args...)
}

type FengniaoAppToken struct {
    AppId       string `json:"appId"`
    AccessToken string `json:"token"`
}

type FengniaoAppTokenResponse struct {
    Code string               `json:"code"`
    Msg  int                  `json:"msg"`
    Data FengniaoAppTokenData `json:"data"`
}

type FengniaoAppTokenData struct {
    AccessToken string `json:"access_token"`
    ExpireTime  int    `json:"expire_time"`
}

func fengniaoAppTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := fengniaoAppConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*FengniaoAppConfig)

    appId := config.AppId
    salt := gokits.StrFromInt(newSalt())
    plainText := "app_id=" + appId + "&salt=" + salt + "&secret_key=" + config.SecretKey
    signature := fmt.Sprintf("%x", md5.Sum([]byte(url.QueryEscape(plainText))))

    result, err := gokits.NewHttpReq(fengniaoAppTokenURL).Params(
        "app_id", appId, "salt", salt, "signature", signature).
        Prop("Content-Type", "application/json").Get()
    golog.Debugf("Request FengniaoAppToken Response:(%s) %+v", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(FengniaoAppTokenResponse)).(*FengniaoAppTokenResponse)
    if nil == response || "" == response.Data.AccessToken {
        return nil, errors.New("Request FengniaoAppToken Failed: " + result)
    }
    return map[string]string{
        "AppId":       config.AppId,
        "AccessToken": response.Data.AccessToken,
        "ExpireTime":  gokits.StrFromInt(response.Data.ExpireTime)}, nil
}

type QueryFengniaoAppToken struct {
    FengniaoAppToken
    Updated    string
    ExpireTime int64
}

func (q *QueryFengniaoAppToken) GetUpdated() string {
    return q.Updated
}

func (q *QueryFengniaoAppToken) GetExpireTime() int64 {
    return q.ExpireTime
}

func fengniaoAppTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoader(
        "FengniaoAppToken",
        &QueryFengniaoAppToken{},
        queryFengniaoAppTokenSQL,
        func(queryDest UpdatedRecord) interface{} {
            query := queryDest.(*QueryFengniaoAppToken)
            return &FengniaoAppToken{
                AppId:       query.AppId,
                AccessToken: query.AccessToken,
            }
        },
        createFengniaoAppTokenSQL,
        updateFengniaoAppTokenSQL,
        fengniaoAppTokenRequestor,
        uncompleteFengniaoAppTokenSQL,
        completeFengniaoAppTokenSQL,
        func(response map[string]string, lifeSpan time.Duration) map[string]interface{} {
            return map[string]interface{}{
                "AccessToken": response["AccessToken"],
                // 过期时间设置为12小时: 12(h)*60(m)*60(s)
                "ExpireTime": time.Now().Unix() + 12*60*60, // in second
            }
        },
        func(response map[string]string) interface{} {
            return &FengniaoAppToken{
                AppId:       response["AppId"],
                AccessToken: response["AccessToken"],
            }
        },
        fengniaoAppTokenLifeSpan,
        fengniaoAppTokenTempLifeSpan,
        codeName, args...)
}
