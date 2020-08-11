package main

import (
    "time"
)

var wechatCorpTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"

var wechatCorpConfigLifeSpan = time.Minute * 60         // config cache 60 min default
var wechatCorpTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

const DefaultWechatCorpProxyURL = "https://qyapi.weixin.qq.com/cgi-bin/"

var wechatCorpProxyURL = DefaultWechatCorpProxyURL

func wechatCorpTokenLoad(configMap map[string]string) {
    urlConfigLoader(configMap["wechatCorpTokenURL"],
        func(configURL string) {
            wechatCorpTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatCorpConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpTokenMaxLifeSpan"],
        func(configVal time.Duration) {
            wechatCorpTokenMaxLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatCorpTokenExpireCriticalSpan"],
        func(configVal time.Duration) {
            wechatCorpTokenExpireCriticalSpan = configVal * time.Second
        })

    urlConfigLoader(configMap["wechatCorpProxyURL"],
        func(configURL string) {
            wechatCorpProxyURL = configURL
        })

    wechatCorpTokenInitialize()
    wechatCorpProxyInitialize()
}
