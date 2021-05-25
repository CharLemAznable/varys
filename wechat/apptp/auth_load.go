package apptp

import (
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "time"
)

var preAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="
var queryAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token="
var refreshAuthURL = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token="

var authTokenLifeSpan = time.Minute * 5     // stable token cache 5 min default
var authTokenTempLifeSpan = time.Minute * 1 // temporary token cache 1 min default

const DefaultAuthProxyURL = "https://api.weixin.qq.com/"
const DefaultAuthMpLoginProxyURL = "https://api.weixin.qq.com/sns/component/"

var authProxyURL = DefaultAuthProxyURL
var authMpLoginProxyURL = DefaultAuthMpLoginProxyURL

type AuthConfig struct {
    WechatTpPreAuthCodeURL        string
    WechatTpQueryAuthURL          string
    WechatTpRefreshAuthURL        string
    WechatTpAuthTokenLifeSpan     gokits.Duration
    WechatTpAuthTokenTempLifeSpan gokits.Duration
    WechatTpAuthProxyURL          string
    WechatTpAuthMpLoginProxyURL   string
}

var authConfig = &AuthConfig{}

func init() {
    base.RegisterLoader(func(configFile string) {
        base.LoadConfig(configFile, authConfig)
        fixAuthConfig()

        authCacheInitialize()
        authProxyInitialize()
        authMpLoginProxyInitialize()
    })
}

func fixAuthConfig() {
    gokits.If("" != authConfig.WechatTpPreAuthCodeURL, func() {
        preAuthCodeURL = authConfig.WechatTpPreAuthCodeURL
    })
    gokits.If("" != authConfig.WechatTpQueryAuthURL, func() {
        queryAuthURL = authConfig.WechatTpQueryAuthURL
    })
    gokits.If("" != authConfig.WechatTpRefreshAuthURL, func() {
        refreshAuthURL = authConfig.WechatTpRefreshAuthURL
    })

    gokits.If(0 != authConfig.WechatTpAuthTokenLifeSpan.Duration, func() {
        authTokenLifeSpan = authConfig.WechatTpAuthTokenLifeSpan.Duration
    })
    gokits.If(0 != authConfig.WechatTpAuthTokenTempLifeSpan.Duration, func() {
        authTokenTempLifeSpan = authConfig.WechatTpAuthTokenTempLifeSpan.Duration
    })

    gokits.If("" != authConfig.WechatTpAuthProxyURL, func() {
        authProxyURL = authConfig.WechatTpAuthProxyURL
    })
    gokits.If("" != authConfig.WechatTpAuthMpLoginProxyURL, func() {
        authMpLoginProxyURL = authConfig.WechatTpAuthMpLoginProxyURL
    })

    golog.Infof("wechat/apptp/auth config: %+v", *config)
}
