package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "time"
)

type WechatCorpTpPermanentCodeResponse struct {
    Errcode       int          `json:"errcode"`
    Errmsg        string       `json:"errmsg"`
    AccessToken   string       `json:"access_token"`
    ExpiresIn     int          `json:"expires_in"`
    PermanentCode string       `json:"permanent_code"`
    AuthCorpInfo  AuthCorpInfo `json:"auth_corp_info"`
    AuthInfo      AuthInfo     `json:"auth_info"`
    AuthUserInfo  AuthUserInfo `json:"auth_user_info"`
}

type AuthCorpInfo struct {
    Corpid            string `json:"corpid"`
    CorpName          string `json:"corp_name"`
    CorpType          string `json:"corp_type"`
    CorpSquareLogoUrl string `json:"corp_square_logo_url"`
    CorpUserMax       int    `json:"corp_user_max"`
    CorpAgentMax      int    `json:"corp_agent_max"`
    CorpFullName      string `json:"corp_full_name"`
    VerifiedEndTime   int64  `json:"verified_end_time"`
    SubjectType       int    `json:"subject_type"`
    CorpWxqrcode      string `json:"corp_wxqrcode"`
    CorpScale         string `json:"corp_scale"`
    CorpIndustry      string `json:"corp_industry"`
    CorpSubIndustry   string `json:"corp_sub_industry"`
    Location          string `json:"location"`
}

type AuthInfo struct {
    Agent []Agent `json:"agent"`
}

type Agent struct {
    Agentid       int64     `json:"agentid"`
    Name          string    `json:"name"`
    RoundLogoUrl  string    `json:"round_logo_url"`
    SquareLogoUrl string    `json:"square_logo_url"`
    Appid         int64     `json:"appid"`
    Privilege     Privilege `json:"privilege"`
}

type Privilege struct {
    Level      int      `json:"level"`
    AllowParty []int    `json:"allow_party"`
    AllowUser  []string `json:"allow_user"`
    AllowTag   []int    `json:"allow_tag"`
    ExtraParty []int    `json:"extra_party"`
    ExtraUser  []string `json:"extra_user"`
    ExtraTag   []int    `json:"extra_tag"`
}

type AuthUserInfo struct {
    Userid string `json:"userid"`
    Name   string `json:"name"`
    Avatar string `json:"avatar"`
}

func wechatCorpTpPermanentCodeRequestor(codeName, authCode interface{}) (map[string]string, error) {
    cache, err := wechatCorpTpTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatCorpTpToken)

    result, err := gokits.NewHttpReq(wechatCorpTpPermanentCodeURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{"auth_code": authCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request WechatCorpTpPermanentCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpTpPermanentCodeResponse)).(*WechatCorpTpPermanentCodeResponse)
    if nil == response || "" == response.PermanentCode {
        return nil, errors.New("Request WechatCorpTpPermanentCode Failed: " + result)
    }

    // 过期时间增量: token实际有效时长
    expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
    return map[string]string{
        "SuiteId":       tokenItem.SuiteId,
        "CorpId":        response.AuthCorpInfo.Corpid,
        "PermanentCode": response.PermanentCode,
        "AccessToken":   response.AccessToken,
        "ExpireTime":    gokits.StrFromInt64(expireTime)}, nil
}

func wechatCorpTpAuthCreator(codeName, authCode interface{}) {
    response, err := wechatCorpTpPermanentCodeRequestor(codeName, authCode)
    if nil != err {
        golog.Warnf("Request WechatCorpTpPermanentCode Failed:(%s, authCode:%s) %s", codeName, authCode, err.Error())
        return
    }

    corpId := response["CorpId"]
    permanentCode := response["PermanentCode"]
    _, _ = db.NamedExec(enableWechatCorpTpAuthSQL, map[string]interface{}{
        "CodeName": codeName, "CorpId": corpId, "PermanentCode": permanentCode})

    accessToken := response["AccessToken"]
    expireTime := response["ExpireTime"]
    arg := map[string]interface{}{"CodeName": codeName,
        "CorpId": corpId, "AccessToken": accessToken, "ExpireTime": expireTime}
    _, err = db.NamedExec(createWechatCorpTpAuthTokenSQL, arg)
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        golog.Warnf("Create WechatCorpTpAuthToken Failed:(%s, corpId:%s) %s", codeName, corpId, err.Error())
        _, _ = db.NamedExec(updateWechatCorpTpAuthTokenSQL, arg)
        // 忽略更新记录的结果
        // 如果当前存在有效期内的token, 则token不会被更新, 重复请求微信也会返回同样的token
    }
}

func wechatCorpTpCreateAuth(codeName string, infoData *WechatCorpTpInfoData) {
    go wechatCorpTpAuthCreator(codeName, infoData.AuthCode)
}

func wechatCorpTpCancelAuth(codeName string, infoData *WechatCorpTpInfoData) {
    authCorpId := infoData.AuthCorpId
    _, _ = db.NamedExec(disableWechatCorpTpAuthSQL,
        map[string]interface{}{"CodeName": codeName, "CorpId": authCorpId})

    // delete cache
    key := WechatCorpTpAuthKey{CodeName: codeName, CorpId: authCorpId}
    _, _ = wechatCorpTpPermanentCodeCache.Delete(key)
    _, _ = wechatCorpTpAuthTokenCache.Delete(key)
}
