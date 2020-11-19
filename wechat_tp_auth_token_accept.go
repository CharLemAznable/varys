package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "time"
)

type WechatTpQueryAuthResponse struct {
    AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

type AuthorizationInfo struct {
    AuthorizerAppId        string     `json:"authorizer_appid"`
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
    golog.Debugf("Request WechatTpAuthToken Response:(%s, %s) %s", codeName, authorizationCode, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpQueryAuthResponse)).(*WechatTpQueryAuthResponse)
    if nil == response || "" == response.AuthorizationInfo.AuthorizerAccessToken {
        return nil, errors.New("Request WechatTpAuthToken Failed: " + result)
    }
    return map[string]string{
        "AppId":                  tokenItem.AppId,
        "AuthorizerAppId":        response.AuthorizationInfo.AuthorizerAppId,
        "AuthorizerAccessToken":  response.AuthorizationInfo.AuthorizerAccessToken,
        "AuthorizerRefreshToken": response.AuthorizationInfo.AuthorizerRefreshToken,
        "ExpiresIn":              gokits.StrFromInt(response.AuthorizationInfo.ExpiresIn)}, nil
}

func wechatTpAuthCompleteArg(response map[string]string, lifeSpan time.Duration) map[string]interface{} {
    expiresIn, _ := gokits.IntFromStr(response["ExpiresIn"])
    return map[string]interface{}{"AuthorizerAppId": response["AuthorizerAppId"],
        "AuthorizerAccessToken":  response["AuthorizerAccessToken"],
        "AuthorizerRefreshToken": response["AuthorizerRefreshToken"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        "ExpiresIn": expiresIn - int(wechatTpAuthTokenLifeSpan.Seconds()*1.1)}
}

func wechatTpAuthTokenBuilder(response map[string]string) interface{} {
    return &WechatTpAuthToken{AppId: response["AppId"],
        AuthorizerAppId:       response["AuthorizerAppId"],
        AuthorizerAccessToken: response["AuthorizerAccessToken"]}
}

func wechatTpAuthTokenCreator(codeName, authorizerAppId, authorizationCode interface{}) {
    _, err := db.NamedExec(createWechatTpAuthTokenSQL,
        map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        // 强制更新记录, 不论token是否已过期
        count, err := db.NamedExecX(updateWechatTpAuthTokenForceSQL,
            map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
        for nil != err || count < 1 { // 尝试更新记录失败, 则休眠(1 sec)后再次尝试更新记录
            time.Sleep(1 * time.Second)
            count, err = db.NamedExecX(updateWechatTpAuthTokenForceSQL,
                map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
        }
    }

    // 锁定成功, 开始更新
    response, err := wechatTpQueryAuthRequestor(codeName, authorizationCode)
    if nil != err {
        _, _ = db.NamedExec(uncompleteWechatTpAuthTokenSQL,
            map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
        golog.Warnf("Request WechatTpAuthToken Failed:(%s, %s) %s", codeName, authorizerAppId, err.Error())
        return
    }
    completeArg := wechatTpAuthCompleteArg(response, wechatTpAuthTokenLifeSpan)
    completeArg["CodeName"] = codeName
    _, err = db.NamedExec(completeWechatTpAuthTokenSQL, completeArg)
    if nil != err {
        golog.Warnf("Record new WechatTpAuthToken Failed:(%s, %s) %s", codeName, authorizerAppId, err.Error())
        return
    }

    token := wechatTpAuthTokenBuilder(response)
    golog.Infof("Request WechatTpAuthToken:(%s, %s) %+v", codeName, authorizerAppId, token)
    wechatTpAuthTokenCache.Add(WechatTpAuthKey{
        CodeName: codeName.(string), AuthorizerAppId: authorizerAppId.(string)},
        wechatTpAuthTokenLifeSpan, token)
}

func wechatTpAuthorized(codeName string, infoData *WechatTpInfoData) {
    authorizerAppId := infoData.AuthorizerAppId
    authorizationCode := infoData.AuthorizationCode
    _, _ = db.NamedExec(enableWechatTpAuthSQL,
        map[string]interface{}{"CodeName": codeName,
            "AuthorizerAppId":   authorizerAppId,
            "AuthorizationCode": authorizationCode,
            "PreAuthCode":       infoData.PreAuthCode})
    wechatTpAuthTokenCreator(codeName, authorizerAppId, authorizationCode)
}

func wechatTpUnauthorized(codeName string, infoData *WechatTpInfoData) {
    authorizerAppId := infoData.AuthorizerAppId
    _, _ = db.NamedExec(disableWechatTpAuthSQL,
        map[string]interface{}{"CodeName": codeName,
            "AuthorizerAppId": authorizerAppId})
    // delete cache, publish to cluster nodes
    publishToClusterNodes(func(address string) {
        rsp, err := gokits.NewHttpReq(address + gokits.PathJoin(
            cleanWechatTpAuthTokenPath, codeName, authorizerAppId)).Get()
        if nil != err {
            golog.Errorf("Publish to %s Error: %s", address, err.Error())
        }
        golog.Debugf("Publish to %s Response: %s", address, rsp)
    })
}

func wechatTpAuthorizedMp(codeName string, infoData *WechatTpInfoData) {
    mpAppId := infoData.MpAppId
    mpAuthCode := infoData.MpAuthCode
    _, _ = db.NamedExec(enableWechatTpAuthSQL,
        map[string]interface{}{"CodeName": codeName,
            "AuthorizerAppId":   mpAppId,
            "AuthorizationCode": mpAuthCode,
            "PreAuthCode":       ""})
    wechatTpAuthTokenCreator(codeName, mpAppId, mpAuthCode)
}
