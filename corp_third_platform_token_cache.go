package main

import (
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "time"
)

var wechatCorpThirdPlatformConfigCache *gokits.CacheTable
var wechatCorpThirdPlatformCryptorCache *gokits.CacheTable
var wechatCorpThirdPlatformTokenCache *gokits.CacheTable
var wechatCorpThirdPlatformPermanentCodeCache *gokits.CacheTable
var wechatCorpThirdPlatformCorpTokenCache *gokits.CacheTable

func wechatCorpThirdPlatformAuthorizerTokenInitialize() {
    wechatCorpThirdPlatformConfigCache = gokits.CacheExpireAfterWrite("WechatCorpThirdPlatformConfig")
    wechatCorpThirdPlatformConfigCache.SetDataLoader(wechatCorpThirdPlatformConfigLoader)
    wechatCorpThirdPlatformCryptorCache = gokits.CacheExpireAfterWrite("WechatCorpThirdPlatformCryptor")
    wechatCorpThirdPlatformCryptorCache.SetDataLoader(wechatCorpThirdPlatformCryptorLoader)
    wechatCorpThirdPlatformTokenCache = gokits.CacheExpireAfterWrite("WechatCorpThirdPlatformCryptor")
    wechatCorpThirdPlatformTokenCache.SetDataLoader(wechatCorpThirdPlatformTokenLoader)
    wechatCorpThirdPlatformPermanentCodeCache = gokits.CacheExpireAfterWrite("WechatCorpThirdPlatformPermanentCode")
    wechatCorpThirdPlatformPermanentCodeCache.SetDataLoader(wechatCorpThirdPlatformPermanentCodeLoader)
    wechatCorpThirdPlatformCorpTokenCache = gokits.CacheExpireAfterWrite("WechatCorpThirdPlatformCorpToken")
    wechatCorpThirdPlatformCorpTokenCache.SetDataLoader(wechatCorpThirdPlatformCorpTokenLoader)
}

type WechatCorpThirdPlatformConfig struct {
    SuiteId     string
    SuiteSecret string
    Token       string
    AesKey      string
    RedirectURL string
}

func wechatCorpThirdPlatformConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return configLoader(
        "WechatCorpThirdPlatformConfig",
        queryWechatCorpThirdPlatformConfigSQL,
        wechatCorpThirdPlatformConfigLifeSpan,
        func(resultItem map[string]string) interface{} {
            config := new(WechatCorpThirdPlatformConfig)
            config.SuiteId = resultItem["SUITE_ID"]
            config.SuiteSecret = resultItem["SUITE_SECRET"]
            config.Token = resultItem["TOKEN"]
            config.AesKey = resultItem["AES_KEY"]
            config.RedirectURL = resultItem["REDIRECT_URL"]
            if 0 == len(config.SuiteId) || 0 == len(config.SuiteSecret) ||
                0 == len(config.Token) || 0 == len(config.AesKey) {
                return nil
            }
            return config
        },
        codeName, args...)
}

func wechatCorpThirdPlatformCryptorLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    cache, err := wechatCorpThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, &UnexpectedError{Message: "Require WechatCorpThirdPlatformConfig with key: " + codeName.(string)} // require config
    }
    config := cache.Data().(*WechatCorpThirdPlatformConfig)
    gokits.LOG.Trace("Query WechatCorpThirdPlatformConfig Cache:(%s) %s", codeName, gokits.Json(config))

    cryptor, err := wechataes.NewWechatCryptor(config.SuiteId, config.Token, config.AesKey)
    if nil != err {
        return nil, err // require legal config
    }
    gokits.LOG.Info("Load WechatCorpThirdPlatformCryptor Cache:(%s) %s", codeName, cryptor)
    return gokits.NewCacheItem(codeName, wechatCorpThirdPlatformCryptorLifeSpan, cryptor), nil
}

type WechatCorpThirdPlatformToken struct {
    SuiteId     string
    AccessToken string
}

func wechatCorpThirdPlatformTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatCorpThirdPlatformToken)
    tokenItem.SuiteId = resultItem["SUITE_ID"]
    tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
    return tokenItem
}

type WechatCorpThirdPlatformTokenResponse struct {
    Errcode          int    `json:"errcode"`
    Errmsg           string `json:"errmsg"`
    SuiteAccessToken string `json:"suite_access_token"`
    ExpiresIn        int    `json:"expires_in"`
}

func wechatCorpThirdPlatformTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatCorpThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatCorpThirdPlatformConfig)

    ticket, err := queryWechatCorpThirdPlatformTicket(codeName.(string))
    if nil != err {
        return nil, err
    }

    result, err := gokits.NewHttpReq(wechatCorpThirdPlatformTokenURL).
        RequestBody(gokits.Json(map[string]string{
            "suite_id":     config.SuiteId,
            "suite_secret": config.SuiteSecret,
            "suite_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatCorpThirdPlatformToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpThirdPlatformTokenResponse)).(*WechatCorpThirdPlatformTokenResponse)
    if nil == response || 0 == len(response.SuiteAccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatCorpThirdPlatformToken Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SUITE_ID":     config.SuiteId,
        "ACCESS_TOKEN": response.SuiteAccessToken,
        "EXPIRE_TIME":  gokits.StrFromInt64(expireTime)}, nil
}

func wechatCorpThirdPlatformTokenSQLParamBuilder(resultItem map[string]string, codeName interface{}) []interface{} {
    expireTime, _ := gokits.Int64FromStr(resultItem["EXPIRE_TIME"])
    return []interface{}{resultItem["ACCESS_TOKEN"], expireTime, codeName}
}

func wechatCorpThirdPlatformTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    return tokenLoaderStrict(
        "WechatCorpThirdPlatformToken",
        queryWechatCorpThirdPlatformTokenSQL,
        createWechatCorpThirdPlatformTokenSQL,
        updateWechatCorpThirdPlatformTokenSQL,
        wechatCorpThirdPlatformTokenMaxLifeSpan,
        wechatCorpThirdPlatformTokenExpireCriticalSpan,
        wechatCorpThirdPlatformTokenBuilder,
        wechatCorpThirdPlatformTokenRequestor,
        wechatCorpThirdPlatformTokenSQLParamBuilder,
        codeName, args...)
}

type WechatCorpThirdPlatformPermanentCodeResponse struct {
    Errcode       int          `json:"errcode"`
    Errmsg        string       `json:"errmsg"`
    AccessToken   string       `json:"access_token"`
    ExpiresIn     int          `json:"expires_in"`
    PermanentCode string       `json:"permanent_code"`
    AuthCorpInfo  AuthCorpInfo `json:"auth_corp_info"`
    AuthInfo      AuthInfo     `json:"auth_info"`
    AuthUserInfo  AuthUserInfo `json:"auth_user_info"`
}

type AuthCorpInfo struct {
    Corpid            string `json:"corpid"`
    CorpName          string `json:"corp_name"`
    CorpType          string `json:"corp_type"`
    CorpSquareLogoUrl string `json:"corp_square_logo_url"`
    CorpUserMax       int    `json:"corp_user_max"`
    CorpAgentMax      int    `json:"corp_agent_max"`
    CorpFullName      string `json:"corp_full_name"`
    VerifiedEndTime   int64  `json:"verified_end_time"`
    SubjectType       int    `json:"subject_type"`
    CorpWxqrcode      string `json:"corp_wxqrcode"`
    CorpScale         string `json:"corp_scale"`
    CorpIndustry      string `json:"corp_industry"`
    CorpSubIndustry   string `json:"corp_sub_industry"`
    Location          string `json:"location"`
}

type AuthInfo struct {
    Agent []Agent `json:"agent"`
}

type Agent struct {
    Agentid       int64     `json:"agentid"`
    Name          string    `json:"name"`
    RoundLogoUrl  string    `json:"round_logo_url"`
    SquareLogoUrl string    `json:"square_logo_url"`
    Appid         int64     `json:"appid"`
    Privilege     Privilege `json:"privilege"`
}

type Privilege struct {
    Level      int      `json:"level"`
    AllowParty []int    `json:"allow_party"`
    AllowUser  []string `json:"allow_user"`
    AllowTag   []int    `json:"allow_tag"`
    ExtraParty []int    `json:"extra_party"`
    ExtraUser  []string `json:"extra_user"`
    ExtraTag   []int    `json:"extra_tag"`
}

type AuthUserInfo struct {
    Userid string `json:"userid"`
    Name   string `json:"name"`
    Avatar string `json:"avatar"`
}

func wechatCorpThirdPlatformPermanentCodeRequestor(codeName, authCode interface{}) (map[string]string, error) {
    cache, err := wechatCorpThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatCorpThirdPlatformToken)

    result, err := gokits.NewHttpReq(wechatCorpThirdPlatformPermanentCodeURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{"auth_code": authCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatCorpThirdPlatformPermanentCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpThirdPlatformPermanentCodeResponse)).(*WechatCorpThirdPlatformPermanentCodeResponse)
    if nil == response || 0 == len(response.PermanentCode) {
        return nil, &UnexpectedError{Message: "Request WechatCorpThirdPlatformPermanentCode Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SUITE_ID":       tokenItem.SuiteId,
        "CORP_ID":        response.AuthCorpInfo.Corpid,
        "PERMANENT_CODE": response.PermanentCode,
        "ACCESS_TOKEN":   response.AccessToken,
        "EXPIRE_TIME":    gokits.StrFromInt64(expireTime)}, nil
}

func wechatCorpThirdPlatformAuthorizeCreator(codeName, authCode interface{}) {
    resultItem, err := wechatCorpThirdPlatformPermanentCodeRequestor(codeName, authCode)
    if nil != err {
        _ = gokits.LOG.Warn("Request WechatCorpThirdPlatformPermanentCode Failed:(%s, authCode:%s) %s",
            codeName, authCode, err.Error())
        return
    }

    corpId := resultItem["CORP_ID"]
    _, _ = enableWechatCorpThirdPlatformAuthorizer(codeName.(string), corpId, resultItem["PERMANENT_CODE"])

    accessToken := resultItem["ACCESS_TOKEN"]
    expireTime := resultItem["EXPIRE_TIME"]
    _, err = db.New().Sql(createWechatCorpThirdPlatformCorpTokenSQL).
        Params(corpId, accessToken, expireTime, codeName).Execute()
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        _ = gokits.LOG.Warn("Create WechatCorpThirdPlatformCorpToken Failed:(%s, corpId:%s) %s",
            codeName, corpId, err.Error())

        _, _ = db.New().Sql(updateWechatCorpThirdPlatformCorpTokenSQL).
            Params(accessToken, expireTime, codeName, corpId).Execute()
        // 忽略更新记录的结果
        // 如果当前存在有效期内的token, 则token不会被更新, 重复请求微信也会返回同样的token
    }
}

