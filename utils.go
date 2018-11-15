package varys

import (
    "encoding/json"
    "github.com/CharLemAznable/httpreq"
)

var wechatAPITokenURL = "https://api.weixin.qq.com/cgi-bin/token"

type WechatAPITokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn int `json:"expires_in"`
}

func requestWechatAPIToken(appId string) (*WechatAPITokenResponse, error) {
    config, err := GetWechatAPITokenConfig(appId)
    if nil != err {
        return nil, err
    }
    result, err := httpreq.New(wechatAPITokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    if nil != err {
        return nil, err
    }

    response := new(WechatAPITokenResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return nil, err
    }
    return response, nil
}
