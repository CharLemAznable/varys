package corptp

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var tokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_suite_token"

var configLifeSpan = time.Minute * 60         // config cache 60 min default
var cryptorLifeSpan = time.Minute * 60        // cryptor cache 60 min default
var tokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var tokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

type Config struct {
    WechatCorpTpTokenURL                string
    WechatCorpTpConfigLifeSpan          gokits.Duration
    WechatCorpTpCryptorLifeSpan         gokits.Duration
    WechatCorpTpTokenMaxLifeSpan        gokits.Duration
    WechatCorpTpTokenExpireCriticalSpan gokits.Duration
}

var config = &Config{}

func init() {
    base.RegisterLoader(func(configFile string) {
        base.LoadConfig(configFile, config)
        fixConfig()

        cacheInitialize()
    })
}

func fixConfig() {
    gokits.If("" != config.WechatCorpTpTokenURL, func() {
        tokenURL = config.WechatCorpTpTokenURL
    })

    gokits.If(0 != config.WechatCorpTpConfigLifeSpan.Duration, func() {
        configLifeSpan = config.WechatCorpTpConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpCryptorLifeSpan.Duration, func() {
        cryptorLifeSpan = config.WechatCorpTpCryptorLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpTokenMaxLifeSpan.Duration, func() {
        tokenMaxLifeSpan = config.WechatCorpTpTokenMaxLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpTokenExpireCriticalSpan.Duration, func() {
        tokenExpireCriticalSpan = config.WechatCorpTpTokenExpireCriticalSpan.Duration
    })

    golog.Infof("wechat/corptp config: %+v", *config)
}
