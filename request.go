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

func wechatAPITokenRequestor(appId interface{}) (string, int, error) {
    cache, err := wechatAPITokenConfigCache.Value(appId)
    if nil != err {
        return "", 0, err
    }
    config := cache.Data().(*WechatAPITokenConfig)

    result, err := httpreq.New(wechatAPITokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    if nil != err {
        return "", 0, err
    }

    response := new(WechatAPITokenResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", 0, err
    }
    return response.AccessToken, response.ExpiresIn, nil
}

var wechatThirdPlatformTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

type WechatThirdPlatformTokenResponse struct {
    ComponentAccessToken string `json:"component_access_token"`
    ExpiresIn            int    `json:"expires_in"`
}

func wechatThirdPlatformTokenRequestor(appId interface{}) (string, int, error) {
    cache, err := wechatThirdPlatformConfigCache.Value(appId)
    if nil != err {
        return "", 0, err
    }
    config := cache.Data().(*WechatThirdPlatformConfig)

    ticket, err := QueryWechatThirdPlatformTicket(appId.(string))
    if nil != err {
        return "", 0, err
    }

    result, err := httpreq.New(wechatThirdPlatformTokenURL).
        RequestBody(Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Get()
    if nil != err {
        return "", 0, err
    }

    response := new(WechatThirdPlatformTokenResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", 0, err
    }
    return response.ComponentAccessToken, response.ExpiresIn, nil
}

var wechatThirdPlatformPreAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="

type WechatThirdPlatformPreAuthCodeResponse struct {
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatThirdPlatformPreAuthCodeRequestor(appId interface{}) (string, int, error) {
    cache, err := wechatThirdPlatformTokenCache.Value(appId)
    if nil != err {
        return "", 0, err
    }
    tokenItem := cache.Data().(*WechatThirdPlatformToken)

    result, err := httpreq.New(wechatThirdPlatformPreAuthCodeURL + tokenItem.ComponentAccessToken).
        RequestBody(Json(map[string]string{"component_appid": appId.(string)})).
        Prop("Content-Type", "application/json").Get()
    if nil != err {
        return "", 0, err
    }

    response := new(WechatThirdPlatformPreAuthCodeResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", 0, err
    }
    return response.PreAuthCode, response.ExpiresIn, nil
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

func wechatThirdPlatformQueryAuthRequestor(appId, authorizationCode interface{}) (string, string, int, error) {
    cache, err := wechatThirdPlatformTokenCache.Value(appId)
    if nil != err {
        return "", "", 0, err
    }
    tokenItem := cache.Data().(*WechatThirdPlatformToken)

    result, err := httpreq.New(wechatThirdPlatformQueryAuthURL + tokenItem.ComponentAccessToken).
        RequestBody(Json(map[string]string{
            "component_appid":    appId.(string),
            "authorization_code": authorizationCode.(string)})).
        Prop("Content-Type", "application/json").Get()
    if nil != err {
        return "", "", 0, err
    }

    response := new(WechatThirdPlatformQueryAuthResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", "", 0, err
    }
    return response.AuthorizationInfo.AuthorizerAccessToken,
        response.AuthorizationInfo.AuthorizerRefreshToken,
        response.AuthorizationInfo.ExpiresIn, nil
}

var wechatThirdPlatformRefreshAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="

type WechatThirdPlatformRefreshAuthResponse struct {
    AuthorizerAccessToken  string `json:"authorizer_access_token"`
    ExpiresIn              int    `json:"expires_in"`
    AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

func wechatThirdPlatformRefreshAuthRequestor(appId, authorizerAppId, authorizerRefreshToken string) (string, string, int, error) {
    cache, err := wechatThirdPlatformTokenCache.Value(appId)
    if nil != err {
        return "", "", 0, err
    }
    tokenItem := cache.Data().(*WechatThirdPlatformToken)

    result, err := httpreq.New(wechatThirdPlatformRefreshAuthURL + tokenItem.ComponentAccessToken).
        RequestBody(Json(map[string]string{
            "component_appid":          appId,
            "authorizer_appid":         authorizerAppId,
            "authorizer_refresh_token": authorizerRefreshToken})).
        Prop("Content-Type", "application/json").Get()
    if nil != err {
        return "", "", 0, err
    }

    response := new(WechatThirdPlatformRefreshAuthResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", "", 0, err
    }
    return response.AuthorizerAccessToken,
        response.AuthorizerRefreshToken,
        response.ExpiresIn, nil
}
