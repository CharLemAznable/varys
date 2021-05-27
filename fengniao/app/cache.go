package app

import (
    "crypto/sha256"
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var configCache *gokits.CacheTable
var tokenCache *gokits.CacheTable

func cacheInitialize() {
    configCache = gokits.CacheExpireAfterWrite("fengniao.app.config")
    configCache.SetDataLoader(configLoader)
    tokenCache = gokits.CacheExpireAfterWrite("fengniao.app.token")
    tokenCache.SetDataLoader(tokenLoader)
}

type FengniaoAppConfig struct {
    DevId       string
    AppId       string
    AppSecret   string
    CallbackUrl string
}

func configLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return base.ConfigLoader(
        "Fengniao App",
        &FengniaoAppConfig{},
        queryConfigSQL,
        configLifeSpan,
        codeName, args...)
}

type Response struct {
    Sign string               `json:"sign"`
    Code string               `json:"code"`
    Msg  string               `json:"msg"`
    Data ResponseBusinessData `json:"business_data"`
}

type ResponseBusinessData struct {
    AppId        string `json:"app_id"`
    MerchantId   string `json:"merchant_id"`
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpireIn     string `json:"expire_in"`
    ReExpireIn   string `json:"re_expire_in"`
}

func tokenCreator(codeName string, config *FengniaoAppConfig, callback *AuthCallbackRequest) {
    appId := config.AppId
    code := callback.Code
    merchantId := callback.MerchantId

    params := make(map[string]string)
    params["grant_type"] = "authorization_code"
    params["code"] = code
    params["app_id"] = appId
    params["merchant_id"] = merchantId
    timestamp := gokits.StrFromInt64(time.Now().UnixNano() / 1e6)
    params["timestamp"] = timestamp
    plainText := config.AppSecret + "app_id=" + appId +
        "&code=" + code + "&grant_type=authorization_code" +
        "&merchant_id=" + merchantId + "&timestamp=" + timestamp
    signature := fmt.Sprintf("%x", sha256.Sum256([]byte(plainText)))
    params["signature"] = signature

    result, err := gokits.NewHttpReq(tokenURL).
        RequestBody(gokits.Json(params)).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request Fengniao App Token Response:(%s %s) %s",
        codeName, merchantId, result)
    if nil != err {
        golog.Errorf("Request Fengniao App Token Failed: %s", err.Error())
        return
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.Data.AccessToken {
        golog.Errorf("Request Fengniao App Token Failed: %s", err.Error())
        return
    }

    // 剩余有效时间, 单位: 秒
    expireIn, _ := gokits.IntFromStr(response.Data.ExpireIn)
    // 剩余有效时间, 单位: 秒
    reExpireIn, _ := gokits.IntFromStr(response.Data.ReExpireIn)
    // 更新提前时间: token缓存时长 * 缓存提前更新系数(1.1)
    updateAhead := int(tokenLifeSpan.Seconds() * 1.1)
    _, err = base.DB.NamedExecX(createTokenSQL,
        map[string]interface{}{
            "CodeName":     codeName,
            "MerchantId":   merchantId,
            "Code":         code,
            "AccessToken":  response.Data.AccessToken,
            "RefreshToken": response.Data.RefreshToken,
            "ExpireIn":     expireIn - updateAhead,
            "ReExpireIn":   reExpireIn - updateAhead})
    if nil != err {
        golog.Errorf("Create Fengniao App Token Failed: %s", err.Error())
    }
}

type FengniaoAppTokenKey struct {
    CodeName   string
    MerchantId string
}

type FengniaoAppToken struct {
    AppId       string `json:"appId"`
    MerchantId  string `json:"merchantId"`
    AccessToken string `json:"token"`
}

type QueryFengniaoAppToken struct {
    FengniaoAppToken
    ExpireTime   int64
    RefreshToken string
    ReExpireTime int64
}

func tokenLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    tokenKey, ok := key.(FengniaoAppTokenKey)
    if !ok {
        return nil, errors.New("FengniaoAppTokenKey type error") // key type error
    }

    query := &QueryFengniaoAppToken{}
    err := base.DB.NamedGet(query, queryTokenSQL, tokenKey)
    if nil != err {
        return nil, errors.New(fmt.Sprintf("Unauthorized merchant: %+v", tokenKey)) // requires that the token already exists
    }
    golog.Debugf("Query Fengniao App Token:(%+v) %+v", tokenKey, query)

    isExpired := time.Now().Unix() >= query.ExpireTime
    if isExpired { // 已过期 -> 触发更新
        // 该接口10分钟有效期，如果10分钟内多次调用，
        // 只有第一次会刷新access_token和refresh_token，
        // 后续调用会返回第一次刷新的access_token和refresh_token。
        golog.Debugf("Try to request and update Fengniao App Token:(%+v)", tokenKey)
        token, err := tokenRefresher(tokenKey, query.RefreshToken)
        if nil != err {
            return nil, err
        }
        golog.Infof("Request and Update Fengniao App Token:(%+v) %+v", tokenKey, token)
        return gokits.NewCacheItem(tokenKey, tokenLifeSpan, token), nil
    }

    effectiveSpan := time.Duration(query.ExpireTime-time.Now().Unix()) * time.Second // in second
    // token有效期少于缓存时长, 则仅缓存剩余有效期时长
    ls := gokits.Condition(effectiveSpan > tokenLifeSpan, tokenLifeSpan, effectiveSpan).(time.Duration)
    token := &FengniaoAppToken{AppId: query.AppId, MerchantId: query.MerchantId, AccessToken: query.AccessToken}
    golog.Infof("Load Fengniao App Token Cache:(%+v) %+v, cache %3.1f min", tokenKey, token, ls.Minutes())
    return gokits.NewCacheItem(tokenKey, ls, token), nil
}

func tokenRefresher(tokenKey FengniaoAppTokenKey, refreshToken string) (*FengniaoAppToken, error) {
    codeName := tokenKey.CodeName
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, errors.New("CodeName is Illegal")
    }
    config := cache.Data().(*FengniaoAppConfig)
    appId := config.AppId
    merchantId := tokenKey.MerchantId

    params := make(map[string]string)
    params["grant_type"] = "refresh_token"
    params["app_id"] = appId
    params["merchant_id"] = merchantId
    timestamp := gokits.StrFromInt64(time.Now().UnixNano() / 1e6)
    params["timestamp"] = timestamp
    params["refresh_token"] = refreshToken
    plainText := config.AppSecret + "app_id=" + appId +
        "&grant_type=refresh_token" + "&merchant_id=" + merchantId +
        "&refresh_token" + refreshToken + "&timestamp=" + timestamp
    signature := fmt.Sprintf("%x", sha256.Sum256([]byte(plainText)))
    params["signature"] = signature

    result, err := gokits.NewHttpReq(refreshURL).
        RequestBody(gokits.Json(params)).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Refresh Fengniao App Token Response:(%s %s) %s",
        codeName, merchantId, result)
    if nil != err {
        return nil, errors.New("Refresh Fengniao App Token Failed: " + err.Error())
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.Data.AccessToken {
        return nil, errors.New("Refresh Fengniao App Token Failed: " + err.Error())
    }

    // 剩余有效时间, 单位: 秒
    expireIn, _ := gokits.IntFromStr(response.Data.ExpireIn)
    // 剩余有效时间, 单位: 秒
    reExpireIn, _ := gokits.IntFromStr(response.Data.ReExpireIn)
    // 更新提前时间: token缓存时长 * 缓存提前更新系数(1.1)
    updateAhead := int(tokenLifeSpan.Seconds() * 1.1)
    _, err = base.DB.NamedExecX(updateTokenSQL,
        map[string]interface{}{
            "CodeName":     codeName,
            "MerchantId":   merchantId,
            "AccessToken":  response.Data.AccessToken,
            "RefreshToken": response.Data.RefreshToken,
            "ExpireIn":     expireIn - updateAhead,
            "ReExpireIn":   reExpireIn - updateAhead})
    if nil != err {
        return nil, errors.New("Update Fengniao App Token Failed: " + err.Error())
    }

    return &FengniaoAppToken{AppId: appId, MerchantId: merchantId,
        AccessToken: response.Data.AccessToken}, nil
}
