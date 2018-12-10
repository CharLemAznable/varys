package varys

import (
    . "github.com/CharLemAznable/goutils"
    "github.com/CharLemAznable/httpreq"
    log "github.com/CharLemAznable/log4go"
)

var wechatAPITokenURL = "https://api.weixin.qq.com/cgi-bin/token"

type WechatAPITokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatAPITokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatAPITokenConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatAPITokenConfig)

    result, err := httpreq.New(wechatAPITokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    log.Trace("Request WechatAPIToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatAPITokenResponse)).(*WechatAPITokenResponse)
    if nil == response || 0 == len(response.AccessToken) {
        return nil, &UnexpectedError{Message:
        "Request access_token Failed: " + result}
    }
    return map[string]string{
        "APP_ID":       config.AppId,
        "ACCESS_TOKEN": response.AccessToken,
        "EXPIRES_IN":   StrFromInt(response.ExpiresIn)}, nil
}
