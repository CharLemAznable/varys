package main

import (
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "net/http"
    "strings"
)

type WechatCorpTpPreAuthCodeResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpTpPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatCorpTpTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatCorpTpToken)

    result, err := gokits.NewHttpReq(wechatCorpTpPreAuthCodeURL + tokenItem.AccessToken).Get()
    golog.Debugf("Request WechatCorpTpPreAuthCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpTpPreAuthCodeResponse)).(*WechatCorpTpPreAuthCodeResponse)
    if nil == response || 0 == len(response.PreAuthCode) {
        return nil, errors.New("Request WechatCorpTpPreAuthCode Failed: " + result)
    }
    return map[string]string{
        "SuiteId":     tokenItem.SuiteId,
        "PreAuthCode": response.PreAuthCode,
        "ExpiresIn":   gokits.StrFromInt(response.ExpiresIn)}, nil
}

// /wechat-corp-tp-authorize-component/{codeName:string}
const wechatCorpTpAuthComponentPath = "/wechat-corp-tp-authorize-component/"
const wechatCorpTpAuthComponentPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/wechat-corp-tp-authorize-component/")) + "/wechat-corp-tp-authorize-redirect/%s";
    location.replace(
        "https://open.work.weixin.qq.com/3rdapp/install?suite_id=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri) + "&state=%s"
    );
</script></head></html>
`

func wechatCorpTpAuthComponent(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, wechatCorpTpAuthComponentPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    response, err := wechatCorpTpPreAuthCodeRequestor(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    suiteId := response["SuiteId"]
    preAuthCode := response["PreAuthCode"]
    state := request.URL.Query().Get("state")

    gokits.ResponseHtml(writer, fmt.Sprintf(
        wechatCorpTpAuthComponentPageHtmlFormat, codeName, suiteId, preAuthCode, state))
}

// /wechat-corp-tp-authorize-redirect/{codeName:string}
const wechatCorpTpAuthRedirectPath = "/wechat-corp-tp-authorize-redirect/"
const wechatCorpTpAuthRedirectPageHtmlFormat = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style><script>var p="%s";0!=p.length&&location.replace(p);</script></head><body><div class="tips">授权成功</div></body></html>
`

func wechatCorpTpAuthRedirect(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, wechatCorpTpAuthRedirectPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    redirectQuery := request.URL.RawQuery
    authCode := request.URL.Query().Get("auth_code")
    if 0 == len(authCode) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Corp Unauthorized"}))
        return
    }

    cache, err := wechatCorpTpConfigCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Illegal"}))
        return
    }
    config := cache.Data().(*WechatCorpTpConfig)
    redirectUrl := config.RedirectURL

    go wechatCorpTpAuthCreator(codeName, authCode)

    if 0 != len(redirectUrl) && 0 != len(redirectQuery) {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }

    gokits.ResponseHtml(writer, fmt.Sprintf(wechatCorpTpAuthRedirectPageHtmlFormat, redirectUrl))
}

// /query-wechat-corp-tp-auth-token/{codeName:string}/{corpId:string}
const queryWechatCorpTpAuthTokenPath = "/query-wechat-corp-tp-auth-token/"

func queryWechatCorpTpAuthToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := trimPrefixPath(request, queryWechatCorpTpAuthTokenPath)
    if 0 == len(pathParams) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/corpId"}))
        return
    }

    codeName := ids[0]
    corpId := ids[1]
    cache, err := wechatCorpTpAuthTokenCache.Value(
        WechatCorpTpAuthKey{CodeName: codeName, CorpId: corpId})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatCorpTpAuthToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}
