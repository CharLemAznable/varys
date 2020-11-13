package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var wechatTpTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

var wechatTpConfigLifeSpan = time.Minute * 60   // config cache 60 min default
var wechatTpCryptorLifeSpan = time.Minute * 60  // cryptor cache 60 min default
var wechatTpTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatTpTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

const DefaultWechatTpProxyURL = "https://api.weixin.qq.com/cgi-bin/"

var wechatTpProxyURL = DefaultWechatTpProxyURL

func wechatTpTokenLoad(config *Config) {
    gokits.If(0 != len(config.WechatTpTokenURL), func() {
        wechatTpTokenURL = config.WechatTpTokenURL
    })

    gokits.If(0 != config.WechatTpConfigLifeSpan.Duration, func() {
        wechatTpConfigLifeSpan = config.WechatTpConfigLifeSpan.Duration
    })
    gokits.If(0 != config.WechatTpCryptorLifeSpan.Duration, func() {
        wechatTpCryptorLifeSpan = config.WechatTpCryptorLifeSpan.Duration
    })
    gokits.If(0 != config.WechatTpTokenLifeSpan.Duration, func() {
        wechatTpTokenLifeSpan = config.WechatTpTokenLifeSpan.Duration
    })
    gokits.If(0 != config.WechatTpTokenTempLifeSpan.Duration, func() {
        wechatTpTokenTempLifeSpan = config.WechatTpTokenTempLifeSpan.Duration
    })

    gokits.If(0 != len(config.WechatTpProxyURL), func() {
        wechatTpProxyURL = config.WechatTpProxyURL
    })

    wechatTpTokenInitialize()
    wechatTpProxyInitialize()
}
