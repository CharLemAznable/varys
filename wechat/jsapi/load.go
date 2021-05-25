package jsapi

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
)

var getticketURL = "https://api.weixin.qq.com/cgi-bin/ticket/getticket"

type Config struct {
    WechatJsapiTicketURL string
}

var config = &Config{}

func init() {
    base.RegisterLoader(func(configFile string) {
        base.LoadConfig(configFile, config)
        fixConfig()
    })
}

func fixConfig() {
    gokits.If("" != config.WechatJsapiTicketURL, func() {
        getticketURL = config.WechatJsapiTicketURL
    })

    golog.Infof("wechat/jsapi config: %+v", *config)
}
