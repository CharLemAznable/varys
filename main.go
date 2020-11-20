package main

import (
    . "github.com/CharLemAznable/gokits"
    _ "github.com/go-sql-driver/mysql"
    "github.com/kataras/golog"
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
    HandleFunc(mux, queryWechatAppJsConfigPath, queryWechatAppJsConfig)

    HandleFunc(mux, acceptWechatTpInfoPath, acceptWechatTpInfo)
    HandleFunc(mux, queryWechatTpTokenPath, queryWechatTpToken)
    HandleFunc(mux, proxyWechatTpPath, proxyWechatTp, GzipResponseDisabled)

    HandleFunc(mux, wechatTpAuthorizeScanPath, wechatTpAuthorizeScan)
    HandleFunc(mux, wechatTpAuthorizeLinkPath, wechatTpAuthorizeLink)
    HandleFunc(mux, wechatTpAuthorizeRedirectPath, wechatTpAuthorizeRedirect)
    HandleFunc(mux, cleanWechatTpAuthTokenPath, cleanWechatTpAuthToken)
    HandleFunc(mux, queryWechatTpAuthTokenPath, queryWechatTpAuthToken)
    HandleFunc(mux, proxyWechatTpAuthPath, proxyWechatTpAuth, GzipResponseDisabled)
    HandleFunc(mux, queryWechatTpAuthJsConfigPath, queryWechatTpAuthJsConfig)

    HandleFunc(mux, queryWechatCorpTokenPath, queryWechatCorpToken)
    HandleFunc(mux, proxyWechatCorpPath, proxyWechatCorp, GzipResponseDisabled)

    HandleFunc(mux, acceptWechatCorpTpInfoPath, acceptWechatCorpTpInfo)

    HandleFunc(mux, wechatCorpTpAuthComponentPath, wechatCorpTpAuthComponent)
    HandleFunc(mux, wechatCorpTpAuthRedirectPath, wechatCorpTpAuthRedirect)
    HandleFunc(mux, cleanWechatCorpTpAuthTokenPath, cleanWechatCorpTpAuthToken)
    HandleFunc(mux, queryWechatCorpTpAuthTokenPath, queryWechatCorpTpAuthToken)

    HandleFunc(mux, queryToutiaoAppTokenPath, queryToutiaoAppToken)

    HandleFunc(mux, queryFengniaoAppTokenPath, queryFengniaoAppToken)
    HandleFunc(mux, proxyFengniaoAppPath, proxyFengniaoApp, GzipResponseDisabled)
    HandleFunc(mux, callbackFengniaoOrderPath, callbackFengniaoOrder)

    server := http.Server{Addr: ":" + StrFromInt(globalConfig.Port), Handler: mux}
    if err := server.ListenAndServe(); err != nil {
        golog.Errorf("Start server Error: %s", err.Error())
    }
}
