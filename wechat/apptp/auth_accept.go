package apptp

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/varys/base"
    "github.com/CharLemAznable/varys/wechat/jsapi"
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
    cache, err := tokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatTpToken)

    result, err := gokits.NewHttpReq(queryAuthURL + tokenItem.AccessToken).
        RequestBody(gokits.Json(map[string]string{
            "component_appid":    tokenItem.AppId,
            "authorization_code": authorizationCode.(string)})).
        Prop("Content-Type", "application/json").Post()
    golog.Debugf("Request Wechat Tp Auth Token Response:(%s, %s) %s", codeName, authorizationCode, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatTpQueryAuthResponse)).(*WechatTpQueryAuthResponse)
    if nil == response || "" == response.AuthorizationInfo.AuthorizerAccessToken {
        return nil, errors.New("Request Wechat Tp Auth Token Failed: " + result)
    }

    jsapiTicket := jsapi.TicketRequestor(codeName.(string),
        response.AuthorizationInfo.AuthorizerAccessToken)

    return map[string]string{
        "AppId":                  tokenItem.AppId,
        "AuthorizerAppId":        response.AuthorizationInfo.AuthorizerAppId,
        "AuthorizerAccessToken":  response.AuthorizationInfo.AuthorizerAccessToken,
        "AuthorizerRefreshToken": response.AuthorizationInfo.AuthorizerRefreshToken,
        "AuthorizerJsapiTicket":  jsapiTicket,
        "ExpiresIn":              gokits.StrFromInt(response.AuthorizationInfo.ExpiresIn)}, nil
}

func wechatTpAuthCompleteArg(response map[string]string, lifeSpan time.Duration) map[string]interface{} {
    expiresIn, _ := gokits.IntFromStr(response["ExpiresIn"])
    return map[string]interface{}{"AuthorizerAppId": response["AuthorizerAppId"],
        "AuthorizerAccessToken":  response["AuthorizerAccessToken"],
        "AuthorizerRefreshToken": response["AuthorizerRefreshToken"],
        "AuthorizerJsapiTicket":  response["AuthorizerJsapiTicket"],
        // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
        "ExpiresIn": expiresIn - int(authTokenLifeSpan.Seconds()*1.1)}
}

func wechatTpAuthTokenBuilder(response map[string]string) interface{} {
    return &WechatTpAuthToken{AppId: response["AppId"],
        AuthorizerAppId:       response["AuthorizerAppId"],
        AuthorizerAccessToken: response["AuthorizerAccessToken"],
        AuthorizerJsapiTicket: response["AuthorizerJsapiTicket"]}
}

func wechatTpAuthTokenCreator(codeName, authorizerAppId, authorizationCode interface{}) {
    _, err := base.DB.NamedExec(createAuthTokenSQL,
        map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
    if nil != err { // 尝试插入记录失败, 则尝试更新记录
        // 强制更新记录, 不论token是否已过期
        count, err := base.DB.NamedExecX(updateAuthTokenForceSQL,
            map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
        for nil != err || count < 1 { // 尝试更新记录失败, 则休眠(1 sec)后再次尝试更新记录
            time.Sleep(1 * time.Second)
            count, err = base.DB.NamedExecX(updateAuthTokenForceSQL,
                map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
        }
    }

    // 锁定成功, 开始更新
    response, err := wechatTpQueryAuthRequestor(codeName, authorizationCode)
    if nil != err {
        _, _ = base.DB.NamedExec(uncompleteAuthTokenSQL,
            map[string]interface{}{"CodeName": codeName, "AuthorizerAppId": authorizerAppId})
        golog.Warnf("Request Wechat Tp Auth Token Failed:(%s, %s) %s", codeName, authorizerAppId, err.Error())
        return
    }
    completeArg := wechatTpAuthCompleteArg(response, authTokenLifeSpan)
    completeArg["CodeName"] = codeName
    _, err = base.DB.NamedExec(completeAuthTokenSQL, completeArg)
    if nil != err {
        golog.Warnf("Record new Wechat Tp Auth Token Failed:(%s, %s) %s", codeName, authorizerAppId, err.Error())
        return
    }

    token := wechatTpAuthTokenBuilder(response)
    golog.Infof("Request Wechat Tp Auth Token:(%s, %s) %+v", codeName, authorizerAppId, token)
    authTokenCache.Add(WechatTpAuthKey{
        CodeName: codeName.(string), AuthorizerAppId: authorizerAppId.(string)},
        authTokenLifeSpan, token)
}

func wechatTpAuthorized(codeName string, infoData *WechatTpInfoData) {
    authorizerAppId := infoData.AuthorizerAppId
    authorizationCode := infoData.AuthorizationCode
    _, _ = base.DB.NamedExec(enableAuthSQL,
        map[string]interface{}{"CodeName": codeName,
            "AuthorizerAppId":   authorizerAppId,
            "AuthorizationCode": authorizationCode,
            "PreAuthCode":       infoData.PreAuthCode})
    wechatTpAuthTokenCreator(codeName, authorizerAppId, authorizationCode)
}

func wechatTpUnauthorized(codeName string, infoData *WechatTpInfoData) {
    authorizerAppId := infoData.AuthorizerAppId
    _, _ = base.DB.NamedExec(disableAuthSQL,
        map[string]interface{}{"CodeName": codeName,
            "AuthorizerAppId": authorizerAppId})
    // delete cache, publish to cluster nodes
    base.PublishToClusterNodes(func(address string) {
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
    _, _ = base.DB.NamedExec(enableAuthSQL,
        map[string]interface{}{"CodeName": codeName,
            "AuthorizerAppId":   mpAppId,
            "AuthorizationCode": mpAuthCode,
            "PreAuthCode":       ""})
    wechatTpAuthTokenCreator(codeName, mpAppId, mpAuthCode)
}
