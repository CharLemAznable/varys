package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http"
    "strings"
)

// /query-wechat-corp-token/{codeName:string}
const queryWechatCorpTokenPath = "/query-wechat-corp-token/"

func queryWechatCorpToken(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, queryWechatCorpTokenPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatCorpTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatCorpToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}

// /proxy-wechat-corp/{codeName:string}/...
const proxyWechatCorpPath = "/proxy-wechat-corp/"

func proxyWechatCorp(writer http.ResponseWriter, request *http.Request) {
    codePath := trimPrefixPath(request, proxyWechatCorpPath)
    splits := strings.SplitN(codePath, "/", 2)

    codeName := splits[0]
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatCorpTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatCorpToken).AccessToken

    actualPath := splits[1]
    if "" == actualPath {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "proxy PATH is Empty"}))
        return
    }

    req := request
    if req.URL.RawQuery == "" {
        req.URL.RawQuery = req.URL.RawQuery + "access_token=" + token
    } else {
        req.URL.RawQuery = req.URL.RawQuery + "&" + "access_token=" + token
    }
    req.URL.Path = actualPath
    wechatCorpProxy.ServeHTTP(writer, req)
}
