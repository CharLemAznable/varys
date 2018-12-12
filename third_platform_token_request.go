package varys

import (
    . "github.com/CharLemAznable/goutils"
    "github.com/CharLemAznable/httpreq"
    log "github.com/CharLemAznable/log4go"
)

var wechatThirdPlatformTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

type WechatThirdPlatformTokenResponse struct {
    ComponentAccessToken string `json:"component_access_token"`
    ExpiresIn            int    `json:"expires_in"`
}

func wechatThirdPlatformTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatThirdPlatformConfig)

    ticket, err := queryWechatThirdPlatformTicket(codeName.(string))
    if nil != err {
        return nil, err
    }

    result, err := httpreq.New(wechatThirdPlatformTokenURL).
        RequestBody(Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatThirdPlatformToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatThirdPlatformTokenResponse)).(*WechatThirdPlatformTokenResponse)
    if nil == response || 0 == len(response.ComponentAccessToken) {
        return nil, &UnexpectedError{Message:
        "Request component_access_token Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                 config.AppId,
        "COMPONENT_ACCESS_TOKEN": response.ComponentAccessToken,
        "EXPIRES_IN":             StrFromInt(response.ExpiresIn)}, nil
}

var wechatThirdPlatformPreAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="

type WechatThirdPlatformPreAuthCodeResponse struct {
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatThirdPlatformPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatThirdPlatformToken)

    result, err := httpreq.New(wechatThirdPlatformPreAuthCodeURL + tokenItem.ComponentAccessToken).
        RequestBody(Json(map[string]string{"component_appid": tokenItem.AppId})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatThirdPlatformPreAuthCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatThirdPlatformPreAuthCodeResponse)).(*WechatThirdPlatformPreAuthCodeResponse)
    if nil == response || 0 == len(response.PreAuthCode) {
        return nil, &UnexpectedError{Message:
        "Request pre_auth_code Failed: " + result}
    }
    return map[string]string{
        "APP_ID":        tokenItem.AppId,
        "PRE_AUTH_CODE": response.PreAuthCode,
        "EXPIRES_IN":    StrFromInt(response.ExpiresIn)}, nil
}

var wechatThirdPlatformQueryAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token="

type WechatThirdPlatformQueryAuthResponse struct {
    AuthorizationInfo WechatThirdPlatformQueryAuthInfo `json:"authorization_info"`
}

type WechatThirdPlatformQueryAuthInfo struct {
    AuthorizerAppid        string              `json:"authorizer_appid"`
    AuthorizerAccessToken  string              `json:"authorizer_access_token"`
    ExpiresIn              int                 `json:"expires_in"`
    AuthorizerRefreshToken string              `json:"authorizer_refresh_token"`
    FuncInfo               []FuncscopeCategory `json:"func_info"`
}

type FuncscopeCategory struct {
    Content FuncscopeCategoryContent `json:"funcscope_category"`
}

type FuncscopeCategoryContent struct {
    Id int `json:"id"`
}

func wechatThirdPlatformQueryAuthRequestor(codeName, authorizationCode interface{}) (map[string]string, error) {
    cache, err := wechatThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatThirdPlatformToken)

    result, err := httpreq.New(wechatThirdPlatformQueryAuthURL + tokenItem.ComponentAccessToken).
        RequestBody(Json(map[string]string{
            "component_appid":    tokenItem.AppId,
            "authorization_code": authorizationCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatThirdPlatformQueryAuth Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatThirdPlatformQueryAuthResponse)).(*WechatThirdPlatformQueryAuthResponse)
    if nil == response || 0 == len(response.AuthorizationInfo.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message:
        "Request authorizer_access_token Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         response.AuthorizationInfo.AuthorizerAppid,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizationInfo.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizationInfo.AuthorizerRefreshToken,
        "EXPIRES_IN":               StrFromInt(response.AuthorizationInfo.ExpiresIn)}, nil
}

var wechatThirdPlatformRefreshAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="

type WechatThirdPlatformRefreshAuthResponse struct {
    AuthorizerAccessToken  string `json:"authorizer_access_token"`
    ExpiresIn              int    `json:"expires_in"`
    AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

func wechatThirdPlatformRefreshAuthRequestor(codeName, authorizerAppId, authorizerRefreshToken string) (map[string]string, error) {
    cache, err := wechatThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatThirdPlatformToken)

    result, err := httpreq.New(wechatThirdPlatformRefreshAuthURL + tokenItem.ComponentAccessToken).
        RequestBody(Json(map[string]string{
            "component_appid":          tokenItem.AppId,
            "authorizer_appid":         authorizerAppId,
            "authorizer_refresh_token": authorizerRefreshToken})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatThirdPlatformRefreshAuth Response:(%s, %s) %s", codeName, authorizerAppId, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatThirdPlatformRefreshAuthResponse)).(*WechatThirdPlatformRefreshAuthResponse)
    if nil == response || 0 == len(response.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message:
        "Refresh authorizer_access_token Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         authorizerAppId,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizerRefreshToken,
        "EXPIRES_IN":               StrFromInt(response.ExpiresIn)}, nil
}