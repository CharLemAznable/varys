package main

import (
    "encoding/xml"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "github.com/kataras/golog"
    "net/http"
)

type WechatCorpTpInfoData struct {
    XMLName     xml.Name `xml:"xml"`
    SuiteId     string   `xml:"SuiteId"`
    InfoType    string   `xml:"InfoType"`
    TimeStamp   string   `xml:"TimeStamp"`
    SuiteTicket string   `xml:"SuiteTicket"`
    AuthCode    string   `xml:"AuthCode"`
    AuthCorpId  string   `xml:"AuthCorpId"`
    EchoStr     string
}

func parseWechatCorpTpInfoData(codeName string, request *http.Request) (*WechatCorpTpInfoData, error) {
    cache, err := wechatCorpTpCryptorCache.Value(codeName)
    if nil != err {
        golog.Warnf("Load WechatCorpTpCryptor Cache error:(%s) %s", codeName, err.Error())
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
    echostr := params.Get("echostr")
    if 0 != len(echostr) { // 验证推送URL
        msg, err := cryptor.DecryptMsgContent(msgSign, timeStamp, nonce, echostr)
        if nil != err {
            golog.Warnf("WechatCryptor DecryptMsg EchoStr error:(%s) %s", codeName, err.Error())
            return nil, err
        }
        golog.Infof("WechatCorpTpVerifyEchoStrMsg:(%s) %s", codeName, msg)
        echoData := new(WechatCorpTpInfoData)
        echoData.InfoType = "echostr"
        echoData.EchoStr = msg
        return echoData, nil
    }

    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, body)
    if nil != err {
        golog.Warnf("WechatCryptor DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    gokits.LOG.Info("WechatCorpTpInfoData:(%s) %s", codeName, decryptMsg)
    infoData := new(WechatCorpTpInfoData)
    err = xml.Unmarshal([]byte(decryptMsg), infoData)
    if nil != err {
        golog.Warnf("Unmarshal DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    return infoData, nil
}

// /accept-wechat-corp-tp-info/{codeName:string}
const acceptWechatCorpTpInfoPath = "/accept-wechat-corp-tp-info/"

func acceptWechatCorpTpInfo(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, acceptWechatCorpTpInfoPath)
    if 0 != len(codeName) {
        infoData, err := parseWechatCorpTpInfoData(codeName, request)
        if nil == err {
            switch infoData.InfoType {

            case "echostr": // 验证推送URL
                gokits.ResponseText(writer, infoData.EchoStr)
                return

            case "suite_ticket":
                _, _ = db.NamedExec(updateWechatCorpTpTicketSQL,
                    map[string]interface{}{"CodeName": codeName,
                        "Ticket": infoData.SuiteTicket})

            case "create_auth":
                wechatCorpTpCreateAuth(codeName, infoData)

            case "change_auth":
                // ignore

            case "cancel_auth":
                wechatCorpTpCancelAuth(codeName, infoData)

            }
        }
    }
    gokits.ResponseText(writer, "success")
}
