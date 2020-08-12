package main

import (
    "time"
)

var toutiaoAppTokenURL = "https://developer.toutiao.com/api/apps/token"

var toutiaoAppConfigLifeSpan = time.Minute * 60   // config cache 60 min default
var toutiaoAppTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var toutiaoAppTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

func toutiaoAppTokenLoad(configMap map[string]string) {
    urlConfigLoader(configMap["toutiaoAppTokenURL"],
        func(configURL string) {
            toutiaoAppTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["toutiaoAppConfigLifeSpan"],
        func(configVal time.Duration) {
            toutiaoAppConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["toutiaoAppTokenLifeSpan"],
        func(configVal time.Duration) {
            toutiaoAppTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["toutiaoAppTokenTempLifeSpan"],
        func(configVal time.Duration) {
            toutiaoAppTokenTempLifeSpan = configVal * time.Minute
        })

    toutiaoAppTokenInitialize()
}
