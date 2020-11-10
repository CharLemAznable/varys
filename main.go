package main

import (
    . "github.com/CharLemAznable/gokits"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
)

func main() {
    mux := http.NewServeMux()
    HandleFunc(mux, "/", EmptyHandler, DumpRequestDisabled)
    HandleFunc(mux, welcomePath, welcome, DumpRequestDisabled)

    HandleFunc(mux, queryWechatAppTokenPath, queryWechatAppToken)
    HandleFunc(mux, proxyWechatAppPath, proxyWechatApp, GzipResponseDisabled)
    HandleFunc(mux, proxyWechatMpPath, proxyWechatMp, GzipResponseDisabled)
    HandleFunc(mux, proxyWechatMpLoginPath, proxyWechatMpLogin, GzipResponseDisabled)

    HandleFunc(mux, acceptWechatTpInfoPath, acceptWechatTpInfo)
    HandleFunc(mux, queryWechatTpTokenPath, queryWechatTpToken)
    HandleFunc(mux, proxyWechatTpPath, proxyWechatTp, GzipResponseDisabled)

    HandleFunc(mux, wechatTpAuthorizeScanPath, wechatTpAuthorizeScan)
    HandleFunc(mux, wechatTpAuthorizeLinkPath, wechatTpAuthorizeLink)
    HandleFunc(mux, wechatTpAuthorizeRedirectPath, wechatTpAuthorizeRedirect)
    HandleFunc(mux, queryWechatTpAuthTokenPath, queryWechatTpAuthToken)

    HandleFunc(mux, queryWechatCorpTokenPath, queryWechatCorpToken)
    HandleFunc(mux, proxyWechatCorpPath, proxyWechatCorp, GzipResponseDisabled)

    HandleFunc(mux, acceptCorpAuthorizationPath, acceptCorpAuthorization)
    HandleFunc(mux, corpAuthorizeComponentPath, corpAuthorizeComponent)
    HandleFunc(mux, corpAuthorizeRedirectPath, corpAuthorizeRedirect)
    HandleFunc(mux, queryWechatCorpAuthorizerTokenPath, queryWechatCorpAuthorizerToken)

    HandleFunc(mux, queryToutiaoAppTokenPath, queryToutiaoAppToken)

    server := http.Server{Addr: ":" + StrFromInt(appConfig.Port), Handler: mux}
    if err := server.ListenAndServe(); err != nil {
        LOG.Crashf("Start server Error: %s", err.Error())
    }
}
