package corptp

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var preAuthCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_pre_auth_code?suite_access_token="
var permanentCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_permanent_code?suite_access_token="
var authTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_corp_token?suite_access_token="

var permanentCodeLifeSpan = time.Minute * 60      // permanent_code cache 60 min default
var authTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var authTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

type AuthConfig struct {
    WechatCorpTpPreAuthCodeURL              string
    WechatCorpTpPermanentCodeURL            string
    WechatCorpTpAuthTokenURL                string
    WechatCorpTpPermanentCodeLifeSpan       gokits.Duration
    WechatCorpTpAuthTokenMaxLifeSpan        gokits.Duration
    WechatCorpTpAuthTokenExpireCriticalSpan gokits.Duration
}

var authConfig = &AuthConfig{}

func init() {
    base.RegisterLoader(func(configFile string) {
        base.LoadConfig(configFile, authConfig)
        fixAuthConfig()

        authCacheInitialize()
    })
}

func fixAuthConfig() {
    gokits.If("" != authConfig.WechatCorpTpPreAuthCodeURL, func() {
        preAuthCodeURL = authConfig.WechatCorpTpPreAuthCodeURL
    })
    gokits.If("" != authConfig.WechatCorpTpPermanentCodeURL, func() {
        permanentCodeURL = authConfig.WechatCorpTpPermanentCodeURL
    })
    gokits.If("" != authConfig.WechatCorpTpAuthTokenURL, func() {
        authTokenURL = authConfig.WechatCorpTpAuthTokenURL
    })

    gokits.If(0 != authConfig.WechatCorpTpPermanentCodeLifeSpan.Duration, func() {
        permanentCodeLifeSpan = authConfig.WechatCorpTpPermanentCodeLifeSpan.Duration
    })
    gokits.If(0 != authConfig.WechatCorpTpAuthTokenMaxLifeSpan.Duration, func() {
        authTokenMaxLifeSpan = authConfig.WechatCorpTpAuthTokenMaxLifeSpan.Duration
    })
    gokits.If(0 != authConfig.WechatCorpTpAuthTokenExpireCriticalSpan.Duration, func() {
        authTokenExpireCriticalSpan = authConfig.WechatCorpTpAuthTokenExpireCriticalSpan.Duration
    })

    golog.Infof("wechat/corptp/auth config: %+v", *config)
}
