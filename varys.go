package varys

import (
    . "github.com/CharLemAznable/gokits"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
    "os"
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
    return NewVarys("", "")
}

func (varys *varys) Run() {
    if nil == varys.server {
        LOG.Error("Initial varys Error")
        os.Exit(-1)
    }
    LOG.Info("varys Server Started...")
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
