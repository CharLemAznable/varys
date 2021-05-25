package app

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var tokenURL = "https://api.weixin.qq.com/cgi-bin/token"

var configLifeSpan = time.Minute * 60   // config cache 60 min default
var tokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var tokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

const DefaultProxyURL = "https://api.weixin.qq.com/"
const DefaultMpLoginProxyURL = "https://api.weixin.qq.com/sns/"

var proxyURL = DefaultProxyURL
var mpLoginProxyURL = DefaultMpLoginProxyURL

type Config struct {
    WechatAppTokenURL          string
    WechatAppConfigLifeSpan    gokits.Duration
    WechatAppTokenLifeSpan     gokits.Duration
    WechatAppTokenTempLifeSpan gokits.Duration
    WechatAppProxyURL          string
    WechatAppMpLoginProxyURL   string
}

var config = &Config{}

func init() {
    base.RegisterLoader(func(configFile string) {
        base.LoadConfig(configFile, config)
        fixConfig()

        cacheInitialize()
        proxyInitialize()
        mpLoginProxyInitialize()
    })
}

func fixConfig() {
    gokits.If("" != config.WechatAppTokenURL, func() {
        tokenURL = config.WechatAppTokenURL
    })

    gokits.If(0 != config.WechatAppConfigLifeSpan.Duration, func() {
        configLifeSpan = config.WechatAppConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatAppTokenLifeSpan.Duration, func() {
        tokenLifeSpan = config.WechatAppTokenLifeSpan.Duration
    })
    gokits.If(0 != config.WechatAppTokenTempLifeSpan.Duration, func() {
        tokenTempLifeSpan = config.WechatAppTokenTempLifeSpan.Duration
    })

    gokits.If("" != config.WechatAppProxyURL, func() {
        proxyURL = config.WechatAppProxyURL
    })
    gokits.If("" != config.WechatAppMpLoginProxyURL, func() {
        mpLoginProxyURL = config.WechatAppMpLoginProxyURL
    })

    golog.Infof("wechat/app config: %+v", *config)
}
