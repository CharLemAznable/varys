package app

import (
    "crypto/md5"
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var configCache *gokits.CacheTable
var tokenCache *gokits.CacheTable

func cacheInitialize() {
    configCache = gokits.CacheExpireAfterWrite("shansong.app.config")
    configCache.SetDataLoader(configLoader)
    tokenCache = gokits.CacheExpireAfterWrite("shansong.app.token")
    tokenCache.SetDataLoader(tokenLoader)
}

type ShansongAppConfig struct {
    AppId       string
    AppSecret   string
    RedirectURL string
    CallbackURL string
}

func configLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return ConfigLoader(
        "Shansong App",
        &ShansongAppConfig{},
        queryConfigSQL,
        configLifeSpan,
        codeName, args...)
}

type Response struct {
    Status int          `json:"status"`
    Msg    string       `json:"msg"`
    Data   ResponseData `json:"data"`
}

type ResponseData struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
}

func tokenCreator(codeName, appId, merchantCode, authCode string) {
    result, err := gokits.NewHttpReq(tokenURL).
        Params("clientId", appId, "code", authCode).Post()
    golog.Debugf("Request Shansong App Token Response:(%s %s) %s",
        codeName, merchantCode, result)
    if nil != err {
        golog.Errorf("Request Shansong App Token Failed: %s", err.Error())
        return
    }

    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.Data.AccessToken {
        golog.Errorf("Request Shansong App Token Failed: %s", err.Error())
        return
    }

    _, err = DB.NamedExecX(createTokenSQL,
        map[string]interface{}{
            "CodeName":     codeName,
            "MerchantCode": merchantCode,
            "Code":         authCode,
            "AccessToken":  response.Data.AccessToken,
            // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
            "ExpireIn":     response.Data.ExpiresIn - int(tokenLifeSpan.Seconds()*1.1),
            "RefreshToken": response.Data.RefreshToken})
    if nil != err {
        golog.Errorf("Create Shansong App Token Failed: %s", err.Error())
    }
}

type ShansongAppTokenKey struct {
    CodeName     string
    MerchantCode string
}

type ShansongAppToken struct {
    AppId        string `json:"appId"`
    MerchantCode string `json:"merchantCode"`
    AccessToken  string `json:"token"`
}

type QueryShansongAppToken struct {
    ShansongAppToken
    Updated      string
    ExpireTime   int64
    RefreshToken string
}

func (q *QueryShansongAppToken) GetUpdated() string {
    return q.Updated
}

func (q *QueryShansongAppToken) GetExpireTime() int64 {
    return q.ExpireTime
}

func tokenLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    tokenKey, ok := key.(ShansongAppTokenKey)
    if !ok {
        return nil, errors.New("ShansongAppTokenKey type error") // key type error
    }

    query := &QueryShansongAppToken{}
    err := DB.NamedGet(query, queryTokenSQL, tokenKey)
    if nil != err {
        return nil, errors.New(fmt.Sprintf("Unauthorized merchant: %+v", tokenKey)) // requires that the token already exists
    }
    golog.Debugf("Query Shansong App Token:(%+v) %+v", tokenKey, query)
    appId := query.AppId

    isExpired := time.Now().Unix() >= query.GetExpireTime()
    isUpdated := "1" == query.GetUpdated()
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        golog.Debugf("Try to request and update Shansong App Token:(%+v)", tokenKey)
        count, err := DB.NamedExecX(updateTokenSQL, tokenKey)
        if nil == err && count > 0 {
            token, err := tokenUpdater(tokenKey, appId, query.RefreshToken)
            if nil != err {
                return nil, err
            }
            golog.Infof("Request and Update Shansong App Token:(%+v) %+v", tokenKey, token)
            return gokits.NewCacheItem(key, tokenLifeSpan, token), nil
        }
        golog.Warnf("Give up request and update Shansong App Token:(%+v), use Query result Temporarily", tokenKey)
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := gokits.Condition(isExpired, tokenTempLifeSpan, tokenLifeSpan).(time.Duration)
    token := &ShansongAppToken{AppId: query.AppId,
        MerchantCode: query.MerchantCode,
        AccessToken:  query.AccessToken}
    golog.Infof("Load Shansong App Token Cache:(%+v) %+v, cache %3.1f min", tokenKey, token, ls.Minutes())
    return gokits.NewCacheItem(key, ls, token), nil
}

func tokenUpdater(tokenKey ShansongAppTokenKey, appId, refreshToken string) (*ShansongAppToken, error) {
    response, err := tokenRefresher(tokenKey, refreshToken)
    if nil != err {
        _, _ = DB.NamedExec(uncompleteTokenSQL, tokenKey)
        return nil, err
    }

    _, err = DB.NamedExecX(completeTokenSQL,
        map[string]interface{}{
            "CodeName":     tokenKey.CodeName,
            "MerchantCode": tokenKey.MerchantCode,
            "AccessToken":  response.Data.AccessToken,
            // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
            "ExpireIn":     response.Data.ExpiresIn - int(tokenLifeSpan.Seconds()*1.1)})
    if nil != err {
        return nil, errors.New("Update Shansong App Token Failed: " + err.Error())
    }

    return &ShansongAppToken{AppId: appId,
        MerchantCode: tokenKey.MerchantCode,
        AccessToken:  response.Data.AccessToken}, nil
}

func tokenRefresher(tokenKey ShansongAppTokenKey, refreshToken string) (*Response, error) {
    codeName := tokenKey.CodeName
    cache, err := configCache.Value(codeName)
    if nil != err {
        return nil, errors.New("CodeName is Illegal")
    }
    config := cache.Data().(*ShansongAppConfig)
    appId := config.AppId
    appSecret := config.AppSecret

    data := gokits.Json(map[string]string{"refreshToken": refreshToken})
    timestamp := gokits.StrFromInt64(time.Now().UnixNano() / 1e6)
    plainText := appSecret + "clientId" + appId + "data" + data + "timestamp" + timestamp
    signature := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))

    result, err := gokits.NewHttpReq(refreshTokenURL).Params(
        "clientId", appId, "data", data, "timestamp", timestamp, "sign", signature).Post()
    golog.Debugf("Refresh Shansong App Token Response:(%+v) %s", tokenKey, result)
    if nil != err {
        return nil, errors.New("Refresh Shansong App Token Failed: " + err.Error())
    }
    response := gokits.UnJson(result, new(Response)).(*Response)
    if nil == response || "" == response.Data.AccessToken {
        return nil, errors.New("Refresh Shansong App Token Failed: " + err.Error())
    }
    return response, nil
}
