package varys

import (
    "time"
)

var wechatAppThirdPlatformTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"
var wechatAppThirdPlatformPreAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="
var wechatAppThirdPlatformQueryAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token="
var wechatAppThirdPlatformRefreshAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="

var wechatAppThirdPlatformConfigLifeSpan = time.Minute * 60             // config cache 60 min default
var wechatAppThirdPlatformCryptorLifeSpan = time.Minute * 60            // cryptor cache 60 min default
var wechatAppThirdPlatformTokenLifeSpan = time.Minute * 5               // stable component token cache 5 min default
var wechatAppThirdPlatformTokenTempLifeSpan = time.Minute * 1           // temporary component token cache 1 min default
var wechatAppThirdPlatformAuthorizerTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var wechatAppThirdPlatformAuthorizerTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

func wechatAppThirdPlatformAuthorizerTokenLoad(configMap map[string]string) {
    urlConfigLoader(configMap["wechatAppThirdPlatformTokenURL"],
        func(configURL string) {
            wechatAppThirdPlatformTokenURL = configURL
        })
    urlConfigLoader(configMap["wechatAppThirdPlatformPreAuthCodeURL"],
        func(configURL string) {
            wechatAppThirdPlatformPreAuthCodeURL = configURL
        })
    urlConfigLoader(configMap["wechatAppThirdPlatformQueryAuthURL"],
        func(configURL string) {
            wechatAppThirdPlatformQueryAuthURL = configURL
        })
    urlConfigLoader(configMap["wechatAppThirdPlatformRefreshAuthURL"],
        func(configURL string) {
            wechatAppThirdPlatformRefreshAuthURL = configURL
        })

    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformConfigLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformConfigLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformCryptorLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformCryptorLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformTokenTempLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformAuthorizerTokenLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformAuthorizerTokenLifeSpan = configVal * time.Minute
        })
    lifeSpanConfigLoader(
        configMap["wechatAppThirdPlatformAuthorizerTokenTempLifeSpan"],
        func(configVal time.Duration) {
            wechatAppThirdPlatformAuthorizerTokenTempLifeSpan = configVal * time.Minute
        })

    wechatAppThirdPlatformAuthorizerTokenInitialize()
}
