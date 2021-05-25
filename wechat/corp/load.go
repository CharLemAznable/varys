package corp

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var tokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"

var configLifeSpan = time.Minute * 60         // config cache 60 min default
var tokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var tokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

const DefaultProxyURL = "https://qyapi.weixin.qq.com/"

var proxyURL = DefaultProxyURL

type Config struct {
    WechatCorpTokenURL                string
    WechatCorpConfigLifeSpan          gokits.Duration
    WechatCorpTokenMaxLifeSpan        gokits.Duration
    WechatCorpTokenExpireCriticalSpan gokits.Duration
    WechatCorpProxyURL                string
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
    gokits.If("" != config.WechatCorpTokenURL, func() {
        tokenURL = config.WechatCorpTokenURL
    })

    gokits.If(0 != config.WechatCorpConfigLifeSpan.Duration, func() {
        configLifeSpan = config.WechatCorpConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTokenMaxLifeSpan.Duration, func() {
        tokenMaxLifeSpan = config.WechatCorpTokenMaxLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTokenExpireCriticalSpan.Duration, func() {
        tokenExpireCriticalSpan = config.WechatCorpTokenExpireCriticalSpan.Duration
    })

    gokits.If("" != config.WechatCorpProxyURL, func() {
        proxyURL = config.WechatCorpProxyURL
    })

    golog.Infof("wechat/corp config: %+v", *config)
}
