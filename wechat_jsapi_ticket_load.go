package main

import (
    "github.com/CharLemAznable/gokits"
)

var wechatJsapiTicketURL = "https://api.weixin.qq.com/cgi-bin/ticket/getticket"

func wechatJsapiTokenLoad(config *Config) {
    gokits.If("" != config.WechatJsapiTicketURL, func() {
        wechatJsapiTicketURL = config.WechatJsapiTicketURL
    })
}
