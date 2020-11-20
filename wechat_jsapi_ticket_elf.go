package main

import (
    "crypto/sha1"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "time"
)

type WechatJsapiTicketResponse struct {
    Errcode   int    `json:"errcode"`
    Errmsg    string `json:"errmsg"`
    Ticket    string `json:"ticket"`
    ExpiresIn int    `json:"expires_in"`
}

func wechatJsapiTicketRequestor(codeName, accessToken string) string {
    // request ticket maybe failed, maybe wechat mini app
    ticketResult, err := gokits.NewHttpReq(wechatJsapiTicketURL).Params(
        "type", "jsapi", "access_token", accessToken).
        Prop("Content-Type", "application/x-www-form-urlencoded").Get()
    golog.Debugf("Request WechatJsapiTicket Response:(%s) %s", codeName, ticketResult)
    if nil != err {
        golog.Warnf("Request WechatJsapiTicket Error: %s", err.Error())
    }

    ticketResponse := new(WechatJsapiTicketResponse)
    gokits.UnJson(ticketResult, ticketResponse)
    if "" == ticketResponse.Ticket {
        golog.Warnf("Request WechatJsapiTicket Error: %d - %s", ticketResponse.Errcode, ticketResponse.Errmsg)
    }
    return ticketResponse.Ticket
}

type WechatJsConfig struct {
    AppId     string `json:"appId"`
    NonceStr  string `json:"nonceStr"`
    Timestamp int64  `json:"timestamp"`
    Signature string `json:"signature"`
}

func wechatJsConfigBuilder(appId, jsapiTicket, url string) *WechatJsConfig {
    noncestr := gokits.RandomString(32)
    timestamp := time.Now().Unix()
    plainText := "jsapi_ticket=" + jsapiTicket + "&noncestr=" + noncestr +
        "&timestamp=" + gokits.StrFromInt64(timestamp) + "&url=" + url
    signature := fmt.Sprintf("%x", sha1.Sum([]byte(plainText)))
    return &WechatJsConfig{AppId: appId, NonceStr: noncestr,
        Timestamp: timestamp, Signature: signature}
}
