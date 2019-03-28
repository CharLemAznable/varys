package varys

import (
    . "github.com/CharLemAznable/gokits"
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

    If(0 != len(path), func() {
        Condition(strings.HasPrefix(path, "/"),
            func() { _path = path },
            func() { _path = "/" + path })
        If(strings.HasSuffix(_path, "/"),
            func() { _path = _path[:len(_path)-1] })
    })
    If(0 != len(port), func() {
        Condition(strings.HasPrefix(port, ":"),
            func() { _port = port },
            func() { _port = ":" + port })
    })

    varysMux := http.NewServeMux()
    varysMux.Handle("/", http.FileServer(http.Dir("varys"))) // static resources
    varysMux.HandleFunc(JoinPathComponent(_path, welcomePath), welcome)

    varysMux.HandleFunc(JoinPathComponent(_path, queryWechatAppTokenPath), queryWechatAppToken)
    varysMux.HandleFunc(JoinPathComponent(_path, proxyWechatAppPath), proxyWechatApp)

    varysMux.HandleFunc(JoinPathComponent(_path, acceptAppAuthorizationPath), acceptAppAuthorization)
    varysMux.HandleFunc(JoinPathComponent(_path, appAuthorizeComponentScanPath), appAuthorizeComponentScan)
    varysMux.HandleFunc(JoinPathComponent(_path, appAuthorizeComponentLinkPath), appAuthorizeComponentLink)
    varysMux.HandleFunc(JoinPathComponent(_path, appAuthorizeRedirectPath), appAuthorizeRedirect)
    varysMux.HandleFunc(JoinPathComponent(_path, queryWechatAppAuthorizerTokenPath), queryWechatAppAuthorizerToken)

    varysMux.HandleFunc(JoinPathComponent(_path, queryWechatCorpTokenPath), queryWechatCorpToken)
    varysMux.HandleFunc(JoinPathComponent(_path, proxyWechatCorpPath), proxyWechatCorp)

    varysMux.HandleFunc(JoinPathComponent(_path, acceptCorpAuthorizationPath), acceptCorpAuthorization)
    varysMux.HandleFunc(JoinPathComponent(_path, corpAuthorizeComponentPath), corpAuthorizeComponent)
    varysMux.HandleFunc(JoinPathComponent(_path, corpAuthorizeRedirectPath), corpAuthorizeRedirect)
    varysMux.HandleFunc(JoinPathComponent(_path, queryWechatCorpAuthorizerTokenPath), queryWechatCorpAuthorizerToken)

    varys := new(varys)
    varys.server = &http.Server{Addr: _port, Handler: varysMux}
    return varys
}

func Default() *varys {
    path, port := "", ""
    configFile, err := ReadYamlFile("varys.yaml")
    if nil == err {
        path, _ = configFile.GetString("path")
        port, _ = configFile.GetString("port")
    }
    return NewVarys(path, port)
}

func (varys *varys) Run() {
    if nil == varys.server {
        _ = LOG.Error("Initial varys Error")
        os.Exit(-1)
    }
    LOG.Info("varys Server Started ...")
    err := varys.server.ListenAndServe()
    if nil != err {
        _ = LOG.Error("Start varys Error: %s", err.Error())
        os.Exit(-1)
    }
}

const welcomePath = "/welcome"

func welcome(writer http.ResponseWriter, request *http.Request) {
    ResponseText(writer, `Three great men, a king, a priest, and a rich man.
Between them stands a common sellsword.
Each great man bids the sellsword kill the other two.
Who lives, who dies?
`)
}
