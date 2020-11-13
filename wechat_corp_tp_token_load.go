package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var wechatCorpTpTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_suite_token"

var wechatCorpTpConfigLifeSpan = time.Minute * 60         // config cache 60 min default
var wechatCorpTpCryptorLifeSpan = time.Minute * 60        // cryptor cache 60 min default
var wechatCorpTpTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpTpTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

func wechatCorpTpTokenLoad(config *Config) {
    gokits.If(0 != len(config.WechatCorpTpTokenURL), func() {
        wechatCorpTpTokenURL = config.WechatCorpTpTokenURL
    })

    gokits.If(0 != config.WechatCorpTpConfigLifeSpan.Duration, func() {
        wechatCorpTpConfigLifeSpan = config.WechatCorpTpConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpCryptorLifeSpan.Duration, func() {
        wechatCorpTpCryptorLifeSpan = config.WechatCorpTpCryptorLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpTokenMaxLifeSpan.Duration, func() {
        wechatCorpTpTokenMaxLifeSpan = config.WechatCorpTpTokenMaxLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpTokenExpireCriticalSpan.Duration, func() {
        wechatCorpTpTokenExpireCriticalSpan = config.WechatCorpTpTokenExpireCriticalSpan.Duration
    })

    wechatCorpTpTokenInitialize()
}
