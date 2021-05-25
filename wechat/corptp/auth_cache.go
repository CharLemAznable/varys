package corptp

import (
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var permanentCodeCache *gokits.CacheTable
var authTokenCache *gokits.CacheTable

func authCacheInitialize() {
    permanentCodeCache = gokits.CacheExpireAfterWrite("wechat.corptp.permanentcode")
    permanentCodeCache.SetDataLoader(permanentCodeLoader)
    authTokenCache = gokits.CacheExpireAfterWrite("wechat.corptp.auth.token")
    authTokenCache.SetDataLoader(authTokenLoader)
}

type WechatCorpTpAuthKey struct {
    CodeName string
    CorpId   string
}

type WechatCorpTpPermanentCode struct {
    SuiteId       string
    CorpId        string
    PermanentCode string
}

func permanentCodeLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    result := &WechatCorpTpPermanentCode{}
    err := base.DB.NamedGet(result, queryPermanentCodeSQL, key)
    if nil != err {
        return nil, errors.New(fmt.Sprintf("Unauthorized corp: %+v", key))
    }
    golog.Infof("Load Wechat Corp Tp PermanentCode Cache:(%+v) %+v", key, result)
    return gokits.NewCacheItem(key, permanentCodeLifeSpan, result), nil
}

type WechatCorpTpAuthToken struct {
    SuiteId         string `json:"suiteId"`
    CorpId          string `json:"corpId"`
    CorpAccessToken string `json:"token"`
}

type AuthResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func authRequestor(codeName, corpId interface{}) (map[string]string, error) {
    tokenCache, err := tokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := tokenCache.Data().(*WechatCorpTpToken)

    codeCache, err := permanentCodeCache.Value(
        WechatCorpTpAuthKey{CodeName: codeName.(string), CorpId: corpId.(string)})
    if nil != err {
        return nil, err
    }
    codeItem := codeCache.Data().(*WechatCorpTpPermanentCode)

    result, err := gokits.NewHttpReq(authTokenURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{
            "auth_corpid":    corpId.(string),
            "permanent_code": codeItem.PermanentCode})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request Wechat Corp Tp Auth Token Response:(%s, %s) %s", codeName, corpId, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(AuthResponse)).(*AuthResponse)
    if nil == response || "" == response.AccessToken {
        return nil, errors.New("Request Wechat Corp Tp Auth Token Failed: " + result)
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SuiteId":         tokenItem.SuiteId,
        "CorpId":          corpId.(string),
        "CorpAccessToken": response.AccessToken,
        "ExpireTime":      gokits.StrFromInt64(expireTime)}, nil
}

func wechatCorpTpAuthResponseTokenBuilder(response map[string]string) interface{} {
    return &WechatCorpTpAuthToken{
        SuiteId:         response["SuiteId"],
        CorpId:          response["CorpId"],
        CorpAccessToken: response["CorpAccessToken"],
    }
}

type QueryWechatCorpTpAuthToken struct {
    WechatCorpTpAuthToken
    ExpireTime int64
}

func wechatCorpTpAuthQueryTokenBuilder(query *QueryWechatCorpTpAuthToken) interface{} {
    return &WechatCorpTpAuthToken{
        SuiteId:         query.SuiteId,
        CorpId:          query.CorpId,
        CorpAccessToken: query.CorpAccessToken,
    }
}

func authTokenLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    tpAuthKey, ok := key.(WechatCorpTpAuthKey)
    if !ok {
        return nil, errors.New("WechatCorpTpAuthKey type error") // key type error
    }

    query := &QueryWechatCorpTpAuthToken{}
    err := base.DB.NamedGet(query, queryAuthTokenSQL, tpAuthKey)
    if nil != err {
        return nil, errors.New(fmt.Sprintf("Unauthorized corp: %+v", tpAuthKey)) // requires that the token already exists
    }
    golog.Debugf("Query Wechat Corp Tp Auth Token:(%+v) %+v", tpAuthKey, query)

    effectiveSpan := time.Duration(query.ExpireTime-time.Now().Unix()) * time.Second // in second
    // 即将过期 -> 触发更新
    if effectiveSpan <= authTokenExpireCriticalSpan {
        time.Sleep(authTokenExpireCriticalSpan) // 休眠后再请求最新的access_token
        golog.Debugf("Try to request and update Wechat Corp Tp Auth Token:(%+v)", tpAuthKey)

        response, err := authRequestor(tpAuthKey.CodeName, tpAuthKey.CorpId)
        if nil != err {
            golog.Warnf("Request Wechat Corp Tp Auth Token Failed:(%+v) %s", tpAuthKey, err.Error())
            return nil, err
        }

        expireTime, _ := gokits.Int64FromStr(response["ExpireTime"])
        count, err := base.DB.NamedExecX(updateAuthTokenSQL, map[string]interface{}{
            "CodeName": tpAuthKey.CodeName, "CorpId": tpAuthKey.CorpId,
            "AccessToken": response["CorpAccessToken"], "ExpireTime": expireTime})
        if nil != err || count < 1 { // 记录入库失败, 则查询记录并返回
            err := base.DB.NamedGet(query, queryAuthTokenSQL, tpAuthKey)
            if nil != err {
                return nil, errors.New(fmt.Sprintf("Query Wechat Corp Tp Auth Token:(%+v) Failed", tpAuthKey))
            }

            effectiveSpan := time.Duration(query.ExpireTime-time.Now().Unix()) * time.Second // in second
            if effectiveSpan <= authTokenExpireCriticalSpan {
                return nil, errors.New(fmt.Sprintf("Query Wechat Corp Tp Auth Token:(%+v) expireTime Failed", tpAuthKey))
            }

            // 查询记录成功, 缓存最大缓存时长
            token := wechatCorpTpAuthQueryTokenBuilder(query)
            golog.Infof("Request and ReQuery Wechat Corp Tp Auth Token:(%+v) %+v", tpAuthKey, token)
            return gokits.NewCacheItem(key, authTokenMaxLifeSpan, token), nil
        }

        // 记录入库成功, 缓存最大缓存时长
        token := wechatCorpTpAuthResponseTokenBuilder(response)
        golog.Infof("Request and Update Wechat Corp Tp Auth Token:(%+v) %+v", tpAuthKey, token)
        return gokits.NewCacheItem(key, authTokenMaxLifeSpan, token), nil
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长
    ls := gokits.Condition(effectiveSpan > authTokenMaxLifeSpan,
        authTokenMaxLifeSpan, effectiveSpan).(time.Duration)
    token := wechatCorpTpAuthQueryTokenBuilder(query)
    golog.Infof("Load Wechat Corp Tp Auth Token Cache:(%+v) %+v, cache %3.1f min", tpAuthKey, token, ls.Minutes())
    return gokits.NewCacheItem(key, ls, token), nil
}
