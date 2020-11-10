package main

import (
    "encoding/xml"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "net/http"
    "strings"
)

type WechatTpInfoData struct {
    XMLName                      xml.Name `xml:"xml"`
    AppId                        string   `xml:"AppId"`
    CreateTime                   string   `xml:"CreateTime"`
    InfoType                     string   `xml:"InfoType"`
    ComponentVerifyTicket        string   `xml:"ComponentVerifyTicket"`
    AuthorizerAppid              string   `xml:"AuthorizerAppid"`
    AuthorizationCode            string   `xml:"AuthorizationCode"`
    AuthorizationCodeExpiredTime string   `xml:"AuthorizationCodeExpiredTime"`
    PreAuthCode                  string   `xml:"PreAuthCode"`
}

func parseWechatTpInfoData(codeName string, request *http.Request) (*WechatTpInfoData, error) {
    cache, err := wechatTpCryptorCache.Value(codeName)
    if nil != err {
        _ = gokits.LOG.Warn("Load WechatTpCryptor Cache error:(%s) %s", codeName, err.Error())
        return nil, err
    }
    cryptor := cache.Data().(*wechataes.WechatCryptor)

    body, err := gokits.RequestBody(request)
    if nil != err {
        _ = gokits.LOG.Warn("Request read Body error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    err = request.ParseForm()
    if nil != err {
        _ = gokits.LOG.Warn("Request ParseForm error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    params := request.Form
    msgSign := params.Get("msg_signature")
    timeStamp := params.Get("timestamp")
    nonce := params.Get("nonce")
    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, body)
    if nil != err {
        _ = gokits.LOG.Warn("WechatCryptor DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    gokits.LOG.Info("WechatTpInfoData:(%s) %s", codeName, decryptMsg)
    infoData := new(WechatTpInfoData)
    err = xml.Unmarshal([]byte(decryptMsg), infoData)
    if nil != err {
        _ = gokits.LOG.Warn("Unmarshal DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    return infoData, nil
}

// /accept-wechant-tp-info/{codeName:string}
const acceptWechatTpInfoPath = "/accept-wechant-tp-info/"

func acceptWechatTpInfo(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, acceptWechatTpInfoPath)
    if 0 != len(codeName) {
        infoData, err := parseWechatTpInfoData(codeName, request)
        if nil == err {
            switch infoData.InfoType {

            case "component_verify_ticket":
                _, _ = updateWechatTpTicket(codeName, infoData.ComponentVerifyTicket)

            case "authorized":
                wechatTpAuthorized(codeName, infoData)

            case "updateauthorized":
                wechatTpAuthorized(codeName, infoData)

            case "unauthorized":
                wechatTpUnauthorized(codeName, infoData)

            }
        }
    }
    // 接收到定时推送component_verify_ticket后必须直接返回字符串success
    gokits.ResponseText(writer, "success")
}

// /query-wechat-tp-token/{codeName:string}
const queryWechatTpTokenPath = "/query-wechat-tp-token/"

func queryWechatTpToken(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, queryWechatTpTokenPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatTpTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpToken)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{"appId": token.AppId, "token": token.AccessToken}))
}

// /proxy-wechat-tp/{codeName:string}/...
const proxyWechatTpPath = "/proxy-wechat-tp/"

func proxyWechatTp(writer http.ResponseWriter, request *http.Request) {
    codePath := trimPrefixPath(request, proxyWechatTpPath)
    splits := strings.SplitN(codePath, "/", 2)

    codeName := splits[0]
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatTpTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpToken).AccessToken

    actualPath := splits[1]
    if 0 == len(actualPath) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "proxy PATH is Empty"}))
        return
    }

    req := request
    if req.URL.RawQuery == "" {
        req.URL.RawQuery = req.URL.RawQuery + "component_access_token=" + token
    } else {
        req.URL.RawQuery = req.URL.RawQuery + "&" + "component_access_token=" + token
    }
    req.URL.Path = actualPath
    wechatTpProxy.ServeHTTP(writer, req)
}
