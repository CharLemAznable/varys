package varys

import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/kataras/iris"
)

var path = "/varys"
var port = ":4236"

func Run() {
    app := iris.Default()
    party := app.Party(path)
    {
        party.Get(welcomePath, welcome)
        party.Get(queryWechatAPITokenPath, queryWechatAPIToken)
    }
    app.Run(iris.Addr(port))
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
