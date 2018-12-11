package varys

import (
    . "github.com/CharLemAznable/goutils"
    "github.com/CharLemAznable/httpreq"
    log "github.com/CharLemAznable/log4go"
    "time"
)

var wechatCorpThirdPlatformTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_suite_token"

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

    result, err := httpreq.New(wechatCorpThirdPlatformTokenURL).
        RequestBody(Json(map[string]string{
            "suite_id":     config.SuiteId,
            "suite_secret": config.SuiteSecret,
            "suite_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatCorpThirdPlatformToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatCorpThirdPlatformTokenResponse)).
    (*WechatCorpThirdPlatformTokenResponse)
    if nil == response || 0 == len(response.SuiteAccessToken) {
        return nil, &UnexpectedError{Message:
        "Request suite_access_token Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SUITE_ID":           config.SuiteId,
        "SUITE_ACCESS_TOKEN": response.SuiteAccessToken,
        "EXPIRE_TIME":        StrFromInt64(expireTime)}, nil
}

var wechatCorpThirdPlatformPreAuthCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_pre_auth_code?suite_access_token="

type WechatCorpThirdPlatformPreAuthCodeResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpThirdPlatformPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatCorpThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatCorpThirdPlatformToken)

    result, err := httpreq.New(wechatCorpThirdPlatformPreAuthCodeURL + tokenItem.SuiteAccessToken).Get()
    log.Trace("Request WechatCorpThirdPlatformPreAuthCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatCorpThirdPlatformPreAuthCodeResponse)).
    (*WechatCorpThirdPlatformPreAuthCodeResponse)
    if nil == response || 0 == len(response.PreAuthCode) {
        return nil, &UnexpectedError{Message:
        "Request corp pre_auth_code Failed: " + result}
    }
    return map[string]string{
        "SUITE_ID":      tokenItem.SuiteId,
        "PRE_AUTH_CODE": response.PreAuthCode,
        "EXPIRES_IN":    StrFromInt(response.ExpiresIn)}, nil
}