type WechatCorpThirdPlatformAuthorizerKey struct {
    CodeName string
    CorpId   string
}

type WechatCorpThirdPlatformPermanentCode struct {
    SuiteId       string
    CorpId        string
    PermanentCode string
}

func wechatCorpThirdPlatformPermanentCodeLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    corpKey, ok := key.(WechatCorpThirdPlatformAuthorizerKey)
    if !ok {
        return nil, &UnexpectedError{Message: "WechatCorpThirdPlatformAuthorizerKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatCorpThirdPlatformPermanentCodeSQL).
        Params(corpKey.CodeName, corpKey.CorpId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Unauthorized corp: " + gokits.Json(key)}).(error) // requires that the permanent code already exists
    }
    gokits.LOG.Trace("Query WechatCorpThirdPlatformPermanentCode:(%s) %s", gokits.Json(key), resultMap)

    resultItem := resultMap[0]
    codeItem := new(WechatCorpThirdPlatformPermanentCode)
    codeItem.SuiteId = resultItem["SUITE_ID"]
    codeItem.CorpId = resultItem["CORP_ID"]
    codeItem.PermanentCode = resultItem["PERMANENT_CODE"]
    gokits.LOG.Info("Load WechatCorpThirdPlatformPermanentCode Cache:(%s) %s", gokits.Json(key), gokits.Json(codeItem))
    return gokits.NewCacheItem(key, wechatCorpThirdPlatformPermanentCodeLifeSpan, codeItem), nil
}

type WechatCorpThirdPlatformCorpToken struct {
    SuiteId         string
    CorpId          string
    CorpAccessToken string
}

func wechatCorpThirdPlatformCorpTokenBuilder(resultItem map[string]string) interface{} {
    tokenItem := new(WechatCorpThirdPlatformCorpToken)
    tokenItem.SuiteId = resultItem["SUITE_ID"]
    tokenItem.CorpId = resultItem["CORP_ID"]
    tokenItem.CorpAccessToken = resultItem["CORP_ACCESS_TOKEN"]
    return tokenItem
}

type WechatCorpThirdPlatformCorpTokenResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpThirdPlatformCorpTokenRequestor(codeName, corpId interface{}) (map[string]string, error) {
    tokenCache, err := wechatCorpThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := tokenCache.Data().(*WechatCorpThirdPlatformToken)

    codeCache, err := wechatCorpThirdPlatformPermanentCodeCache.Value(
        WechatCorpThirdPlatformAuthorizerKey{CodeName: codeName.(string), CorpId: corpId.(string)})
    if nil != err {
        return nil, err
    }
    codeItem := codeCache.Data().(*WechatCorpThirdPlatformPermanentCode)

    result, err := gokits.NewHttpReq(wechatCorpThirdPlatformCorpTokenURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{
            "auth_corpid":    corpId.(string),
            "permanent_code": codeItem.PermanentCode})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatCorpThirdPlatformCorpToken Response:(%s, %s) %s", codeName, corpId, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpThirdPlatformCorpTokenResponse)).(*WechatCorpThirdPlatformCorpTokenResponse)
    if nil == response || 0 == len(response.AccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatCorpThirdPlatformCorpToken Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SUITE_ID":          tokenItem.SuiteId,
        "CORP_ID":           corpId.(string),
        "CORP_ACCESS_TOKEN": response.AccessToken,
        "EXPIRE_TIME":       gokits.StrFromInt64(expireTime)}, nil
}

func wechatCorpThirdPlatformCorpTokenLoader(key interface{}, args ...interface{}) (*gokits.CacheItem, error) {
    corpKey, ok := key.(WechatCorpThirdPlatformAuthorizerKey)
    if !ok {
        return nil, &UnexpectedError{Message: "WechatCorpThirdPlatformAuthorizerKey type error"} // key type error
    }

    resultMap, err := db.New().Sql(queryWechatCorpThirdPlatformCorpTokenSQL).
        Params(corpKey.CodeName, corpKey.CorpId).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Unauthorized corp: " + gokits.Json(key)}).(error) // requires that the token already exists
    }
    gokits.LOG.Trace("Query WechatCorpThirdPlatformCorpToken:(%s) %s", gokits.Json(key), resultMap)

    resultItem := resultMap[0]
    expireTime, err := gokits.Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
    if nil != err {
        return nil, err
    }
    effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
    // 即将过期 -> 触发更新
    if effectiveSpan <= wechatCorpThirdPlatformCorpTokenExpireCriticalSpan {
        time.Sleep(wechatCorpThirdPlatformCorpTokenExpireCriticalSpan) // 休眠后再请求最新的access_token
        gokits.LOG.Info("Try to request and update WechatCorpThirdPlatformCorpToken:(%s)", gokits.Json(key))

        resultItem, err := wechatCorpThirdPlatformCorpTokenRequestor(corpKey.CodeName, corpKey.CorpId)
        if nil != err {
            _ = gokits.LOG.Warn("Request WechatCorpThirdPlatformCorpToken Failed:(%s) %s", key, err.Error())
            return nil, err
        }

        expireTime, _ := gokits.Int64FromStr(resultItem["EXPIRE_TIME"])
        count, err := db.New().Sql(updateWechatCorpThirdPlatformCorpTokenSQL).
            Params(resultItem["CORP_ACCESS_TOKEN"], expireTime, corpKey.CodeName, corpKey.CorpId).Execute()
        if nil != err || count < 1 { // 记录入库失败, 则查询记录并返回
            resultMap, err := db.New().Sql(queryWechatCorpThirdPlatformCorpTokenSQL).
                Params(corpKey.CodeName, corpKey.CorpId).Query()
            if nil != err || 1 != len(resultMap) {
                return nil, gokits.DefaultIfNil(err, &UnexpectedError{Message: fmt.Sprintf("Query WechatCorpThirdPlatformCorpToken:(%s) Failed", gokits.Json(key))}).(error)
            }

            resultItem := resultMap[0]
            expireTime, err := gokits.Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
            if nil != err {
                return nil, err
            }
            effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
            if effectiveSpan <= wechatCorpThirdPlatformCorpTokenExpireCriticalSpan {
                return nil, &UnexpectedError{Message: fmt.Sprintf("Query WechatCorpThirdPlatformCorpToken:(%s) expireTime Failed", gokits.Json(key))}
            }

            // 查询记录成功, 缓存最大缓存时长
            tokenItem := wechatCorpThirdPlatformCorpTokenBuilder(resultItem)
            gokits.LOG.Info("Request and ReQuery WechatCorpThirdPlatformCorpToken:(%s) %s", gokits.Json(key), gokits.Json(tokenItem))
            return gokits.NewCacheItem(key, wechatCorpThirdPlatformCorpTokenMaxLifeSpan, tokenItem), nil
        }

        // 记录入库成功, 缓存最大缓存时长
        tokenItem := wechatCorpThirdPlatformCorpTokenBuilder(resultItem)
        gokits.LOG.Info("Request and Update WechatCorpThirdPlatformCorpToken:(%s) %s", gokits.Json(key), gokits.Json(tokenItem))
        return gokits.NewCacheItem(key, wechatCorpThirdPlatformCorpTokenMaxLifeSpan, tokenItem), nil
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长
    ls := gokits.Condition(effectiveSpan > wechatCorpThirdPlatformCorpTokenMaxLifeSpan,
        wechatCorpThirdPlatformCorpTokenMaxLifeSpan, effectiveSpan).(time.Duration)
    tokenItem := wechatCorpThirdPlatformCorpTokenBuilder(resultItem)
    gokits.LOG.Info("Load WechatCorpThirdPlatformCorpToken Cache:(%s) %s, cache %3.1f min",
        gokits.Json(key), gokits.Json(tokenItem), ls.Minutes())
    return gokits.NewCacheItem(key, ls, tokenItem), nil
}
