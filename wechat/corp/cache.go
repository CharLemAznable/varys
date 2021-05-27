package corp

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var configCache *gokits.CacheTable
var tokenCache *gokits.CacheTable

func cacheInitialize() {
    configCache = gokits.CacheExpireAfterWrite("wechat.corp.config")
    configCache.SetDataLoader(configLoader)
    tokenCache = gokits.CacheExpireAfterWrite("wechat.corp.token")
    tokenCache.SetDataLoader(tokenLoader)
}

type WechatCorpConfig struct {
    CorpId     string
    CorpSecret string
}

func configLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return ConfigLoader(
        "Wechat Corp",
        &WechatCorpConfig{},
        queryConfigSQL,
        configLifeSpan,
        codeName, args...)
}

type WechatCorpToken struct {
    CorpId      string `json:"corpId"`
    AccessToken string `json:"token"`
}

type Response struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func tokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatCorpConfig)

    result, err := gokits.NewHttpReq(tokenURL).Params(
        "corpid", config.CorpId, "corpsecret", config.CorpSecret).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request Wechat Corp Token Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || 0 != response.Errcode || "" == response.AccessToken {
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

func tokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return TokenLoaderStrict(
        "Wechat Corp",
        &QueryWechatCorpToken{},
        queryTokenSQL,
        func(queryDest ExpireTimeRecord) interface{} {
            query := queryDest.(*QueryWechatCorpToken)
            return &WechatCorpToken{
                CorpId:      query.CorpId,
                AccessToken: query.AccessToken,
            }
        },
        tokenRequestor,
        createTokenSQL,
        updateTokenSQL,
        func(response map[string]string) map[string]interface{} {
            expireTime, _ := gokits.Int64FromStr(response["ExpireTime"])
            return map[string]interface{}{
                "AccessToken": response["AccessToken"],
                "ExpireTime":  expireTime}
        },
        tokenExpireCriticalSpan,
        func(response map[string]string) interface{} {
            return &WechatCorpToken{
                CorpId:      response["CorpId"],
                AccessToken: response["AccessToken"],
            }
        },
        tokenMaxLifeSpan,
        codeName, args...)
}
