package varys

import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/kataras/iris"
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

    app := iris.Default()
    party := app.Party(_path)
    {
        party.Get(welcomePath, welcome)
        party.Get(queryWechatAPITokenPath, queryWechatAPIToken)
    }
    app.Run(iris.Addr(_port))
}

const welcomePath = "/welcome"

func welcome(context iris.Context) {
    context.Text(`Three great men, a king, a priest, and a rich man.
Between them stands a common sellsword.
Each great man bids the sellsword kill the other two.
Who lives, who dies?
`)
}

const queryWechatAPITokenPath = "/query-wechat-api-token/{appId:string}"

func queryWechatAPIToken(context iris.Context) {
    appId := context.Params().Get("appId")
    token, err := GetWechatAPIToken(appId)
    context.JSON(ConditionFunc(nil != err, func() interface{} {
        return map[string]string{"appId": appId, "error": err.Error()}
    }, func() interface{} {
        return map[string]string{"appId": appId, "token": token.AccessToken}
    }))
}
