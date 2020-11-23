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
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatAppToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}

// /proxy-wechat-app/{codeName:string}/...
const proxyWechatAppPath = "/proxy-wechat-app/"

func proxyWechatApp(writer http.ResponseWriter, request *http.Request) {
    codePath := trimPrefixPath(request, proxyWechatAppPath)
    splits := strings.SplitN(codePath, "/", 2)

    codeName := splits[0]
    if "" == codeName {
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
    wechatAppProxy.ServeHTTP(writer, req)
}

// /proxy-wechat-app-mp-login/{codeName:string}?js_code=JSCODE
const proxyWechatAppMpLoginPath = "/proxy-wechat-app-mp-login/"

func proxyWechatAppMpLogin(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, proxyWechatAppMpLoginPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatAppConfigCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    config := cache.Data().(*WechatAppConfig)

    if "" == request.URL.Query().Get("js_code") {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "js_code is Empty"}))
        return
    }

    req := request
    req.URL.RawQuery = req.URL.RawQuery +
        "&appid=" + config.AppId +
        "&secret=" + config.AppSecret +
        "&grant_type=authorization_code"
    req.URL.Path = "jscode2session"
    wechatAppMpLoginProxy.ServeHTTP(writer, req)
}

// /query-wechat-app-js-config/{codeName:string}?url=URL
const queryWechatAppJsConfigPath = "/query-wechat-app-js-config/"

func queryWechatAppJsConfig(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, queryWechatAppJsConfigPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatAppToken)
    appId := token.AppId
    jsapiTicket := token.JsapiTicket
    if "" == jsapiTicket {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "jsapi_ticket is Empty"}))
        return
    }

    url := request.URL.Query().Get("url")
    if "" == url {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "url is Empty"}))
        return
    }

    gokits.ResponseJson(writer, gokits.Json(
        wechatJsConfigBuilder(appId, jsapiTicket, url)))
}
