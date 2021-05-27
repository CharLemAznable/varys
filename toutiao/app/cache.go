package app

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
    configCache = gokits.CacheExpireAfterWrite("toutiao.app.config")
    configCache.SetDataLoader(configLoader)
    tokenCache = gokits.CacheExpireAfterWrite("toutiao.app.token")
    tokenCache.SetDataLoader(tokenLoader)
}

type ToutiaoAppConfig struct {
    AppId     string
    AppSecret string
}

func configLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return ConfigLoader(
        "Toutiao App",
        &ToutiaoAppConfig{},
        queryConfigSQL,
        configLifeSpan,
        codeName, args...)
}

type ToutiaoAppToken struct {
    AppId       string `json:"appId"`
    AccessToken string `json:"token"`
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
    config := cache.Data().(*ToutiaoAppConfig)

    result, err := gokits.NewHttpReq(tokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request Toutiao App Token Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.AccessToken {
        return nil, errors.New("Request Toutiao App Token Failed: " + result)
    }
    return map[string]string{
        "AppId":       config.AppId,
        "AccessToken": response.AccessToken,
        "ExpiresIn":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

type QueryToutiaoAppToken struct {
    ToutiaoAppToken
    Updated    string
    ExpireTime int64
}

func (q *QueryToutiaoAppToken) GetUpdated() string {
    return q.Updated
}

func (q *QueryToutiaoAppToken) GetExpireTime() int64 {
    return q.ExpireTime
}

func tokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return TokenLoader(
        "Toutiao App",
        &QueryToutiaoAppToken{},
        queryTokenSQL,
        func(queryDest UpdatedRecord) interface{} {
            query := queryDest.(*QueryToutiaoAppToken)
            return &ToutiaoAppToken{
                AppId:       query.AppId,
                AccessToken: query.AccessToken,
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
        tokenLifeSpan,
        tokenTempLifeSpan,
        codeName, args...)
}
