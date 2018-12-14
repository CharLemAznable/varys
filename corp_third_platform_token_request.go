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
        "Request WechatCorpThirdPlatformToken Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SUITE_ID":     config.SuiteId,
        "ACCESS_TOKEN": response.SuiteAccessToken,
        "EXPIRE_TIME":  StrFromInt64(expireTime)}, nil
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

    result, err := httpreq.New(wechatCorpThirdPlatformPreAuthCodeURL + tokenItem.AccessToken).Get()
    log.Trace("Request WechatCorpThirdPlatformPreAuthCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatCorpThirdPlatformPreAuthCodeResponse)).
    (*WechatCorpThirdPlatformPreAuthCodeResponse)
    if nil == response || 0 == len(response.PreAuthCode) {
        return nil, &UnexpectedError{Message:
        "Request WechatCorpThirdPlatformPreAuthCode Failed: " + result}
    }
    return map[string]string{
        "SUITE_ID":      tokenItem.SuiteId,
        "PRE_AUTH_CODE": response.PreAuthCode,
        "EXPIRES_IN":    StrFromInt(response.ExpiresIn)}, nil
}

var wechatCorpThirdPlatformPermanentCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_permanent_code?suite_access_token="

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

    result, err := httpreq.New(wechatCorpThirdPlatformPermanentCodeURL + tokenItem.AccessToken).
        RequestBody(Json(map[string]string{"auth_code": authCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatCorpThirdPlatformPermanentCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatCorpThirdPlatformPermanentCodeResponse)).
    (*WechatCorpThirdPlatformPermanentCodeResponse)
    if nil == response || 0 == len(response.PermanentCode) {
        return nil, &UnexpectedError{Message:
            "Request WechatCorpThirdPlatformPermanentCode Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SUITE_ID":       tokenItem.SuiteId,
        "CORP_ID":        response.AuthCorpInfo.Corpid,
        "PERMANENT_CODE": response.PermanentCode,
        "ACCESS_TOKEN":   response.AccessToken,
        "EXPIRE_TIME":    StrFromInt64(expireTime)}, nil
}

var wechatCorpThirdPlatformCorpTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_corp_token?suite_access_token="

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

    result, err := httpreq.New(wechatCorpThirdPlatformCorpTokenURL + tokenItem.AccessToken).
        RequestBody(Json(map[string]string{
            "auth_corpid":    corpId.(string),
            "permanent_code": codeItem.PermanentCode})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatCorpThirdPlatformCorpToken Response:(%s, %s) %s", codeName, corpId, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatCorpThirdPlatformCorpTokenResponse)).
    (*WechatCorpThirdPlatformCorpTokenResponse)
    if nil == response || 0 == len(response.AccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatCorpThirdPlatformCorpToken Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SUITE_ID":          tokenItem.SuiteId,
        "CORP_ID":           corpId.(string),
        "CORP_ACCESS_TOKEN": response.AccessToken,
        "EXPIRE_TIME":       StrFromInt64(expireTime)}, nil
}
