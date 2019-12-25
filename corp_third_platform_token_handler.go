package main

import (
    "encoding/xml"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/wechataes"
    "io/ioutil"
    "net/http"
    "strings"
)

type WechatCorpThirdPlatformAuthorizeData struct {
    XMLName     xml.Name `xml:"xml"`
    SuiteId     string   `xml:"SuiteId"`
    InfoType    string   `xml:"InfoType"`
    TimeStamp   string   `xml:"TimeStamp"`
    SuiteTicket string   `xml:"SuiteTicket"`
    AuthCode    string   `xml:"AuthCode"`
    AuthCorpId  string   `xml:"AuthCorpId"`
    EchoStr     string
}

func parseWechatCorpThirdPlatformAuthorizeData(codeName string, request *http.Request) (*WechatCorpThirdPlatformAuthorizeData, error) {
    cache, err := wechatCorpThirdPlatformCryptorCache.Value(codeName)
    if nil != err {
        _ = gokits.LOG.Warn("Load WechatCorpThirdPlatformCryptor Cache error:(%s) %s", codeName, err.Error())
        return nil, err
    }
    cryptor := cache.Data().(*wechataes.WechatCryptor)

    body, err := ioutil.ReadAll(request.Body)
    if nil != err {
        _ = gokits.LOG.Warn("Request read Body error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    err = request.ParseForm()
    if nil != err {
        _ = gokits.LOG.Warn("Request ParseForm error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    params := request.Form
    msgSign := params.Get("msg_signature")
    timeStamp := params.Get("timestamp")
    nonce := params.Get("nonce")
    echostr := params.Get("echostr")
    if 0 != len(echostr) { // 验证推送URL
        msg, err := cryptor.DecryptMsgContent(msgSign, timeStamp, nonce, echostr)
        if nil != err {
            _ = gokits.LOG.Warn("WechatCryptor DecryptMsg EchoStr error:(%s) %s", codeName, err.Error())
            return nil, err
        }
        gokits.LOG.Info("WechatCorpThirdPlatformVerifyEchoStrMsg:(%s) %s", codeName, msg)
        echoData := new(WechatCorpThirdPlatformAuthorizeData)
        echoData.InfoType = "echostr"
        echoData.EchoStr = msg
        return echoData, nil
    }

    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, string(body))
    if nil != err {
        _ = gokits.LOG.Warn("WechatCryptor DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    gokits.LOG.Info("WechatCorpThirdPlatformAuthorizeData:(%s) %s", codeName, decryptMsg)
    authorizeData := new(WechatCorpThirdPlatformAuthorizeData)
    err = xml.Unmarshal([]byte(decryptMsg), authorizeData)
    if nil != err {
        _ = gokits.LOG.Warn("Unmarshal DecryptMsg error:(%s) %s", codeName, err.Error())
        return nil, err
    }

    return authorizeData, nil
}

// /accept-authorization/{codeName:string}
const acceptCorpAuthorizationPath = "/accept-corp-authorization/"

func acceptCorpAuthorization(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, acceptCorpAuthorizationPath)
    if 0 != len(codeName) {
        authorizeData, err := parseWechatCorpThirdPlatformAuthorizeData(codeName, request)
        if nil == err {
            switch authorizeData.InfoType {

            case "echostr": // 验证推送URL
                gokits.ResponseText(writer, authorizeData.EchoStr)
                return

            case "suite_ticket":
                _, _ = updateWechatCorpThirdPlatformTicket(codeName, authorizeData.SuiteTicket)

            case "create_auth":
                go wechatCorpThirdPlatformAuthorizeCreator(codeName, authorizeData.AuthCode)

            case "change_auth":
                // ignore

            case "cancel_auth":
                authCorpId := authorizeData.AuthCorpId
                _, _ = disableWechatCorpThirdPlatformAuthorizer(codeName, authCorpId)
                // delete cache
                key := WechatCorpThirdPlatformAuthorizerKey{CodeName: codeName, CorpId: authCorpId}
                _, _ = wechatCorpThirdPlatformPermanentCodeCache.Delete(key)
                _, _ = wechatCorpThirdPlatformCorpTokenCache.Delete(key)

            }
        }
    }

    gokits.ResponseText(writer, "success")
}

type WechatCorpThirdPlatformPreAuthCodeResponse struct {
    Errcode     int    `json:"errcode"`
    Errmsg      string `json:"errmsg"`
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpThirdPlatformPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
    cache, err := wechatCorpThirdPlatformTokenCache.Value(codeName)
    if nil != err {
        return nil, err
    }
    tokenItem := cache.Data().(*WechatCorpThirdPlatformToken)

    result, err := gokits.NewHttpReq(wechatCorpThirdPlatformPreAuthCodeURL + tokenItem.AccessToken).Get()
    gokits.LOG.Trace("Request WechatCorpThirdPlatformPreAuthCode Response:(%s) %s", codeName, result)
    if nil != err {
        return nil, err
    }

    response := gokits.UnJson(result, new(WechatCorpThirdPlatformPreAuthCodeResponse)).(*WechatCorpThirdPlatformPreAuthCodeResponse)
    if nil == response || 0 == len(response.PreAuthCode) {
        return nil, &UnexpectedError{Message: "Request WechatCorpThirdPlatformPreAuthCode Failed: " + result}
    }
    return map[string]string{
        "SUITE_ID":      tokenItem.SuiteId,
        "PRE_AUTH_CODE": response.PreAuthCode,
        "EXPIRES_IN":    gokits.StrFromInt(response.ExpiresIn)}, nil
}

// /corp-authorize-component/{codeName:string}
const corpAuthorizeComponentPath = "/corp-authorize-component/"
const corpRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/corp-authorize-component/")) + "/corp-authorize-redirect/%s";
    location.replace(
        "https://open.work.weixin.qq.com/3rdapp/install?suite_id=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri) + "&state=%s"
    );
</script></head></html>
`

func corpAuthorizeComponent(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, corpAuthorizeComponentPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    response, err := wechatCorpThirdPlatformPreAuthCodeRequestor(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    suiteId := response["SUITE_ID"]
    preAuthCode := response["PRE_AUTH_CODE"]
    state := request.URL.Query().Get("state")

    gokits.ResponseHtml(writer, fmt.Sprintf(corpRedirectPageHtmlFormat, codeName, suiteId, preAuthCode, state))
}

// /corp-authorize-redirect/{codeName:string}
const corpAuthorizeRedirectPath = "/corp-authorize-redirect/"
const corpAuthorizedPageHtmlFormat = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style><script>var p="%s";0!=p.length&&location.replace(p);</script></head><body><div class="tips">授权成功</div></body></html>
`

func corpAuthorizeRedirect(writer http.ResponseWriter, request *http.Request) {
    codeName := trimPrefixPath(request, corpAuthorizeRedirectPath)
    if 0 == len(codeName) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
        return
    }

    redirectQuery := request.URL.RawQuery
    authCode := request.URL.Query().Get("auth_code")
    if 0 == len(authCode) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Corp Unauthorized"}))
        return
    }

    cache, err := wechatCorpThirdPlatformConfigCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Illegal"}))
        return
    }
    config := cache.Data().(*WechatCorpThirdPlatformConfig)
    redirectUrl := config.RedirectURL

    go wechatCorpThirdPlatformAuthorizeCreator(codeName, authCode)

    if 0 != len(redirectUrl) && 0 != len(redirectQuery) {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }

    gokits.ResponseHtml(writer, fmt.Sprintf(corpAuthorizedPageHtmlFormat, redirectUrl))
}

// /query-wechat-corp-authorizer-token/{codeName:string}/{corpId:string}
const queryWechatCorpAuthorizerTokenPath = "/query-wechat-corp-authorizer-token/"

func queryWechatCorpAuthorizerToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := trimPrefixPath(request, queryWechatCorpAuthorizerTokenPath)
    if 0 == len(pathParams) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/corpId"}))
        return
    }

    codeName := ids[0]
    corpId := ids[1]
    cache, err := wechatCorpThirdPlatformCorpTokenCache.Value(
        WechatCorpThirdPlatformAuthorizerKey{CodeName: codeName, CorpId: corpId})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := cache.Data().(*WechatCorpThirdPlatformCorpToken)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{
        "suiteId": token.SuiteId,
        "corpId":  token.CorpId,
        "token":   token.CorpAccessToken}))
}
