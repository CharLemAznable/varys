package varys

import (
    "encoding/xml"
    log "github.com/CharLemAznable/log4go"
    "github.com/CharLemAznable/wechataes"
    "io/ioutil"
    "net/http"
)

type WechatAuthorizeData struct {
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

func parseWechatAuthorizeData(codeName string, request *http.Request) (*WechatAuthorizeData, error) {
    cache, err := wechatThirdPlatformCryptorCache.Value(codeName)
    if nil != err {
        log.Warn("GetWechatThirdPlatformCryptor error:", err.Error())
        return nil, err
    }
    cryptor := cache.Data().(*wechataes.WechatCryptor)

    body, err := ioutil.ReadAll(request.Body)
    if nil != err {
        log.Warn("Request read Body error:", err.Error())
        return nil, err
    }

    err = request.ParseForm()
    if nil != err {
        log.Warn("Request ParseForm error:", err.Error())
        return nil, err
    }

    params := request.Form
    msgSign := params.Get("msg_signature")
    timeStamp := params.Get("timestamp")
    nonce := params.Get("nonce")
    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, string(body))
    if nil != err {
        log.Warn("WechatCryptor DecryptMsg error:", err.Error())
        return nil, err
    }

    log.Info("微信推送消息(明文):", decryptMsg)
    authorizeData := new(WechatAuthorizeData)
    err = xml.Unmarshal([]byte(decryptMsg), authorizeData)
    if nil != err {
        log.Warn("Unmarshal DecryptMsg error:", err.Error())
        return nil, err
    }

    return authorizeData, nil
}
