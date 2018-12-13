package varys

import (
    "fmt"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
    "os"
    "strings"
)

type varys struct {
    server *http.Server
}

var _path = "/varys"
var _port = ":4236"

func NewVarys(path, port string) *varys {
    load()

    If(0 != len(path), func() { _path = path })
    If(0 != len(port), func() { _port = port })

    varysMux := http.NewServeMux()
    varysMux.Handle("/", http.FileServer(http.Dir("varys"))) // static resources
    varysMux.HandleFunc(_path+welcomePath, welcome)
    varysMux.HandleFunc(_path+queryWechatAppTokenPath, queryWechatAppToken)
    varysMux.HandleFunc(_path+queryWechatAuthorizerTokenPath, queryWechatAuthorizerToken)
    varysMux.HandleFunc(_path+queryWechatCorpTokenPath, queryWechatCorpToken)
    varysMux.HandleFunc(_path+queryWechatCorpAuthorizerTokenPath, queryWechatCorpAuthorizerToken)
    varysMux.HandleFunc(_path+acceptAuthorizationPath, acceptAuthorization)
    varysMux.HandleFunc(_path+acceptCorpAuthorizationPath, acceptCorpAuthorization)
    varysMux.HandleFunc(_path+authorizeComponentScanPath, authorizeComponentScan)
    varysMux.HandleFunc(_path+authorizeComponentLinkPath, authorizeComponentLink)
    varysMux.HandleFunc(_path+authorizeRedirectPath, authorizeRedirect)
    varysMux.HandleFunc(_path+authorizeCorpComponentPath, authorizeCorpComponent)
    varysMux.HandleFunc(_path+authorizeCorpRedirectPath, authorizeCorpRedirect)
    varysServer := &http.Server{Addr: _port, Handler: varysMux}

    varys := new(varys)
    varys.server = varysServer
    return varys
}

func Default() *varys {
    return NewVarys("", "")
}

func (varys *varys) Run() {
    if nil == varys.server {
        log.Error("Initial varys Error")
        os.Exit(-1)
    }
    log.Info("varys Server Started...")
    varys.server.ListenAndServe()
}

const welcomePath = "/welcome"

func welcome(writer http.ResponseWriter, request *http.Request) {
    writer.Write([]byte(`Three great men, a king, a priest, and a rich man.
Between them stands a common sellsword.
Each great man bids the sellsword kill the other two.
Who lives, who dies?
`))
}

// /query-wechat-app-token/{codeName:string}
const queryWechatAppTokenPath = "/query-wechat-app-token/"

func queryWechatAppToken(writer http.ResponseWriter, request *http.Request) {
    writer.Header().Set("Content-Type", "application/json; charset=utf-8")

    codeName := strings.TrimPrefix(request.URL.Path, _path+queryWechatAppTokenPath)
    if 0 == len(codeName) {
        writer.Write([]byte(Json(map[string]string{
            "error": "codeName is Empty"})))
        return
    }

    cache, err := wechatAppTokenCache.Value(codeName)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{
            "error": err.Error()})))
        return
    }
    token := cache.Data().(*WechatAppToken)
    writer.Write([]byte(Json(map[string]string{
        "appId": token.AppId, "token": token.AccessToken})))
}

// /query-wechat-authorizer-token/{codeName:string}/{authorizerAppId:string}
const queryWechatAuthorizerTokenPath = "/query-wechat-authorizer-token/"

