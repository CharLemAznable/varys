package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var wechatCorpTpPreAuthCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_pre_auth_code?suite_access_token="
var wechatCorpTpPermanentCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_permanent_code?suite_access_token="
var wechatCorpTpAuthTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_corp_token?suite_access_token="

var wechatCorpTpPermanentCodeLifeSpan = time.Minute * 60      // permanent_code cache 60 min default
var wechatCorpTpAuthTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpTpAuthTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

func wechatCorpTpAuthTokenLoad(config *Config) {
    gokits.If(0 != len(config.WechatCorpTpPreAuthCodeURL), func() {
        wechatCorpTpPreAuthCodeURL = config.WechatCorpTpPreAuthCodeURL
    })
    gokits.If(0 != len(config.WechatCorpTpPermanentCodeURL), func() {
        wechatCorpTpPermanentCodeURL = config.WechatCorpTpPermanentCodeURL
    })
    gokits.If(0 != len(config.WechatCorpTpAuthTokenURL), func() {
        wechatCorpTpAuthTokenURL = config.WechatCorpTpAuthTokenURL
    })

    gokits.If(0 != config.WechatCorpTpPermanentCodeLifeSpan.Duration, func() {
        wechatCorpTpPermanentCodeLifeSpan = config.WechatCorpTpPermanentCodeLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpAuthTokenMaxLifeSpan.Duration, func() {
        wechatCorpTpAuthTokenMaxLifeSpan = config.WechatCorpTpAuthTokenMaxLifeSpan.Duration
    })
    gokits.If(0 != config.WechatCorpTpAuthTokenExpireCriticalSpan.Duration, func() {
        wechatCorpTpAuthTokenExpireCriticalSpan = config.WechatCorpTpAuthTokenExpireCriticalSpan.Duration
    })

    wechatCorpTpAuthTokenInitialize()
}
