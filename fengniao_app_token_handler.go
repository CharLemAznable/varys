package main

import (
    "bytes"
    "crypto/md5"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
)

// /query-fengniao-app-token/{codeName:string}
const queryFengniaoAppTokenPath = "/query-fengniao-app-token/"

func queryFengniaoAppToken(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, queryFengniaoAppTokenPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := fengniaoAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*FengniaoAppToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}

// /proxy-fengniao-app/{codeName:string}/...
const proxyFengniaoAppPath = "/proxy-fengniao-app/"

func proxyFengniaoApp(writer http.ResponseWriter, request *http.Request) {
    codePath := trimPrefixPath(request, proxyFengniaoAppPath)
    splits := strings.SplitN(codePath, "/", 2)

    codeName := splits[0]
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := fengniaoAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*FengniaoAppToken)
    appId := token.AppId
    accessToken := token.AccessToken

    actualPath := splits[1]
    if "" == actualPath {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "proxy PATH is Empty"}))
        return
    }

    data, err := gokits.RequestBody(request) // 向代理发起的请求仅需提供data字段
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }

    // 改写创建蜂鸟订单接口时传递的notify_url参数
    if "order" == actualPath && "" != globalConfig.FengniaoCallbackAddress {
        dataMap := make(map[string]interface{})
        gokits.UnJson(data, &dataMap)
        notifyUrl := globalConfig.FengniaoCallbackAddress + gokits.PathJoin(
            globalConfig.ContextPath, callbackFengniaoOrderPath, codeName)
        dataMap["notify_url"] = notifyUrl
        data = gokits.Json(dataMap)
    }

    salt := newSalt()
    plainText := "app_id=" + appId + "&access_token=" + accessToken +
        "&data=" + url.QueryEscape(data) + "&salt=" + gokits.StrFromInt(salt)    // 拼接明文
    signature := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))                   // 签名
    body := bytes.NewBuffer([]byte(gokits.Json(map[string]interface{}{
        "app_id": appId, "data": data, "salt": salt, "signature": signature,}))) // 代理请求体

    // 重写请求体, ref: http/request.go(line:838)
    req := request
    req.URL.Path = actualPath
    req.Body = ioutil.NopCloser(body)
    req.ContentLength = int64(body.Len())
    bodyBuf := body.Bytes()
    req.GetBody = func() (io.ReadCloser, error) {
        r := bytes.NewReader(bodyBuf)
        return ioutil.NopCloser(r), nil
    }
    fengniaoAppProxy.ServeHTTP(writer, req)
}

// /callback-fengniao-order/{codeName:string}
const callbackFengniaoOrderPath = "/callback-fengniao-order/"

func callbackFengniaoOrder(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, callbackFengniaoOrderPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := fengniaoAppTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "get access_token error"}))
        return
    }
    token := cache.Data().(*FengniaoAppToken)
    appId := token.AppId
    accessToken := token.AccessToken

    body, err := gokits.RequestBody(request)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "read request body error"}))
        return
    }

    bodyMap := make(map[string]interface{})
    gokits.UnJson(body, &bodyMap)
    if appId != bodyMap["app_id"].(string) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "app_id mismatch"}))
        return
    }
    data := bodyMap["data"].(string)
    salt := bodyMap["salt"].(int)
    signature := bodyMap["signature"].(string)

    plainText := "app_id=" + appId + "&access_token=" + accessToken +
        "&data=" + data + "&salt=" + gokits.StrFromInt(salt) // 拼接明文
    sign := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))    // 验证签名
    if signature != fmt.Sprintf("%x", md5.Sum([]byte(plainText))) {
        golog.Warnf("signature %s != %s mismatch with access_token: %s", signature, sign, accessToken)
    }

    callbackData, err := url.QueryUnescape(data)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "url decode data error"}))
        return
    }

    configCache, err := fengniaoAppConfigCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "get config error"}))
        return
    }
    config := configCache.Data().(*FengniaoAppConfig)
    callbackOrderUrl := config.CallbackOrderUrl

    if "" != callbackOrderUrl {
        rsp, err := gokits.NewHttpReq(callbackOrderUrl).
            RequestBody(callbackData).
            Prop("Content-Type", "application/json").Post()
        if nil != err {
            golog.Errorf("Callback Error: %s", err.Error())
        }
        golog.Debugf("Callback Response: %s", rsp)
    }

    gokits.ResponseJson(writer, gokits.Json(map[string]string{"result": "OK"}))
}