func queryWechatAuthorizerToken(writer http.ResponseWriter, request *http.Request) {
    writer.Header().Set("Content-Type", "application/json; charset=utf-8")

    pathParams := strings.TrimPrefix(request.URL.Path, _path+queryWechatAuthorizerTokenPath)
    if 0 == len(pathParams) {
        writer.Write([]byte(Json(map[string]string{
            "error": "Path Params is Empty"})))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        writer.Write([]byte(Json(map[string]string{
            "error": "Missing param codeName/authorizerAppId"})))
        return
    }

    codeName := ids[0]
    authorizerAppId := ids[1]
    cache, err := wechatThirdPlatformAuthorizerTokenCache.Value(
        WechatThirdPlatformAuthorizerTokenKey{CodeName: codeName, AuthorizerAppId: authorizerAppId})
    if nil != err {
        writer.Write([]byte(Json(map[string]string{
            "error": err.Error()})))
        return
    }
    token := cache.Data().(*WechatThirdPlatformAuthorizerToken)
    writer.Write([]byte(Json(map[string]string{
        "appId": token.AppId, "authorizerAppId": token.AuthorizerAppId,
        "token": token.AuthorizerAccessToken})))
}

// /query-wechat-corp-token/{codeName:string}
const queryWechatCorpTokenPath = "/query-wechat-corp-token/"

func queryWechatCorpToken(writer http.ResponseWriter, request *http.Request) {
    writer.Header().Set("Content-Type", "application/json; charset=utf-8")

    codeName := strings.TrimPrefix(request.URL.Path, _path+queryWechatCorpTokenPath)
    if 0 == len(codeName) {
        writer.Write([]byte(Json(map[string]string{
            "error": "codeName is Empty"})))
        return
    }

    cache, err := wechatCorpTokenCache.Value(codeName)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{
            "error": err.Error()})))
        return
    }
    token := cache.Data().(*WechatCorpToken)
    writer.Write([]byte(Json(map[string]string{
        "corpId": token.CorpId, "token": token.AccessToken})))
}

// /query-wechat-corp-authorizer-token/{codeName:string}/{corpId:string}
const queryWechatCorpAuthorizerTokenPath = "/query-wechat-corp-authorizer-token/"

func queryWechatCorpAuthorizerToken(writer http.ResponseWriter, request *http.Request) {
    writer.Header().Set("Content-Type", "application/json; charset=utf-8")

    pathParams := strings.TrimPrefix(request.URL.Path, _path+queryWechatCorpAuthorizerTokenPath)
    if 0 == len(pathParams) {
        writer.Write([]byte(Json(map[string]string{
            "error": "Path Params is Empty"})))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        writer.Write([]byte(Json(map[string]string{
            "error": "Missing param codeName/corpId"})))
        return
    }

    codeName := ids[0]
    corpId := ids[1]
    cache, err := wechatCorpThirdPlatformCorpTokenCache.Value(
        WechatCorpThirdPlatformAuthorizerKey{CodeName: codeName, CorpId: corpId})
    if nil != err {
        writer.Write([]byte(Json(map[string]string{
            "error": err.Error()})))
        return
    }
    token := cache.Data().(*WechatCorpThirdPlatformCorpToken)
    writer.Write([]byte(Json(map[string]string{
        "suiteId": token.SuiteId, "corpId": token.CorpId,
        "token": token.CorpAccessToken})))
}

// /accept-authorization/{codeName:string}
const acceptAuthorizationPath = "/accept-authorization/"

func acceptAuthorization(writer http.ResponseWriter, request *http.Request) {
    codeName := strings.TrimPrefix(request.URL.Path, _path+acceptAuthorizationPath)
    if 0 != len(codeName) {
        authorizeData, err := parseWechatAuthorizeData(codeName, request)
        if nil == err {
            switch authorizeData.InfoType {

            case "component_verify_ticket":
                updateWechatThirdPlatformTicket(codeName, authorizeData.ComponentVerifyTicket)

            case "authorized":
                enableWechatThirdPlatformAuthorizer(codeName, authorizeData.AuthorizerAppid,
                    authorizeData.AuthorizationCode, authorizeData.PreAuthCode)
                go wechatThirdPlatformAuthorizerTokenCreator(codeName,
                    authorizeData.AuthorizerAppid, authorizeData.AuthorizationCode)

            case "updateauthorized":
                enableWechatThirdPlatformAuthorizer(codeName, authorizeData.AuthorizerAppid,
                    authorizeData.AuthorizationCode, authorizeData.PreAuthCode)
                go wechatThirdPlatformAuthorizerTokenCreator(codeName,
                    authorizeData.AuthorizerAppid, authorizeData.AuthorizationCode)

            case "unauthorized":
                disableWechatThirdPlatformAuthorizer(codeName, authorizeData.AuthorizerAppid)
                // delete cache
                wechatThirdPlatformAuthorizerTokenCache.Delete(
                    WechatThirdPlatformAuthorizerTokenKey{
                        CodeName: codeName, AuthorizerAppId: authorizeData.AuthorizerAppid})

            }
        }
    }

    // 接收到定时推送component_verify_ticket后必须直接返回字符串success
    writer.Write([]byte("success"))
}

