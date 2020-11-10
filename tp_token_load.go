package main

import (
    "time"
)

var wechatTpTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

var wechatTpConfigLifeSpan = time.Minute * 60   // config cache 60 min default
var wechatTpCryptorLifeSpan = time.Minute * 60  // cryptor cache 60 min default
var wechatTpTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatTpTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

const DefaultWechatTpProxyURL = "https://api.weixin.qq.com/cgi-bin/"

var wechatTpProxyURL = DefaultWechatTpProxyURL

func wechatTpTokenLoad(configMap map[string]string) {
    urlConfigLoader(configMap["wechatTpTokenURL"],
        func(configURL string) {
            wechatTpTokenURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatTpConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatTpConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatTpTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatTpTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatTpTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatTpTokenTempLifeSpan = configVal * time.Minute
        })

    urlConfigLoader(configMap["wechatTpProxyURL"],
        func(configURL string) {
            wechatTpProxyURL = configURL
        })

    wechatTpTokenInitialize()
    wechatTpProxyInitialize()
}
