package apptp

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var tokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

var configLifeSpan = time.Minute * 60   // config cache 60 min default
var cryptorLifeSpan = time.Minute * 60  // cryptor cache 60 min default
var tokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var tokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

const DefaultProxyURL = "https://api.weixin.qq.com/"

var proxyURL = DefaultProxyURL

type Config struct {
    WechatTpTokenURL          string
    WechatTpConfigLifeSpan    gokits.Duration
    WechatTpCryptorLifeSpan   gokits.Duration
    WechatTpTokenLifeSpan     gokits.Duration
    WechatTpTokenTempLifeSpan gokits.Duration
    WechatTpProxyURL          string
}

var config = &Config{}

func init() {
    base.RegisterLoader(func(configFile string) {
        base.LoadConfig(configFile, config)
        fixConfig()

        cacheInitialize()
        proxyInitialize()
    })
}

func fixConfig() {
    gokits.If("" != config.WechatTpTokenURL, func() {
        tokenURL = config.WechatTpTokenURL
    })

    gokits.If(0 != config.WechatTpConfigLifeSpan.Duration, func() {
        configLifeSpan = config.WechatTpConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatTpCryptorLifeSpan.Duration, func() {
        cryptorLifeSpan = config.WechatTpCryptorLifeSpan.Duration
    })
    gokits.If(0 != config.WechatTpTokenLifeSpan.Duration, func() {
        tokenLifeSpan = config.WechatTpTokenLifeSpan.Duration
    })
    gokits.If(0 != config.WechatTpTokenTempLifeSpan.Duration, func() {
        tokenTempLifeSpan = config.WechatTpTokenTempLifeSpan.Duration
    })

    gokits.If("" != config.WechatTpProxyURL, func() {
        proxyURL = config.WechatTpProxyURL
    })

    golog.Infof("wechat/apptp config: %+v", *config)
}