// /accept-authorization/{codeName:string}
const acceptCorpAuthorizationPath = "/accept-corp-authorization/"

func acceptCorpAuthorization(writer http.ResponseWriter, request *http.Request) {
    codeName := strings.TrimPrefix(request.URL.Path, _path+acceptCorpAuthorizationPath)
    if 0 != len(codeName) {
        authorizeData, err := parseWechatCorpAuthorizeData(codeName, request)
        if nil == err {
            switch authorizeData.InfoType {

            case "echostr":
                writer.Write([]byte(authorizeData.EchoStr))
                return

            case "suite_ticket":
                updateWechatCorpThirdPlatformTicket(codeName, authorizeData.SuiteTicket)

            case "create_auth":
                go wechatCorpThirdPlatformAuthorizeCreator(codeName, authorizeData.AuthCode)

            case "change_auth":
                // ignore

            case "cancel_auth":
                authCorpId := authorizeData.AuthCorpId
                disableWechatCorpThirdPlatformAuthorizer(codeName, authCorpId)
                // delete cache
                key := WechatCorpThirdPlatformAuthorizerKey{CodeName: codeName, CorpId: authCorpId}
                wechatCorpThirdPlatformPermanentCodeCache.Delete(key)
                wechatCorpThirdPlatformCorpTokenCache.Delete(key)

            }
        }
    }

    writer.Write([]byte("success"))
}

// /authorize-component-scan/{codeName:string}
const authorizeComponentScanPath = "/authorize-component-scan/"
const scanRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/authorize-component-scan/")) + "/authorize-redirect/%s" + "%s";
    location.replace(
        "https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri)
    );
</script></head></html>
`

func authorizeComponentScan(writer http.ResponseWriter, request *http.Request) {
    codeName := strings.TrimPrefix(request.URL.Path, _path+authorizeComponentScanPath)
    if 0 == len(codeName) {
        writer.Write([]byte(Json(map[string]string{"error": "CodeName is Empty"})))
        return
    }

    response, err := wechatThirdPlatformPreAuthCodeRequestor(codeName)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{"error": err.Error()})))
        return
    }
    appId := response["APP_ID"]
    preAuthCode := response["PRE_AUTH_CODE"]

    redirectQuery := request.URL.RawQuery
    if 0 != len(redirectQuery) {
        redirectQuery = "?" + redirectQuery
    }

    writer.Write([]byte(fmt.Sprintf(scanRedirectPageHtmlFormat,
        codeName, redirectQuery, appId, preAuthCode)))
}

// /authorize-component-link/{codeName:string}
const authorizeComponentLinkPath = "/authorize-component-link/"
const linkRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/authorize-component-link/")) + "/authorize-redirect/%s" + "%s";
    location.replace(
        "https://mp.weixin.qq.com/safe/bindcomponent?action=bindcomponent&no_scan=1&component_appid=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri) + "#wechat_redirect"
    );
</script></head></html>
`

func authorizeComponentLink(writer http.ResponseWriter, request *http.Request) {
    codeName := strings.TrimPrefix(request.URL.Path, _path+authorizeComponentLinkPath)
    if 0 == len(codeName) {
        writer.Write([]byte(Json(map[string]string{"error": "CodeName is Empty"})))
        return
    }

    response, err := wechatThirdPlatformPreAuthCodeRequestor(codeName)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{"error": err.Error()})))
        return
    }
    appId := response["APP_ID"]
    preAuthCode := response["PRE_AUTH_CODE"]

    redirectQuery := request.URL.RawQuery
    if 0 != len(redirectQuery) {
        redirectQuery = "?" + redirectQuery
    }

    writer.Write([]byte(fmt.Sprintf(linkRedirectPageHtmlFormat,
        codeName, redirectQuery, appId, preAuthCode)))
}

