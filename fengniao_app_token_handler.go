package main

import (
    "bytes"
    "crypto/md5"
    "fmt"
    "github.com/CharLemAznable/gokits"
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
