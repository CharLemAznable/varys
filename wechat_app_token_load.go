package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var wechatAppTokenURL = "https://api.weixin.qq.com/cgi-bin/token"

var wechatAppConfigLifeSpan = time.Minute * 60   // config cache 60 min default
var wechatAppTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatAppTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

const DefaultWechatAppProxyURL = "https://api.weixin.qq.com/cgi-bin/"
const DefaultWechatMpProxyURL = "https://api.weixin.qq.com/wxa/"
const DefaultWechatMpLoginProxyURL = "https://api.weixin.qq.com/sns/"

var wechatAppProxyURL = DefaultWechatAppProxyURL
var wechatMpProxyURL = DefaultWechatMpProxyURL
var wechatMpLoginProxyURL = DefaultWechatMpLoginProxyURL

func wechatAppTokenLoad(config *Config) {
    gokits.If(0 != len(config.WechatAppTokenURL), func() {
        wechatAppTokenURL = config.WechatAppTokenURL
    })

    gokits.If(0 != config.WechatAppConfigLifeSpan.Duration, func() {
        wechatAppConfigLifeSpan = config.WechatAppConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatAppTokenLifeSpan.Duration, func() {
        wechatAppTokenLifeSpan = config.WechatAppTokenLifeSpan.Duration
    })
    gokits.If(0 != config.WechatAppTokenTempLifeSpan.Duration, func() {
        wechatAppTokenTempLifeSpan = config.WechatAppTokenTempLifeSpan.Duration
    })

    gokits.If(0 != len(config.WechatAppProxyURL), func() {
        wechatAppProxyURL = config.WechatAppProxyURL
    })
    gokits.If(0 != len(config.WechatMpProxyURL), func() {
        wechatMpProxyURL = config.WechatMpProxyURL
    })
    gokits.If(0 != len(config.WechatMpLoginProxyURL), func() {
        wechatMpLoginProxyURL = config.WechatMpLoginProxyURL
    })

    wechatAppTokenInitialize()
    wechatAppProxyInitialize()
    wechatMpProxyInitialize()
    wechatMpLoginProxyInitialize()
}
