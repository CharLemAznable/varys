package varys

import (
    _ "github.com/go-sql-driver/mysql"
    "net/http"
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
    appId := strings.TrimPrefix(request.RequestURI, _path+queryWechatAPITokenPath)
    if 0 == len(appId) {
        writer.Write([]byte(Json(map[string]string{"appId": appId, "error": "AppId is Empty"})))
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
    appId := strings.TrimPrefix(request.RequestURI, _path+acceptComponentVerifyTicketPath)
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
