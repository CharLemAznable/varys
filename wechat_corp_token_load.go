package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var wechatCorpTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"

var wechatCorpConfigLifeSpan = time.Minute * 60         // config cache 60 min default
var wechatCorpTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

const DefaultWechatCorpProxyURL = "https://qyapi.weixin.qq.com/cgi-bin/"

var wechatCorpProxyURL = DefaultWechatCorpProxyURL

func wechatCorpTokenLoad(config *Config) {
    gokits.If(0 != len(config.WechatCorpTokenURL), func() {
        wechatCorpTokenURL = config.WechatCorpTokenURL
    })

    gokits.If(0 != config.WechatCorpConfigLifeSpan.Duration, func() {
        wechatCorpConfigLifeSpan = config.WechatCorpConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTokenMaxLifeSpan.Duration, func() {
        wechatCorpTokenMaxLifeSpan = config.WechatCorpTokenMaxLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTokenExpireCriticalSpan.Duration, func() {
        wechatCorpTokenExpireCriticalSpan = config.WechatCorpTokenExpireCriticalSpan.Duration
    })

    gokits.If(0 != len(config.WechatCorpProxyURL), func() {
        wechatCorpProxyURL = config.WechatCorpProxyURL
    })

    wechatCorpTokenInitialize()
    wechatCorpProxyInitialize()
}
