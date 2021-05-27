package app

import (
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "net/http"
)

func init() {
    RegisterHandler(func(mux *http.ServeMux) {
        gokits.HandleFunc(mux, queryToutiaoAppTokenPath, queryToutiaoAppToken)
    })
}

// /query-toutiao-app-token/{codeName:string}
const queryToutiaoAppTokenPath = "/query-toutiao-app-token/"

func queryToutiaoAppToken(writer http.ResponseWriter, request *http.Request) {
    codeName := TrimPrefixPath(request, queryToutiaoAppTokenPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := tokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*ToutiaoAppToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}
