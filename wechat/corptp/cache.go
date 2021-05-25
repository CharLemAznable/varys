package corptp

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/CharLemAznable/wechataes"
    "github.com/kataras/golog"
    "time"
)

var configCache *gokits.CacheTable
var cryptorCache *gokits.CacheTable
var tokenCache *gokits.CacheTable

func cacheInitialize() {
    configCache = gokits.CacheExpireAfterWrite("wechat.corptp.config")
    configCache.SetDataLoader(configLoader)
    cryptorCache = gokits.CacheExpireAfterWrite("wechat.corptp.cryptor")
    cryptorCache.SetDataLoader(cryptorLoader)
    tokenCache = gokits.CacheExpireAfterWrite("wechat.corptp.token")
    tokenCache.SetDataLoader(tokenLoader)
}

type WechatCorpTpConfig struct {
    SuiteId     string
    SuiteSecret string
    Token       string
    AesKey      string
    RedirectURL string
}

func configLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return base.ConfigLoader(
        "Wechat CorpTp",
        &WechatCorpTpConfig{},
        queryConfigSQL,
        configLifeSpan,
        codeName, args...)
}

func cryptorLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, errors.New("Require Wechat Corp Tp Config with key: " + codeName.(string)) // require config
    }
    config := cache.Data().(*WechatCorpTpConfig)
    golog.Debugf("Query Wechat Corp Tp Config Cache:(%s) %+v", codeName, config)

    cryptor, err := wechataes.NewWechatCryptor(config.SuiteId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    golog.Infof("Load Wechat Corp Tp Cryptor Cache:(%s) %+v", codeName, cryptor)
    return gokits.NewCacheItem(codeName, cryptorLifeSpan, cryptor), nil
}

type WechatCorpTpToken struct {
    SuiteId     string
    AccessToken string
}

type Response struct {
    Errcode          int    `json:"errcode"`
    Errmsg           string `json:"errmsg"`
    SuiteAccessToken string `json:"suite_access_token"`
    ExpiresIn        int    `json:"expires_in"`
}

func wechatCorpTpTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatCorpTpConfig)

    var ticket string
    err = base.DB.NamedGet(&ticket, queryTicketSQL,
        map[string]interface{}{"CodeName": codeName})
    if nil != err {
        return nil, err
    }

    result, err := gokits.NewHttpReq(tokenURL).
        RequestBody(gokits.Json(map[string]string{
            "suite_id":     config.SuiteId,
            "suite_secret": config.SuiteSecret,
            "suite_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request Wechat Corp Tp Token Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.SuiteAccessToken {
        return nil, errors.New("Request Wechat Corp Tp Token Failed: " + result)
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

func tokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return base.TokenLoaderStrict(
        "Wechat Corp Tp",
        &QueryWechatCorpTpToken{},
        queryTokenSQL,
        func(queryDest base.ExpireTimeRecord) interface{} {
            query := queryDest.(*QueryWechatCorpTpToken)
            return &WechatCorpTpToken{
                SuiteId:     query.SuiteId,
                AccessToken: query.AccessToken,
            }
        },
        wechatCorpTpTokenRequestor,
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
            return &WechatCorpTpToken{
                SuiteId:     response["SuiteId"],
                AccessToken: response["AccessToken"],
            }
        },
        tokenMaxLifeSpan,
        codeName, args...)
}
