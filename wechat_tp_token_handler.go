package main

import (
    "encoding/xml"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "github.com/kataras/golog"
    "net/http"
    "strings"
)

type WechatTpInfoData struct {
    XMLName                      xml.Name `xml:"xml" json:"-"`
    AppId                        string   `xml:"AppId"`
    CreateTime                   string   `xml:"CreateTime"`
    InfoType                     string   `xml:"InfoType"`
    ComponentVerifyTicket        string   `xml:"ComponentVerifyTicket"`
    AuthorizerAppId              string   `xml:"AuthorizerAppid"`
    AuthorizationCode            string   `xml:"AuthorizationCode"`
    AuthorizationCodeExpiredTime string   `xml:"AuthorizationCodeExpiredTime"`
    PreAuthCode                  string   `xml:"PreAuthCode"`

    // InfoType == "notify_third_fasteregister" 快速创建小程序 事件回调通知
    MpAppId    string         `xml:"appid"`     // 创建小程序appid
    MpStatus   string         `xml:"status"`    // 0
    MpAuthCode string         `xml:"auth_code"` // 第三方授权码
    MpMsg      string         `xml:"msg"`       // OK
    MpInfo     WechatTpMpInfo `xml:"info"`
}

type WechatTpMpInfo struct {
    MpName               string `xml:"name"`                 // 企业名称
    MpCode               string `xml:"code"`                 // 企业代码
    MpCodeType           string `xml:"code_type"`            // 企业代码类型
    MpLegalPersonaWechat string `xml:"legal_persona_wechat"` // 法人微信号
    MpLegalPersonaName   string `xml:"legal_persona_name"`   // 法人姓名
    MpComponentPhone     string `xml:"component_phone"`      // 第三方联系电话
}

func parseWechatTpInfoData(codeName string, request *http.Request) (*WechatTpInfoData, error) {
    cache, err := wechatTpCryptorCache.Value(codeName)
    if nil != err {
        golog.Warnf("Load WechatTpCryptor Cache error:(%s) %s", codeName, err.Error())
        return nil, err
    }
    cryptor := cache.Data().(*wechataes.WechatCryptor)

    body, err := gokits.RequestBody(request)
    if nil != err {
        golog.Warnf("Request read Body error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    err = request.ParseForm()
    if nil != err {
        golog.Warnf("Request ParseForm error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    params := request.Form
    msgSign := params.Get("msg_signature")
    timeStamp := params.Get("timestamp")
    nonce := params.Get("nonce")
    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, body)
    if nil != err {
        golog.Warnf("WechatCryptor DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    golog.Debugf("WechatTpInfoData:(%s) %s", codeName, decryptMsg)
    infoData := new(WechatTpInfoData)
    err = xml.Unmarshal([]byte(decryptMsg), infoData)
    if nil != err {
        golog.Warnf("Unmarshal DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    return infoData, nil
}

// /accept-wechat-tp-info/{codeName:string}
const acceptWechatTpInfoPath = "/accept-wechat-tp-info/"

func acceptWechatTpInfo(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, acceptWechatTpInfoPath)
    if "" != codeName {
        infoData, err := parseWechatTpInfoData(codeName, request)
        go func() {
            if nil == err {
                switch infoData.InfoType {

                case "component_verify_ticket":
                    _, _ = db.NamedExec(updateWechatTpTicketSQL,
                        map[string]interface{}{"CodeName": codeName,
                            "Ticket": infoData.ComponentVerifyTicket})

                case "authorized":
                    wechatTpAuthorized(codeName, infoData)

                case "updateauthorized":
                    wechatTpAuthorized(codeName, infoData)

                case "unauthorized":
                    wechatTpUnauthorized(codeName, infoData)

                case "notify_third_fasteregister":
                    wechatTpAuthorizedMp(codeName, infoData)
                }

                forwardWechatTpInfo(codeName, infoData)
            }
        }()
    }
    // 直接返回字符串success
    gokits.ResponseText(writer, "success")
}

// 将第三方平台授权事件推送转发到业务服务
func forwardWechatTpInfo(codeName string, infoData *WechatTpInfoData) {
    cache, err := wechatTpConfigCache.Value(codeName)
    if nil != err {
        golog.Errorf("Accept WechatTp Info: CodeName %s is Illegal", codeName)
        return
    }
    config := cache.Data().(*WechatTpConfig)
    forwardUrl := config.AuthForwardUrl

    if "" != forwardUrl {
        rsp, err := gokits.NewHttpReq(forwardUrl).
            RequestBody(gokits.Json(infoData)).
            Prop("Content-Type", "application/json").Post()
        if nil != err {
            golog.Errorf("Forward Error: %s", err.Error())
        }
        golog.Debugf("Forward Response: %s", rsp)
    }
}

// /query-wechat-tp-token/{codeName:string}
const queryWechatTpTokenPath = "/query-wechat-tp-token/"

func queryWechatTpToken(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, queryWechatTpTokenPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    cache, err := wechatTpTokenCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatTpToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}

// /proxy-wechat-tp/{codeName:string}/...
const proxyWechatTpPath = "/proxy-wechat-tp/"

func proxyWechatTp(writer http.ResponseWriter, request *http.Request) {
    codePath := trimPrefixPath(request, proxyWechatTpPath)
    splits := strings.SplitN(codePath, "/", 2)

    codeName := splits[0]
    if "" == codeName {
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
    if "" == actualPath {
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
