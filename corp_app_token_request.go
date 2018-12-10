package varys

import (
    . "github.com/CharLemAznable/goutils"
    "github.com/CharLemAznable/httpreq"
    log "github.com/CharLemAznable/log4go"
    "time"
)

var wechatCorpTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"

type WechatCorpTokenResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpTokenRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatCorpTokenConfigCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    config := cache.Data().(*WechatCorpTokenConfig)

    result, err := httpreq.New(wechatCorpTokenURL).Params(
        "corpid", config.CorpId, "corpsecret", config.CorpSecret).
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    log.Trace("Request WechatCorpToken Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := UnJson(result, new(WechatCorpTokenResponse)).(*WechatCorpTokenResponse)
    if nil == response || 0 != response.Errcode || 0 == len(response.AccessToken) {
        return nil, &UnexpectedError{Message:
        "Request access_token Failed: " + result}
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "CORP_ID":      config.CorpId,
        "ACCESS_TOKEN": response.AccessToken,
        "EXPIRE_TIME":  StrFromInt64(expireTime)}, nil
}
