package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

// 可配置联调地址: https://exam-anubis.ele.me/anubis-webapi/get_access_token
var fengniaoAppTokenURL = "https://open-anubis.ele.me/anubis-webapi/get_access_token"

var fengniaoAppConfigLifeSpan = time.Minute * 60   // config cache 60 min default
var fengniaoAppTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var fengniaoAppTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

// 可配置联调地址: https://exam-anubis.ele.me/anubis-webapi/v2/
const DefaultFengniaoAppProxyURL = "https://exam-anubis.ele.me/anubis-webapi/v2/"

var fengniaoAppProxyURL = DefaultFengniaoAppProxyURL

func fengniaoAppTokenLoad(config *Config) {
    gokits.If("" != config.FengniaoAppTokenURL, func() {
        fengniaoAppTokenURL = config.FengniaoAppTokenURL
    })

    gokits.If(0 != config.FengniaoAppConfigLifeSpan.Duration, func() {
        fengniaoAppConfigLifeSpan = config.FengniaoAppConfigLifeSpan.Duration
    })
    gokits.If(0 != config.FengniaoAppTokenLifeSpan.Duration, func() {
        fengniaoAppTokenLifeSpan = config.FengniaoAppTokenLifeSpan.Duration
    })
    gokits.If(0 != config.FengniaoAppTokenTempLifeSpan.Duration, func() {
        fengniaoAppTokenTempLifeSpan = config.FengniaoAppTokenTempLifeSpan.Duration
    })

    gokits.If("" != config.FengniaoAppProxyURL, func() {
        fengniaoAppProxyURL = config.FengniaoAppProxyURL
    })

    fengniaoAppTokenInitialize()
    fengniaoAppProxyInitialize()
}
