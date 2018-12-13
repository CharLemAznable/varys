package varys

import (
    . "github.com/CharLemAznable/goutils"
    "github.com/CharLemAznable/httpreq"
    log "github.com/CharLemAznable/log4go"
)

var wechatAppThirdPlatformTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

type WechatAppThirdPlatformTokenResponse struct {
    ComponentAccessToken string `json:"component_access_token"`
    ExpiresIn            int    `json:"expires_in"`
}

func wechatAppThirdPlatformTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatAppThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatAppThirdPlatformConfig)

    ticket, err := queryWechatAppThirdPlatformTicket(codeName.(string))
    if nil != err {
        return nil, err
    }

    result, err := httpreq.New(wechatAppThirdPlatformTokenURL).
        RequestBody(Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatAppThirdPlatformToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatAppThirdPlatformTokenResponse)).
    (*WechatAppThirdPlatformTokenResponse)
    if nil == response || 0 == len(response.ComponentAccessToken) {
        return nil, &UnexpectedError{Message:
        "Request WechatAppThirdPlatformToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":       config.AppId,
        "ACCESS_TOKEN": response.ComponentAccessToken,
        "EXPIRES_IN":   StrFromInt(response.ExpiresIn)}, nil
}

var wechatAppThirdPlatformPreAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="

type WechatAppThirdPlatformPreAuthCodeResponse struct {
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatAppThirdPlatformPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatAppThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatAppThirdPlatformToken)

    result, err := httpreq.New(wechatAppThirdPlatformPreAuthCodeURL + tokenItem.AccessToken).
        RequestBody(Json(map[string]string{"component_appid": tokenItem.AppId})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatAppThirdPlatformPreAuthCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatAppThirdPlatformPreAuthCodeResponse)).
    (*WechatAppThirdPlatformPreAuthCodeResponse)
    if nil == response || 0 == len(response.PreAuthCode) {
        return nil, &UnexpectedError{Message:
        "Request WechatAppThirdPlatformPreAuthCode Failed: " + result}
    }
    return map[string]string{
        "APP_ID":        tokenItem.AppId,
        "PRE_AUTH_CODE": response.PreAuthCode,
        "EXPIRES_IN":    StrFromInt(response.ExpiresIn)}, nil
}

var wechatAppThirdPlatformQueryAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token="

type WechatAppThirdPlatformQueryAuthResponse struct {
    AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

type AuthorizationInfo struct {
    AuthorizerAppid        string     `json:"authorizer_appid"`
    AuthorizerAccessToken  string     `json:"authorizer_access_token"`
    ExpiresIn              int        `json:"expires_in"`
    AuthorizerRefreshToken string     `json:"authorizer_refresh_token"`
    FuncInfo               []FuncInfo `json:"func_info"`
}

type FuncInfo struct {
    FuncscopeCategory FuncscopeCategory `json:"funcscope_category"`
}

type FuncscopeCategory struct {
    Id int `json:"id"`
}

func wechatAppThirdPlatformQueryAuthRequestor(codeName, authorizationCode interface{}) (map[string]string, error) {
    cache, err := wechatAppThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatAppThirdPlatformToken)

    result, err := httpreq.New(wechatAppThirdPlatformQueryAuthURL + tokenItem.AccessToken).
        RequestBody(Json(map[string]string{
            "component_appid":    tokenItem.AppId,
            "authorization_code": authorizationCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Request WechatAppThirdPlatformAuthorizerToken Response:(%s, ) %s", codeName, authorizationCode, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatAppThirdPlatformQueryAuthResponse)).
    (*WechatAppThirdPlatformQueryAuthResponse)
    if nil == response || 0 == len(response.AuthorizationInfo.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message:
        "Request WechatAppThirdPlatformAuthorizerToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         response.AuthorizationInfo.AuthorizerAppid,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizationInfo.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizationInfo.AuthorizerRefreshToken,
        "EXPIRES_IN":               StrFromInt(response.AuthorizationInfo.ExpiresIn)}, nil
}

var wechatAppThirdPlatformRefreshAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="

type WechatAppThirdPlatformRefreshAuthResponse struct {
    AuthorizerAccessToken  string `json:"authorizer_access_token"`
    ExpiresIn              int    `json:"expires_in"`
    AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

func wechatAppThirdPlatformRefreshAuthRequestor(codeName, authorizerAppId, authorizerRefreshToken string) (map[string]string, error) {
    cache, err := wechatAppThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatAppThirdPlatformToken)

    result, err := httpreq.New(wechatAppThirdPlatformRefreshAuthURL + tokenItem.AccessToken).
        RequestBody(Json(map[string]string{
            "component_appid":          tokenItem.AppId,
            "authorizer_appid":         authorizerAppId,
            "authorizer_refresh_token": authorizerRefreshToken})).
        Prop("Content-Type", "application/json").Post()
    log.Trace("Refresh WechatAppThirdPlatformAuthorizerToken Response:(%s, %s) %s", codeName, authorizerAppId, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatAppThirdPlatformRefreshAuthResponse)).
        (*WechatAppThirdPlatformRefreshAuthResponse)
    if nil == response || 0 == len(response.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message:
        "Refresh WechatAppThirdPlatformAuthorizerToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         authorizerAppId,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizerRefreshToken,
        "EXPIRES_IN":               StrFromInt(response.ExpiresIn)}, nil
}
