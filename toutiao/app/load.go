package app

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var tokenURL = "https://developer.toutiao.com/api/apps/token"

var configLifeSpan = time.Minute * 60   // config cache 60 min default
var tokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var tokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

type Config struct {
    ToutiaoAppTokenURL          string
    ToutiaoAppConfigLifeSpan    gokits.Duration
    ToutiaoAppTokenLifeSpan     gokits.Duration
    ToutiaoAppTokenTempLifeSpan gokits.Duration
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
    gokits.If("" != config.ToutiaoAppTokenURL, func() {
        tokenURL = config.ToutiaoAppTokenURL
    })

    gokits.If(0 != config.ToutiaoAppConfigLifeSpan.Duration, func() {
        configLifeSpan = config.ToutiaoAppConfigLifeSpan.Duration
    })
    gokits.If(0 != config.ToutiaoAppTokenLifeSpan.Duration, func() {
        tokenLifeSpan = config.ToutiaoAppTokenLifeSpan.Duration
    })
    gokits.If(0 != config.ToutiaoAppTokenTempLifeSpan.Duration, func() {
        tokenTempLifeSpan = config.ToutiaoAppTokenTempLifeSpan.Duration
    })

    golog.Infof("toutiao/app config: %+v", *config)
}
