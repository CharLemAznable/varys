package main

import (
    "fmt"
    "github.com/CharLemAznable/gokits"
    "net/http"
    "strings"
)

type WechatTpPreAuthCodeResponse struct {
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatTpPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatTpTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatTpToken)

    result, err := gokits.NewHttpReq(wechatTpPreAuthCodeURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{"component_appid": tokenItem.AppId})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatTpPreAuthCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpPreAuthCodeResponse)).(*WechatTpPreAuthCodeResponse)
    if nil == response || 0 == len(response.PreAuthCode) {
        return nil, &UnexpectedError{Message: "Request WechatTpPreAuthCode Failed: " + result}
    }
    return map[string]string{
        "APP_ID":        tokenItem.AppId,
        "PRE_AUTH_CODE": response.PreAuthCode,
        "EXPIRES_IN":    gokits.StrFromInt(response.ExpiresIn)}, nil
}

// /wechat-tp-authorize-scan/{codeName:string}
const wechatTpAuthorizeScanPath = "/wechat-tp-authorize-scan/"
const scanRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/wechat-tp-authorize-scan/")) + "/wechat-tp-authorize-redirect/%s" + "%s";
    location.replace(
        "https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri)
    );
</script></head></html>
`

func wechatTpAuthorizeScan(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, wechatTpAuthorizeScanPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    response, err := wechatTpPreAuthCodeRequestor(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    appId := response["APP_ID"]
    preAuthCode := response["PRE_AUTH_CODE"]

    redirectQuery := request.URL.RawQuery
    if 0 != len(redirectQuery) {
        redirectQuery = "?" + redirectQuery
    }

    gokits.ResponseHtml(writer, fmt.Sprintf(scanRedirectPageHtmlFormat, codeName, redirectQuery, appId, preAuthCode))
}

// /wechat-tp-authorize-link/{codeName:string}
const wechatTpAuthorizeLinkPath = "/wechat-tp-authorize-link/"
const linkRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/wechat-tp-authorize-link/")) + "/wechat-tp-authorize-redirect/%s" + "%s";
    location.replace(
        "https://mp.weixin.qq.com/safe/bindcomponent?action=bindcomponent&no_scan=1&component_appid=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri) + "#wechat_redirect"
    );
</script></head></html>
`

func wechatTpAuthorizeLink(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, wechatTpAuthorizeLinkPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    response, err := wechatTpPreAuthCodeRequestor(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    appId := response["APP_ID"]
    preAuthCode := response["PRE_AUTH_CODE"]

    redirectQuery := request.URL.RawQuery
    if 0 != len(redirectQuery) {
        redirectQuery = "?" + redirectQuery
    }

    gokits.ResponseHtml(writer, fmt.Sprintf(linkRedirectPageHtmlFormat, codeName, redirectQuery, appId, preAuthCode))
}

// /wechat-tp-authorize-redirect/{codeName:string}
const wechatTpAuthorizeRedirectPath = "/wechat-tp-authorize-redirect/"
const redirectPageHtmlFormat = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style><script>var p="%s";0!=p.length&&location.replace(p);</script></head><body><div class="tips">授权成功</div></body></html>
`

func wechatTpAuthorizeRedirect(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, wechatTpAuthorizeRedirectPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    cache, err := wechatTpConfigCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Illegal"}))
        return
    }
    config := cache.Data().(*WechatTpConfig)
    redirectUrl := config.RedirectURL
    redirectQuery := request.URL.RawQuery

    if 0 != len(redirectUrl) && 0 != len(redirectQuery) {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }

    gokits.ResponseHtml(writer, fmt.Sprintf(redirectPageHtmlFormat, redirectUrl))
}

// /query-wechat-tp-auth-token/{codeName:string}/{authorizerAppId:string}
const queryWechatTpAuthTokenPath = "/query-wechat-tp-auth-token/"

func queryWechatTpAuthToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := trimPrefixPath(request, queryWechatTpAuthTokenPath)
    if 0 == len(pathParams) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/authorizerAppId"}))
        return
    }

    codeName := ids[0]
    authorizerAppId := ids[1]
    cache, err := wechatTpAuthTokenCache.Value(
        WechatTpAuthKey{CodeName: codeName, AuthorizerAppId: authorizerAppId})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpAuthToken)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{
        "appId":           token.AppId,
        "authorizerAppId": token.AuthorizerAppId,
        "token":           token.AuthorizerAccessToken}))
}
