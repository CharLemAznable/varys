package corptp

import (
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "net/http"
    "strings"
)

func init() {
    RegisterHandler(func(mux *http.ServeMux) {
        gokits.HandleFunc(mux, wechatCorpTpAuthComponentPath, wechatCorpTpAuthComponent)
        gokits.HandleFunc(mux, wechatCorpTpAuthRedirectPath, wechatCorpTpAuthRedirect)
        gokits.HandleFunc(mux, cleanWechatCorpTpAuthTokenPath, cleanWechatCorpTpAuthToken)
        gokits.HandleFunc(mux, queryWechatCorpTpAuthTokenPath, queryWechatCorpTpAuthToken)
    })
}

type WechatCorpTpPreAuthCodeResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpTpPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := tokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatCorpTpToken)

    result, err := gokits.NewHttpReq(preAuthCodeURL + tokenItem.AccessToken).Get()
    golog.Debugf("Request Wechat Corp Tp PreAuth Code Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpTpPreAuthCodeResponse)).(*WechatCorpTpPreAuthCodeResponse)
    if nil == response || "" == response.PreAuthCode {
        return nil, errors.New("Request Wechat Corp Tp PreAuth Code Failed: " + result)
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
    codeName := TrimPrefixPath(request, wechatCorpTpAuthComponentPath)
    if "" == codeName {
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
    codeName := TrimPrefixPath(request, wechatCorpTpAuthRedirectPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    redirectQuery := request.URL.RawQuery
    authCode := request.URL.Query().Get("auth_code")
    if "" == authCode {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Corp Unauthorized"}))
        return
    }

    cache, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Illegal"}))
        return
    }
    config := cache.Data().(*WechatCorpTpConfig)
    redirectUrl := config.RedirectURL

    go wechatCorpTpAuthCreator(codeName, authCode)

    if "" != redirectUrl && "" != redirectQuery {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }
    gokits.ResponseHtml(writer, fmt.Sprintf(wechatCorpTpAuthRedirectPageHtmlFormat, redirectUrl))
}

// /clean-wechat-corp-tp-auth-token/{codeName:string}/{corpId:string}
const cleanWechatCorpTpAuthTokenPath = "/clean-wechat-corp-tp-auth-token/"

func cleanWechatCorpTpAuthToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, cleanWechatCorpTpAuthTokenPath)
    if "" == pathParams {
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
    key := WechatCorpTpAuthKey{CodeName: codeName, CorpId: corpId}
    _, _ = permanentCodeCache.Delete(key)
    _, _ = authTokenCache.Delete(key)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{"result": "OK"}))
}

// /query-wechat-corp-tp-auth-token/{codeName:string}/{corpId:string}
const queryWechatCorpTpAuthTokenPath = "/query-wechat-corp-tp-auth-token/"

func queryWechatCorpTpAuthToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, queryWechatCorpTpAuthTokenPath)
    if "" == pathParams {
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
    cache, err := authTokenCache.Value(
        WechatCorpTpAuthKey{CodeName: codeName, CorpId: corpId})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatCorpTpAuthToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}