// /authorize-redirect/{codeName:string}
const authorizeRedirectPath = "/authorize-redirect/"
const authorizedPageHtmlFormat = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style><script>var p="%s";0!=p.length&&location.replace(p);</script></head><body><div class="tips">授权成功</div></body></html>
`

func authorizeRedirect(writer http.ResponseWriter, request *http.Request) {
    codeName := strings.TrimPrefix(request.URL.Path, _path+authorizeRedirectPath)
    if 0 == len(codeName) {
        writer.Write([]byte(Json(map[string]string{"error": "CodeName is Empty"})))
        return
    }

    cache, err := wechatThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{"error": "CodeName is Illegal"})))
        return
    }
    config := cache.Data().(*WechatThirdPlatformConfig)
    redirectUrl := config.RedirectURL
    redirectQuery := request.URL.RawQuery

    if 0 != len(redirectUrl) && 0 != len(redirectQuery) {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }

    writer.Write([]byte(fmt.Sprintf(authorizedPageHtmlFormat, redirectUrl)))
}

// /authorize-corp-component/{codeName:string}
const authorizeCorpComponentPath = "/authorize-corp-component/"
const corpRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/authorize-corp-component/")) + "/authorize-corp-redirect/%s";
    location.replace(
        "https://open.work.weixin.qq.com/3rdapp/install?suite_id=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri) + "&state=%s"
    );
</script></head></html>
`

func authorizeCorpComponent(writer http.ResponseWriter, request *http.Request) {
    codeName := strings.TrimPrefix(request.URL.Path, _path+authorizeCorpComponentPath)
    if 0 == len(codeName) {
        writer.Write([]byte(Json(map[string]string{"error": "CodeName is Empty"})))
        return
    }

    response, err := wechatCorpThirdPlatformPreAuthCodeRequestor(codeName)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{"error": err.Error()})))
        return
    }
    suiteId := response["SUITE_ID"]
    preAuthCode := response["PRE_AUTH_CODE"]
    state := request.URL.Query().Get("state")

    writer.Write([]byte(fmt.Sprintf(corpRedirectPageHtmlFormat,
        codeName, suiteId, preAuthCode, state)))
}

// /authorize-corp-redirect/{codeName:string}
const authorizeCorpRedirectPath = "/authorize-corp-redirect/"
const authorizedCorpPageHtmlFormat = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style><script>var p="%s";0!=p.length&&location.replace(p);</script></head><body><div class="tips">授权成功</div></body></html>
`

func authorizeCorpRedirect(writer http.ResponseWriter, request *http.Request) {
    codeName := strings.TrimPrefix(request.URL.Path, _path+authorizeCorpRedirectPath)
    if 0 == len(codeName) {
        writer.Write([]byte(Json(map[string]string{"error": "CodeName is Empty"})))
        return
    }

    redirectQuery := request.URL.RawQuery
    authCode := request.URL.Query().Get("auth_code")
    if 0 == len(authCode) {
        writer.Write([]byte(Json(map[string]string{"error": "Corp Unauthorized"})))
        return
    }

    cache, err := wechatCorpThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{"error": "CodeName is Illegal"})))
        return
    }
    config := cache.Data().(*WechatCorpThirdPlatformConfig)
    redirectUrl := config.RedirectURL

    go wechatCorpThirdPlatformAuthorizeCreator(codeName, authCode)

    if 0 != len(redirectUrl) && 0 != len(redirectQuery) {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }

    writer.Write([]byte(fmt.Sprintf(authorizedCorpPageHtmlFormat, redirectUrl)))
}
