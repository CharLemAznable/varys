package apptp

import (
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "github.com/CharLemAznable/varys/wechat/jsapi"
    "github.com/kataras/golog"
    "net/http"
    "strings"
)

func init() {
    RegisterHandler(func(mux *http.ServeMux) {
        gokits.HandleFunc(mux, wechatTpAuthorizeScanPath, wechatTpAuthorizeScan)
        gokits.HandleFunc(mux, wechatTpAuthorizeLinkPath, wechatTpAuthorizeLink)
        gokits.HandleFunc(mux, wechatTpAuthorizeRedirectPath, wechatTpAuthorizeRedirect)
        gokits.HandleFunc(mux, cleanWechatTpAuthTokenPath, cleanWechatTpAuthToken)
        gokits.HandleFunc(mux, queryWechatTpAuthTokenPath, queryWechatTpAuthToken)
        gokits.HandleFunc(mux, proxyWechatTpAuthPath, proxyWechatTpAuth, gokits.GzipResponseDisabled)
        gokits.HandleFunc(mux, proxyWechatTpAuthMpLoginPath, proxyWechatTpAuthMpLogin, gokits.GzipResponseDisabled)
        gokits.HandleFunc(mux, queryWechatTpAuthJsConfigPath, queryWechatTpAuthJsConfig)
    })
}

type WechatTpPreAuthCodeResponse struct {
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatTpPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := tokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatTpToken)

    result, err := gokits.NewHttpReq(preAuthCodeURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{"component_appid": tokenItem.AppId})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request Wechat Tp PreAuth Code Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpPreAuthCodeResponse)).(*WechatTpPreAuthCodeResponse)
    if nil == response || "" == response.PreAuthCode {
        return nil, errors.New("Request Wechat Tp PreAuth Code Failed: " + result)
    }
    return map[string]string{
        "AppId":       tokenItem.AppId,
        "PreAuthCode": response.PreAuthCode,
        "ExpiresIn":   gokits.StrFromInt(response.ExpiresIn)}, nil
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
    codeName := TrimPrefixPath(request, wechatTpAuthorizeScanPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    response, err := wechatTpPreAuthCodeRequestor(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    appId := response["AppId"]
    preAuthCode := response["PreAuthCode"]

    redirectQuery := request.URL.RawQuery
    if "" != redirectQuery {
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
    codeName := TrimPrefixPath(request, wechatTpAuthorizeLinkPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    response, err := wechatTpPreAuthCodeRequestor(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    appId := response["AppId"]
    preAuthCode := response["PreAuthCode"]

    redirectQuery := request.URL.RawQuery
    if "" != redirectQuery {
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
    codeName := TrimPrefixPath(request, wechatTpAuthorizeRedirectPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    cache, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Illegal"}))
        return
    }
    config := cache.Data().(*WechatTpConfig)
    redirectUrl := config.RedirectURL
    redirectQuery := request.URL.RawQuery

    if "" != redirectUrl && "" != redirectQuery {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }

    gokits.ResponseHtml(writer, fmt.Sprintf(redirectPageHtmlFormat, redirectUrl))
}

// /clean-wechat-tp-auth-token/{codeName:string}/{authorizerAppId:string}
const cleanWechatTpAuthTokenPath = "/clean-wechat-tp-auth-token/"

func cleanWechatTpAuthToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, cleanWechatTpAuthTokenPath)
    if "" == pathParams {
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
    key := WechatTpAuthKey{CodeName: codeName, AuthorizerAppId: authorizerAppId}
    _, _ = authTokenCache.Delete(key)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{"result": "OK"}))
}

// /query-wechat-tp-auth-token/{codeName:string}/{authorizerAppId:string}
const queryWechatTpAuthTokenPath = "/query-wechat-tp-auth-token/"

func queryWechatTpAuthToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, queryWechatTpAuthTokenPath)
    if "" == pathParams {
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
    key := WechatTpAuthKey{CodeName: codeName, AuthorizerAppId: authorizerAppId}
    cache, err := authTokenCache.Value(key)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpAuthToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}

// /proxy-wechat-tp-auth/{codeName:string}/{authorizerAppId:string}/...
const proxyWechatTpAuthPath = "/proxy-wechat-tp-auth/"

func proxyWechatTpAuth(writer http.ResponseWriter, request *http.Request) {
    codePath := TrimPrefixPath(request, proxyWechatTpAuthPath)
    splits := strings.SplitN(codePath, "/", 3)
    if 3 != len(splits) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/authorizerAppId/proxy-path"}))
        return
    }

    codeName := splits[0]
    authorizerAppId := splits[1]
    key := WechatTpAuthKey{CodeName: codeName, AuthorizerAppId: authorizerAppId}
    cache, err := authTokenCache.Value(key)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpAuthToken).AuthorizerAccessToken

    actualPath := splits[2]
    if "" == actualPath {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "proxy PATH is Empty"}))
        return
    }

    req := request
    if req.URL.RawQuery == "" {
        req.URL.RawQuery = req.URL.RawQuery + "access_token=" + token
    } else {
        req.URL.RawQuery = req.URL.RawQuery + "&" + "access_token=" + token
    }
    req.URL.Path = actualPath
    authProxy.ServeHTTP(writer, req)
}

// /proxy-wechat-tp-auth-mp-login/{codeName:string}/{authorizerAppId:string}?js_code=JSCODE
const proxyWechatTpAuthMpLoginPath = "/proxy-wechat-tp-auth-mp-login/"

func proxyWechatTpAuthMpLogin(writer http.ResponseWriter, request *http.Request) {
    codePath := TrimPrefixPath(request, proxyWechatTpAuthMpLoginPath)
    splits := strings.SplitN(codePath, "/", 3)
    if 3 != len(splits) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/authorizerAppId/proxy-path"}))
        return
    }

    codeName := splits[0]
    authorizerAppId := splits[1]
    key := WechatTpAuthKey{CodeName: codeName, AuthorizerAppId: authorizerAppId}
    cache, err := authTokenCache.Value(key)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpAuthToken)
    appId := token.AuthorizerAppId // 小程序的AppID
    componentAppId := token.AppId
    componentAccessToken := token.AuthorizerAccessToken

    if "" == request.URL.Query().Get("js_code") {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "js_code is Empty"}))
        return
    }

    req := request
    req.URL.RawQuery = req.URL.RawQuery +
        "&appid=" + appId +
        "&component_appid=" + componentAppId +
        "&component_access_token=" + componentAccessToken +
        "&grant_type=authorization_code"
    req.URL.Path = "jscode2session"
    authMpLoginProxy.ServeHTTP(writer, req)
}

// /query-wechat-tp-auth-js-config/{codeName:string}/{authorizerAppId:string}?url=URL
const queryWechatTpAuthJsConfigPath = "/query-wechat-tp-auth-js-config/"

func queryWechatTpAuthJsConfig(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, queryWechatTpAuthJsConfigPath)
    if "" == pathParams {
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
    key := WechatTpAuthKey{CodeName: codeName, AuthorizerAppId: authorizerAppId}
    cache, err := authTokenCache.Value(key)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpAuthToken)
    appId := token.AuthorizerAppId
    jsapiTicket := token.AuthorizerJsapiTicket
    if "" == jsapiTicket {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "jsapi_ticket is Empty"}))
        return
    }

    url := request.URL.Query().Get("url")
    if "" == url {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "url is Empty"}))
        return
    }

    gokits.ResponseJson(writer, gokits.Json(
        jsapi.JsConfigBuilder(appId, jsapiTicket, url)))
}
