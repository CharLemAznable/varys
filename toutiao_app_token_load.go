package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var toutiaoAppTokenURL = "https://developer.toutiao.com/api/apps/token"

var toutiaoAppConfigLifeSpan = time.Minute * 60   // config cache 60 min default
var toutiaoAppTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var toutiaoAppTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

func toutiaoAppTokenLoad(config *Config) {
    gokits.If(0 != len(config.ToutiaoAppTokenURL), func() {
        toutiaoAppTokenURL = config.ToutiaoAppTokenURL
    })

    gokits.If(0 != config.ToutiaoAppConfigLifeSpan.Duration, func() {
        toutiaoAppConfigLifeSpan = config.ToutiaoAppConfigLifeSpan.Duration
    })
    gokits.If(0 != config.ToutiaoAppTokenLifeSpan.Duration, func() {
        toutiaoAppTokenLifeSpan = config.ToutiaoAppTokenLifeSpan.Duration
    })
    gokits.If(0 != config.ToutiaoAppTokenTempLifeSpan.Duration, func() {
        toutiaoAppTokenTempLifeSpan = config.ToutiaoAppTokenTempLifeSpan.Duration
    })

    toutiaoAppTokenInitialize()
}
