package main

import (
    "github.com/CharLemAznable/gokits"
    "time"
)

type WechatTpQueryAuthResponse struct {
    AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

type AuthorizationInfo struct {
    AuthorizerAppid        string     `json:"authorizer_appid"`
    AuthorizerAccessToken  string     `json:"authorizer_access_token"`
    ExpiresIn              int        `json:"expires_in"`
    AuthorizerRefreshToken string     `json:"authorizer_refresh_token"`
    FuncInfo               []FuncInfo `json:"func_info"`
}

type FuncInfo struct {
    FuncscopeCategory FuncscopeCategory `json:"funcscope_category"`
}

type FuncscopeCategory struct {
    Id int `json:"id"`
}

func wechatTpQueryAuthRequestor(codeName, authorizationCode interface{}) (map[string]string, error) {
    cache, err := wechatTpTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatTpToken)

    result, err := gokits.NewHttpReq(wechatTpQueryAuthURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":    tokenItem.AppId,
            "authorization_code": authorizationCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    gokits.LOG.Trace("Request WechatTpAuthToken Response:(%s, ) %s", codeName, authorizationCode, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpQueryAuthResponse)).(*WechatTpQueryAuthResponse)
    if nil == response || 0 == len(response.AuthorizationInfo.AuthorizerAccessToken) {
        return nil, &UnexpectedError{Message: "Request WechatTpAuthToken Failed: " + result}
    }
    return map[string]string{
        "APP_ID":                   tokenItem.AppId,
        "AUTHORIZER_APPID":         response.AuthorizationInfo.AuthorizerAppid,
        "AUTHORIZER_ACCESS_TOKEN":  response.AuthorizationInfo.AuthorizerAccessToken,
        "AUTHORIZER_REFRESH_TOKEN": response.AuthorizationInfo.AuthorizerRefreshToken,
        "EXPIRES_IN":               gokits.StrFromInt(response.AuthorizationInfo.ExpiresIn)}, nil
}

func wechatTpAuthTokenCreator(codeName, authorizerAppId, authorizationCode interface{}) {
    count, err := db.New().Sql(createWechatTpAuthTokenSQL).
        Params(codeName, authorizerAppId).Execute()
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        // 强制更新记录, 不论token是否已过期
        count, err = db.New().Sql(updateWechatTpAuthTokenForceSQL).
            Params(codeName, authorizerAppId).Execute()
        for nil != err || count < 1 { // 尝试更新记录失败, 则休眠(1 sec)后再次尝试更新记录
            time.Sleep(1 * time.Second)
            count, err = db.New().Sql(updateWechatTpAuthTokenForceSQL).
                Params(codeName, authorizerAppId).Execute()
        }
    }

    // 锁定成功, 开始更新
    resultItem, err := wechatTpQueryAuthRequestor(codeName, authorizationCode)
    if nil != err {
        _, _ = db.New().Sql(uncompleteWechatTpAuthTokenSQL).
            Params(codeName, authorizerAppId).Execute()
        _ = gokits.LOG.Warn("Request WechatTpAuthToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, err.Error())
        return
    }
    count, err = db.New().Sql(completeWechatTpAuthTokenSQL).
        Params(wechatTpAuthTokenCompleteParamBuilder(
            resultItem, wechatTpAuthTokenLifeSpan,
            codeName, authorizerAppId)...).Execute()
    if nil != err || count < 1 {
        _ = gokits.LOG.Warn("Record new WechatTpAuthToken Failed:(%s, %s) %s",
            codeName, authorizerAppId, gokits.DefaultIfNil(err, &UnexpectedError{Message: "Replace WechatTpAuthToken Failed"}).(error).Error())
        return
    }

    tokenItem := wechatTpAuthTokenBuilder(resultItem)
    gokits.LOG.Info("Request WechatTpAuthToken:(%s, %s) %s",
        codeName, authorizerAppId, gokits.Json(tokenItem))
    wechatTpAuthTokenCache.Add(WechatTpAuthKey{
        CodeName: codeName.(string), AuthorizerAppId: authorizerAppId.(string)},
        wechatTpAuthTokenLifeSpan, tokenItem)
}

func wechatTpAuthorized(codeName string, infoData *WechatTpInfoData) {
    AuthorizerAppid := infoData.AuthorizerAppid
    AuthorizationCode := infoData.AuthorizationCode
    _, _ = enableWechatTpAuth(codeName, AuthorizerAppid, AuthorizationCode, infoData.PreAuthCode)
    go wechatTpAuthTokenCreator(codeName, AuthorizerAppid, AuthorizationCode)
}

func wechatTpUnauthorized(codeName string, infoData *WechatTpInfoData) {
    AuthorizerAppid := infoData.AuthorizerAppid
    _, _ = disableWechatTpAuth(codeName, AuthorizerAppid)
    // delete cache
    _, _ = wechatTpAuthTokenCache.Delete(
        WechatTpAuthKey{CodeName: codeName, AuthorizerAppId: AuthorizerAppid})
}
