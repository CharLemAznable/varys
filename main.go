package main

import (
	"github.com/CharLemAznable/gokits"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	handleFunc(mux, welcomePath, welcome, false)

	handleFunc(mux, queryWechatAppTokenPath, queryWechatAppToken, true)
	handleFunc(mux, proxyWechatAppPath, proxyWechatApp, true)

	handleFunc(mux, acceptAppAuthorizationPath, acceptAppAuthorization, true)
	handleFunc(mux, appAuthorizeComponentScanPath, appAuthorizeComponentScan, true)
	handleFunc(mux, appAuthorizeComponentLinkPath, appAuthorizeComponentLink, true)
	handleFunc(mux, appAuthorizeRedirectPath, appAuthorizeRedirect, true)
	handleFunc(mux, queryWechatAppAuthorizerTokenPath, queryWechatAppAuthorizerToken, true)

	handleFunc(mux, queryWechatCorpTokenPath, queryWechatCorpToken, true)
	handleFunc(mux, proxyWechatCorpPath, proxyWechatCorp, true)

	handleFunc(mux, acceptCorpAuthorizationPath, acceptCorpAuthorization, true)
	handleFunc(mux, corpAuthorizeComponentPath, corpAuthorizeComponent, true)
	handleFunc(mux, corpAuthorizeRedirectPath, corpAuthorizeRedirect, true)
	handleFunc(mux, queryWechatCorpAuthorizerTokenPath, queryWechatCorpAuthorizerToken, true)

	server := http.Server{Addr: ":" + gokits.StrFromInt(appConfig.Port), Handler: mux}
	if err := server.ListenAndServe(); err != nil {
		gokits.LOG.Crashf("Start server Error: %s", err.Error())
	}
}

func handleFunc(mux *http.ServeMux, path string, handlerFunc http.HandlerFunc, requiredDump bool) {
	wrap := handlerFunc
	if requiredDump {
		wrap = dumpRequest(handlerFunc)
	}

	wrap = gzipHandlerFunc(wrap)
	handlePath := gokits.PathJoin(appConfig.ContextPath, path)
	mux.HandleFunc(handlePath, wrap)
}
