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
    varysMux.HandleFunc(_path+welcomePath, welcome)

    varysMux.HandleFunc(_path+queryWechatAppTokenPath, queryWechatAppToken)
    varysMux.HandleFunc(_path+proxyWechatAppPath, proxyWechatApp)

    varysMux.HandleFunc(_path+acceptAppAuthorizationPath, acceptAppAuthorization)
    varysMux.HandleFunc(_path+appAuthorizeComponentScanPath, appAuthorizeComponentScan)
    varysMux.HandleFunc(_path+appAuthorizeComponentLinkPath, appAuthorizeComponentLink)
    varysMux.HandleFunc(_path+appAuthorizeRedirectPath, appAuthorizeRedirect)
    varysMux.HandleFunc(_path+queryWechatAppAuthorizerTokenPath, queryWechatAppAuthorizerToken)

    varysMux.HandleFunc(_path+queryWechatCorpTokenPath, queryWechatCorpToken)
    varysMux.HandleFunc(_path+proxyWechatCorpPath, proxyWechatCorp)

    varysMux.HandleFunc(_path+acceptCorpAuthorizationPath, acceptCorpAuthorization)
    varysMux.HandleFunc(_path+corpAuthorizeComponentPath, corpAuthorizeComponent)
    varysMux.HandleFunc(_path+corpAuthorizeRedirectPath, corpAuthorizeRedirect)
    varysMux.HandleFunc(_path+queryWechatCorpAuthorizerTokenPath, queryWechatCorpAuthorizerToken)

    varys := new(varys)
    varys.server = &http.Server{Addr: _port, Handler: varysMux}
    return varys
}

func Default() *varys {
    path, port := "", ""
    configFile, err := ReadYamlFile("varys.yaml")
    if nil == err {
        configMap, err := MapOfYaml(configFile.Root, "root")
        if nil == err {
            path, _ = StringOfYaml(configMap["path"], "path")
            port, _ = StringOfYaml(configMap["port"], "port")
        }
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
    _, _ = writer.Write([]byte(`Three great men, a king, a priest, and a rich man.
Between them stands a common sellsword.
Each great man bids the sellsword kill the other two.
Who lives, who dies?
`))
}
