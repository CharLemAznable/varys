package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

var wechatTpPreAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="
var wechatTpQueryAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token="
var wechatTpRefreshAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="

var wechatTpAuthTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatTpAuthTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

func wechatTpAuthTokenLoad(config *Config) {
    gokits.If(0 != len(config.WechatTpPreAuthCodeURL), func() {
        wechatTpPreAuthCodeURL = config.WechatTpPreAuthCodeURL
    })
    gokits.If(0 != len(config.WechatTpQueryAuthURL), func() {
        wechatTpQueryAuthURL = config.WechatTpQueryAuthURL
    })
    gokits.If(0 != len(config.WechatTpRefreshAuthURL), func() {
        wechatTpRefreshAuthURL = config.WechatTpRefreshAuthURL
    })

    gokits.If(0 != config.WechatTpAuthTokenLifeSpan.Duration, func() {
        wechatTpAuthTokenLifeSpan = config.WechatTpAuthTokenLifeSpan.Duration
    })
    gokits.If(0 != config.WechatTpAuthTokenTempLifeSpan.Duration, func() {
        wechatTpAuthTokenTempLifeSpan = config.WechatTpAuthTokenTempLifeSpan.Duration
    })

    wechatTpAuthTokenInitialize()
}