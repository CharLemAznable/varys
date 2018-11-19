package varys

import (
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
    "net/url"
    "strings"
)

var _path = "/varys"
var _port = ":4236"

func Default() {
    Run("", "")
}

func Run(path, port string) {
    load()

    If(0 != len(path), func() { _path = path })
    If(0 != len(port), func() { _port = port })

    http.HandleFunc(_path+welcomePath, welcome)
    http.HandleFunc(_path+queryWechatAPITokenPath, queryWechatAPIToken)
    http.HandleFunc(_path+acceptComponentVerifyTicketPath, acceptComponentVerifyTicket)
    http.HandleFunc(_path+authorizeComponentPath, authorizeComponent)
    http.HandleFunc(_path+authorizeRedirectPath, authorizeRedirect)
    http.ListenAndServe(_port, nil)
}

const welcomePath = "/welcome"

func welcome(writer http.ResponseWriter, request *http.Request) {
    writer.Write([]byte(`Three great men, a king, a priest, and a rich man.
Between them stands a common sellsword.
Each great man bids the sellsword kill the other two.
Who lives, who dies?
`))
}

const queryWechatAPITokenPath = "/query-wechat-api-token/"

func queryWechatAPIToken(writer http.ResponseWriter, request *http.Request) {
    appId := strings.TrimPrefix(request.URL.Path, _path+queryWechatAPITokenPath)
    if 0 == len(appId) {
        writer.Write([]byte(Json(map[string]string{"error": "AppId is Empty"})))
        return
    }

    cache, err := wechatAPITokenCache.Value(appId)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{
            "appId": appId, "error": err.Error()})))
        return
    }
    token := cache.Data().(*WechatAPIToken)
    writer.Write([]byte(Json(map[string]string{
        "appId": appId, "token": token.AccessToken})))
}

const acceptComponentVerifyTicketPath = "/accept-verify-ticket/"

func acceptComponentVerifyTicket(writer http.ResponseWriter, request *http.Request) {
    appId := strings.TrimPrefix(request.URL.Path, _path+acceptComponentVerifyTicketPath)
    if 0 != len(appId) {
        authorizeData, err := parseWechatAuthorizeData(appId, request)
        if nil == err {

            if "component_verify_ticket" == authorizeData.InfoType {
                UpdateWechatThirdPlatformTicket(appId, authorizeData.ComponentVerifyTicket)

            } else if "authorized" == authorizeData.InfoType {
                // TODO

            } else if "unauthorized" == authorizeData.InfoType {
                // TODO

            } else if "updateauthorized" == authorizeData.InfoType {
                // TODO

            }
        }
    }
    // 接收到定时推送component_verify_ticket后必须直接返回字符串success
    writer.Write([]byte("success"))
}

const authorizeComponentPath = "/authorize-component/"
const redirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/authorize-component/")) + "/authorize-redirect/%s" + "%s";
    location.replace(
        "https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri)
    );
</script></head></html>
`

func authorizeComponent(writer http.ResponseWriter, request *http.Request) {
    appId := strings.TrimPrefix(request.URL.Path, _path+authorizeComponentPath)
    if 0 == len(appId) {
        writer.Write([]byte(Json(map[string]string{"error": "AppId is Empty"})))
        return
    }

    cache, err := wechatThirdPlatformPreAuthCodeCache.Value(appId)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{"appId": appId, "error": err.Error()})))
        return
    }
    codeItem := cache.Data().(*WechatThirdPlatformPreAuthCode)
    code := codeItem.PreAuthCode

    redirectQuery := request.URL.RawQuery
    if 0 != len(redirectQuery) {
        redirectQuery = "?" + redirectQuery
    }

    writer.Write([]byte(fmt.Sprintf(redirectPageHtmlFormat,
        appId, redirectQuery, appId, code)))
}

const authorizeRedirectPath = "/authorize-redirect/"
const authorizedPageHtml = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style></head><body><div class="tips">授权成功</div></body></html>
`

func authorizeRedirect(writer http.ResponseWriter, request *http.Request) {
    appId := strings.TrimPrefix(request.URL.Path, _path+authorizeComponentPath)
    if 0 == len(appId) {
        writer.Write([]byte(Json(map[string]string{"error": "AppId is Empty"})))
        return
    }

    params, err := url.ParseQuery(request.URL.RawQuery)
    if nil != err {
        writer.Write([]byte(Json(map[string]string{"error": err.Error()})))
        return
    }
    fmt.Println(params["auth_code"])
    fmt.Println(params["expires_in"])

    // TODO

    writer.Write([]byte(authorizedPageHtml))
}
