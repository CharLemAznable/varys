package app

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

// 可配置测试环境: http://open.s.bingex.com/auth
var authBaseURL = "http://open.ishansong.com/auth"
// 可配置测试环境: http://open.s.bingex.com/openapi/oauth/token
var tokenURL = "http://open.ishansong.com/openapi/oauth/token"
// 可配置测试环境: http://open.s.bingex.com/openapi/oauth/refresh_token
var refreshTokenURL = "http://open.ishansong.com/openapi/oauth/refresh_token"

var configLifeSpan = time.Minute * 60   // config cache 60 min default
var tokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var tokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

// 可配置测试环境: http://open.s.bingex.com/openapi/developer/v5/
const DefaultDeveloperProxyURL = "http://open.ishansong.com/openapi/developer/v5/"

// 可配置测试环境: http://open.s.bingex.com/openapi/merchantRegister/v5/
const DefaultMerchantProxyURL = "http://open.ishansong.com/openapi/merchantRegister/v5/"

// 可配置测试环境: http://open.s.bingex.com/openapi/file/v5/
const DefaultFileProxyURL = "http://open.ishansong.com/openapi/file/v5/"

var developerProxyURL = DefaultDeveloperProxyURL
var merchantProxyURL = DefaultMerchantProxyURL
var fileProxyURL = DefaultFileProxyURL

type Config struct {
    ShansongAppAuthBaseURL       string
    ShansongAppTokenURL          string
    ShansongAppRefreshTokenURL   string
    ShansongAppConfigLifeSpan    gokits.Duration
    ShansongAppTokenLifeSpan     gokits.Duration
    ShansongAppTokenTempLifeSpan gokits.Duration
    ShansongAppDeveloperProxyURL string
    ShansongAppMerchantProxyURL  string
    ShansongAppFileProxyURL      string
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
    gokits.If("" != config.ShansongAppAuthBaseURL, func() {
        authBaseURL = config.ShansongAppAuthBaseURL
    })
    gokits.If("" != config.ShansongAppTokenURL, func() {
        tokenURL = config.ShansongAppTokenURL
    })
    gokits.If("" != config.ShansongAppRefreshTokenURL, func() {
        refreshTokenURL = config.ShansongAppRefreshTokenURL
    })

    gokits.If(0 != config.ShansongAppConfigLifeSpan.Duration, func() {
        configLifeSpan = config.ShansongAppConfigLifeSpan.Duration
    })
    gokits.If(0 != config.ShansongAppTokenLifeSpan.Duration, func() {
        tokenLifeSpan = config.ShansongAppTokenLifeSpan.Duration
    })
    gokits.If(0 != config.ShansongAppTokenTempLifeSpan.Duration, func() {
        tokenTempLifeSpan = config.ShansongAppTokenTempLifeSpan.Duration
    })

    gokits.If("" != config.ShansongAppDeveloperProxyURL, func() {
        developerProxyURL = config.ShansongAppDeveloperProxyURL
    })
    gokits.If("" != config.ShansongAppMerchantProxyURL, func() {
        merchantProxyURL = config.ShansongAppMerchantProxyURL
    })
    gokits.If("" != config.ShansongAppFileProxyURL, func() {
        fileProxyURL = config.ShansongAppFileProxyURL
    })

    golog.Infof("shansong/app config: %+v", *config)
}
