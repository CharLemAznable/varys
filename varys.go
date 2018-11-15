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
    token, err := GetWechatAPIToken(appId)
    writer.Write([]byte(Json(Condition(nil != err, func() interface{} {
        return map[string]string{"appId": appId, "error": err.Error()}
    }, func() interface{} {
        return map[string]string{"appId": appId, "token": token.AccessToken}
    }))))
}
