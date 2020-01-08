package main

import (
    . "github.com/CharLemAznable/gokits"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
)

func main() {
    mux := http.NewServeMux()
    HandleFunc(mux, welcomePath, welcome, DumpRequestDisabled)

    HandleFunc(mux, queryWechatAppTokenPath, queryWechatAppToken)
    HandleFunc(mux, proxyWechatAppPath, proxyWechatApp)

    HandleFunc(mux, acceptAppAuthorizationPath, acceptAppAuthorization)
    HandleFunc(mux, appAuthorizeComponentScanPath, appAuthorizeComponentScan)
    HandleFunc(mux, appAuthorizeComponentLinkPath, appAuthorizeComponentLink)
    HandleFunc(mux, appAuthorizeRedirectPath, appAuthorizeRedirect)
    HandleFunc(mux, queryWechatAppAuthorizerTokenPath, queryWechatAppAuthorizerToken)

    HandleFunc(mux, queryWechatCorpTokenPath, queryWechatCorpToken)
    HandleFunc(mux, proxyWechatCorpPath, proxyWechatCorp)

    HandleFunc(mux, acceptCorpAuthorizationPath, acceptCorpAuthorization)
    HandleFunc(mux, corpAuthorizeComponentPath, corpAuthorizeComponent)
    HandleFunc(mux, corpAuthorizeRedirectPath, corpAuthorizeRedirect)
    HandleFunc(mux, queryWechatCorpAuthorizerTokenPath, queryWechatCorpAuthorizerToken)

    server := http.Server{Addr: ":" + StrFromInt(appConfig.Port), Handler: mux}
    if err := server.ListenAndServe(); err != nil {
        LOG.Crashf("Start server Error: %s", err.Error())
    }
}
