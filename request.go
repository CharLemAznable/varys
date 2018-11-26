package varys

import (
    "encoding/json"
    "github.com/CharLemAznable/httpreq"
)

var wechatAPITokenURL = "https://api.weixin.qq.com/cgi-bin/token"

type WechatAPITokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatAPITokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatAPITokenConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatAPITokenConfig)

    result, err := httpreq.New(wechatAPITokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    if nil != err {
        return nil, err
    }

    response := new(WechatAPITokenResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err || 0 == len(response.AccessToken) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
            "Request access_token Failed: " + result}).(error)
    }
    return map[string]string{
        "APP_ID":       config.AppId,
        "ACCESS_TOKEN": response.AccessToken,
        "EXPIRES_IN":   StrFromInt(response.ExpiresIn)}, nil
}

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

    ticket, err := QueryWechatThirdPlatformTicket(codeName.(string))
    if nil != err {
        return nil, err
    }

    result, err := httpreq.New(wechatThirdPlatformTokenURL).
        RequestBody(Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Post()
    if nil != err {
        return nil, err
    }

    response := new(WechatThirdPlatformTokenResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err || 0 == len(response.ComponentAccessToken) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
            "Request component_access_token Failed: " + result}).(error)
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
    if nil != err {
        return nil, err
    }

    response := new(WechatThirdPlatformPreAuthCodeResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err || 0 == len(response.PreAuthCode) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
            "Request pre_auth_code Failed: " + result}).(error)
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
    if nil != err {
        return nil, err
    }

    response := new(WechatThirdPlatformQueryAuthResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err || 0 == len(response.AuthorizationInfo.AuthorizerAccessToken) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
            "Request authorizer_access_token Failed: " + result}).(error)
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
    if nil != err {
        return nil, err
    }

    response := new(WechatThirdPlatformRefreshAuthResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err || 0 == len(response.AuthorizerAccessToken) {
        return nil, DefaultIfNil(err, &UnexpectedError{Message:
            "Refresh authorizer_access_token Failed: " + result}).(error)
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         authorizerAppId,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizerRefreshToken,
        "EXPIRES_IN":               StrFromInt(response.ExpiresIn)}, nil
}
