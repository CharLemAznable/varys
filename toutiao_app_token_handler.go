package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http"
)

// /query-toutiao-app-token/{codeName:string}
const queryToutiaoAppTokenPath = "/query-toutiao-app-token/"

func queryToutiaoAppToken(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, queryToutiaoAppTokenPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := toutiaoAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*ToutiaoAppToken)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{"appId": token.AppId, "token": token.AccessToken}))
}
