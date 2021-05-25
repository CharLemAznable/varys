package apptp

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
    configCache = gokits.CacheExpireAfterWrite("wechat.tp.config")
    configCache.SetDataLoader(configLoader)
    cryptorCache = gokits.CacheExpireAfterWrite("wechat.tp.cryptor")
    cryptorCache.SetDataLoader(cryptorLoader)
    tokenCache = gokits.CacheExpireAfterWrite("wechat.tp.token")
    tokenCache.SetDataLoader(tokenLoader)
}

type WechatTpConfig struct {
    AppId          string
    AppSecret      string
    Token          string
    AesKey         string
    RedirectURL    string
    AuthForwardUrl string
    MsgForwardUrl  string
}

func configLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return base.ConfigLoader(
        "Wechat Tp",
        &WechatTpConfig{},
        queryConfigSQL,
        configLifeSpan,
        codeName, args...)
}

func cryptorLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, errors.New("Require Wechat Tp Config with key: " + codeName.(string)) // require config
    }
    config := cache.Data().(*WechatTpConfig)
    golog.Debugf("Query Wechat Tp Config Cache:(%s) %+v", codeName, config)

    cryptor, err := wechataes.NewWechatCryptor(config.AppId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    golog.Infof("Load Wechat Tp Cryptor Cache:(%s) %+v", codeName, cryptor)
    return gokits.NewCacheItem(codeName, cryptorLifeSpan, cryptor), nil
}

type WechatTpToken struct {
    AppId       string `json:"appId"`
    AccessToken string `json:"token"`
}

type Response struct {
    ComponentAccessToken string `json:"component_access_token"`
    ExpiresIn            int    `json:"expires_in"`
}

func tokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatTpConfig)

    var ticket string
    err = base.DB.NamedGet(&ticket, queryTicketSQL,
        map[string]interface{}{"CodeName": codeName})
    if nil != err {
        return nil, err
    }

    result, err := gokits.NewHttpReq(tokenURL).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request Wechat Tp Token Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.ComponentAccessToken {
        return nil, errors.New("Request Wechat Tp Token Failed: " + result)
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
func tokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return base.TokenLoader(
        "Wechat Tp",
        &QueryWechatTpToken{},
        queryTokenSQL,
        func(queryDest base.UpdatedRecord) interface{} {
            query := queryDest.(*QueryWechatTpToken)
            return &WechatTpToken{
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
            return &WechatTpToken{
                AppId:       response["AppId"],
                AccessToken: response["AccessToken"],
            }
        },
        tokenLifeSpan,
        tokenTempLifeSpan,
        codeName, args...)
}
