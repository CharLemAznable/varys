package app

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

// 可配置沙箱环境: https://exam-anubis.ele.me/anubis-webapi/openapi/token
var tokenURL = "https://open-anubis.ele.me/anubis-webapi/openapi/token"
// 可配置沙箱环境: https://exam-anubis.ele.me/anubis-webapi/openapi/refreshToken
var refreshURL = "https://open-anubis.ele.me/anubis-webapi/openapi/refreshToken"

// token刷新接口10分钟有效期，如果10分钟内多次调用，
// 只有第一次会刷新access_token和refresh_token，
// 后续调用会返回第一次刷新的access_token和refresh_token。
// 所以需要保证token缓存有效期在10分钟内。
const MaxTokenLifeSpan = time.Minute * 8

var configLifeSpan = time.Minute * 60 // config cache 60 min default
var tokenLifeSpan = MaxTokenLifeSpan  // stable token cache 8 min default

// 可配置沙箱环境: https://exam-anubis.ele.me/anubis-webapi/v3/invoke/
const DefaultProxyURL = "https://open-anubis.ele.me/anubis-webapi/v3/invoke/"

var proxyURL = DefaultProxyURL

type Config struct {
    FengniaoAppTokenURL       string
    FengniaoAppRefreshURL     string
    FengniaoAppConfigLifeSpan gokits.Duration
    FengniaoAppTokenLifeSpan  gokits.Duration
    FengniaoAppProxyURL       string
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
    gokits.If("" != config.FengniaoAppTokenURL, func() {
        tokenURL = config.FengniaoAppTokenURL
    })
    gokits.If("" != config.FengniaoAppRefreshURL, func() {
        refreshURL = config.FengniaoAppRefreshURL
    })

    gokits.If(0 != config.FengniaoAppConfigLifeSpan.Duration, func() {
        configLifeSpan = config.FengniaoAppConfigLifeSpan.Duration
    })
    gokits.If(0 != config.FengniaoAppTokenLifeSpan.Duration, func() {
        setDuration := config.FengniaoAppTokenLifeSpan.Duration
        tokenLifeSpan = gokits.Condition(setDuration > MaxTokenLifeSpan,
            MaxTokenLifeSpan, setDuration).(time.Duration)
    })

    gokits.If("" != config.FengniaoAppProxyURL, func() {
        proxyURL = config.FengniaoAppProxyURL
    })

    golog.Infof("fengniao/app config: %+v", *config)
}
