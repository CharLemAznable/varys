package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http"
    "strings"
)

// /query-wechat-app-token/{codeName:string}
const queryWechatAppTokenPath = "/query-wechat-app-token/"

func queryWechatAppToken(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, queryWechatAppTokenPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatAppToken)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{"appId": token.AppId, "token": token.AccessToken}))
}

// /proxy-wechat-app/{codeName:string}/...
const proxyWechatAppPath = "/proxy-wechat-app/"

func proxyWechatApp(writer http.ResponseWriter, request *http.Request) {
    codePath := trimPrefixPath(request, proxyWechatAppPath)
    splits := strings.SplitN(codePath, "/", 2)

    codeName := splits[0]
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatAppToken).AccessToken

    actualPath := splits[1]
    if 0 == len(actualPath) {
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
    wechatAppProxy.ServeHTTP(writer, req)
}
