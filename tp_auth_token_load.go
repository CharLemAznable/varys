package main

import (
    "time"
)

var wechatTpPreAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="
var wechatTpQueryAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token="
var wechatTpRefreshAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="

var wechatTpAuthTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatTpAuthTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

func wechatTpAuthTokenLoad(configMap map[string]string) {
    urlConfigLoader(configMap["wechatTpPreAuthCodeURL"],
        func(configURL string) {
            wechatTpPreAuthCodeURL = configURL
        })
    urlConfigLoader(configMap["wechatTpQueryAuthURL"],
        func(configURL string) {
            wechatTpQueryAuthURL = configURL
        })
    urlConfigLoader(configMap["wechatTpRefreshAuthURL"],
        func(configURL string) {
            wechatTpRefreshAuthURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatTpAuthTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatTpAuthTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatTpAuthTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatTpAuthTokenTempLifeSpan = configVal * time.Minute
        })

    wechatTpAuthTokenInitialize()
}
