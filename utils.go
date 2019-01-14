package varys

import (
    "encoding/xml"
    . "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "io/ioutil"
    "net/http"
)

type WechatAppThirdPlatformAuthorizeData struct {
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

func parseWechatAppThirdPlatformAuthorizeData(codeName string, request *http.Request) (*WechatAppThirdPlatformAuthorizeData, error) {
    cache, err := wechatAppThirdPlatformCryptorCache.Value(codeName)
    if nil != err {
        LOG.Warn("Load WechatAppThirdPlatformCryptor Cache error:(%s) %s", codeName, err.Error())
        return nil, err
    }
    cryptor := cache.Data().(*wechataes.WechatCryptor)

    body, err := ioutil.ReadAll(request.Body)
    if nil != err {
        LOG.Warn("Request read Body error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    err = request.ParseForm()
    if nil != err {
        LOG.Warn("Request ParseForm error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    params := request.Form
    msgSign := params.Get("msg_signature")
    timeStamp := params.Get("timestamp")
    nonce := params.Get("nonce")
    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, string(body))
    if nil != err {
        LOG.Warn("WechatCryptor DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    LOG.Info("WechatAppThirdPlatformAuthorizeData:(%s) %s", codeName, decryptMsg)
    authorizeData := new(WechatAppThirdPlatformAuthorizeData)
    err = xml.Unmarshal([]byte(decryptMsg), authorizeData)
    if nil != err {
        LOG.Warn("Unmarshal DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    return authorizeData, nil
}

type WechatCorpThirdPlatformAuthorizeData struct {
    XMLName     xml.Name `xml:"xml"`
    SuiteId     string   `xml:"SuiteId"`
    InfoType    string   `xml:"InfoType"`
    TimeStamp   string   `xml:"TimeStamp"`
    SuiteTicket string   `xml:"SuiteTicket"`
    AuthCode    string   `xml:"AuthCode"`
    AuthCorpId  string   `xml:"AuthCorpId"`
    EchoStr     string
}

func parseWechatCorpThirdPlatformAuthorizeData(codeName string, request *http.Request) (*WechatCorpThirdPlatformAuthorizeData, error) {
    cache, err := wechatCorpThirdPlatformCryptorCache.Value(codeName)
    if nil != err {
        LOG.Warn("Load WechatCorpThirdPlatformCryptor Cache error:(%s) %s", codeName, err.Error())
        return nil, err
    }
    cryptor := cache.Data().(*wechataes.WechatCryptor)

    body, err := ioutil.ReadAll(request.Body)
    if nil != err {
        LOG.Warn("Request read Body error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    err = request.ParseForm()
    if nil != err {
        LOG.Warn("Request ParseForm error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    params := request.Form
    msgSign := params.Get("msg_signature")
    timeStamp := params.Get("timestamp")
    nonce := params.Get("nonce")
    echostr := params.Get("echostr")
    if 0 != len(echostr) { // 验证推送URL
        msg, err := cryptor.DecryptMsgContent(msgSign, timeStamp, nonce, echostr)
        if nil != err {
            LOG.Warn("WechatCryptor DecryptMsg EchoStr error:(%s) %s", codeName, err.Error())
            return nil, err
        }
        LOG.Info("WechatCorpThirdPlatformVerifyEchoStrMsg:(%s) %s", codeName, msg)
        echoData := new(WechatCorpThirdPlatformAuthorizeData)
        echoData.InfoType = "echostr"
        echoData.EchoStr = msg
        return echoData, nil
    }

    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, string(body))
    if nil != err {
        LOG.Warn("WechatCryptor DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    LOG.Info("WechatCorpThirdPlatformAuthorizeData:(%s) %s", codeName, decryptMsg)
    authorizeData := new(WechatCorpThirdPlatformAuthorizeData)
    err = xml.Unmarshal([]byte(decryptMsg), authorizeData)
    if nil != err {
        LOG.Warn("Unmarshal DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    return authorizeData, nil
}
