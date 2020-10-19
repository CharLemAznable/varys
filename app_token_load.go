package main

import (
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

func wechatAppTokenLoad(configMap map[string]string) {
    urlConfigLoader(configMap["wechatAppTokenURL"],
        func(configURL string) {
            wechatAppTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatAppConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatAppConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAppTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAppTokenTempLifeSpan = configVal * time.Minute
        })

    urlConfigLoader(configMap["wechatAppProxyURL"],
        func(configURL string) {
            wechatAppProxyURL = configURL
        })
    urlConfigLoader(configMap["wechatMpProxyURL"],
        func(configURL string) {
            wechatMpProxyURL = configURL
        })
    urlConfigLoader(configMap["wechatMpLoginProxyURL"],
        func(configURL string) {
            wechatMpLoginProxyURL = configURL
        })

    wechatAppTokenInitialize()
    wechatAppProxyInitialize()
    wechatMpProxyInitialize()
    wechatMpLoginProxyInitialize()
}
