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

type WechatAppThirdPlatformAuthorizeData struct {
	XMLName                      xml.Name `xml:"xml"`
	AppId                        string   `xml:"AppId"`
	CreateTime                   string   `xml:"CreateTime"`
	InfoType                     string   `xml:"InfoType"`
	ComponentVerifyTicket        string   `xml:"ComponentVerifyTicket"`
	AuthorizerAppid              string   `xml:"AuthorizerAppid"`
	AuthorizationCode            string   `xml:"AuthorizationCode"`
	AuthorizationCodeExpiredTime string   `xml:"AuthorizationCodeExpiredTime"`
	PreAuthCode                  string   `xml:"PreAuthCode"`
}

func parseWechatAppThirdPlatformAuthorizeData(codeName string, request *http.Request) (*WechatAppThirdPlatformAuthorizeData, error) {
	cache, err := wechatAppThirdPlatformCryptorCache.Value(codeName)
	if nil != err {
		_ = gokits.LOG.Warn("Load WechatAppThirdPlatformCryptor Cache error:(%s) %s", codeName, err.Error())
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
	decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, string(body))
	if nil != err {
		_ = gokits.LOG.Warn("WechatCryptor DecryptMsg error:(%s) %s", codeName, err.Error())
		return nil, err
	}

	gokits.LOG.Info("WechatAppThirdPlatformAuthorizeData:(%s) %s", codeName, decryptMsg)
	authorizeData := new(WechatAppThirdPlatformAuthorizeData)
	err = xml.Unmarshal([]byte(decryptMsg), authorizeData)
	if nil != err {
		_ = gokits.LOG.Warn("Unmarshal DecryptMsg error:(%s) %s", codeName, err.Error())
		return nil, err
	}

	return authorizeData, nil
}

// /accept-app-authorization/{codeName:string}
const acceptAppAuthorizationPath = "/accept-app-authorization/"

func acceptAppAuthorization(writer http.ResponseWriter, request *http.Request) {
	codeName := trimPrefixPath(request, acceptAppAuthorizationPath)
	if 0 != len(codeName) {
		authorizeData, err := parseWechatAppThirdPlatformAuthorizeData(codeName, request)
		if nil == err {
			switch authorizeData.InfoType {

			case "component_verify_ticket":
				_, _ = updateWechatAppThirdPlatformTicket(codeName, authorizeData.ComponentVerifyTicket)

			case "authorized":
				AuthorizerAppid := authorizeData.AuthorizerAppid
				AuthorizationCode := authorizeData.AuthorizationCode
				_, _ = enableWechatAppThirdPlatformAuthorizer(codeName,
					AuthorizerAppid, AuthorizationCode, authorizeData.PreAuthCode)
				go wechatAppThirdPlatformAuthorizerTokenCreator(codeName,
					AuthorizerAppid, AuthorizationCode)

			case "updateauthorized":
				AuthorizerAppid := authorizeData.AuthorizerAppid
				AuthorizationCode := authorizeData.AuthorizationCode
				_, _ = enableWechatAppThirdPlatformAuthorizer(codeName,
					AuthorizerAppid, AuthorizationCode, authorizeData.PreAuthCode)
				go wechatAppThirdPlatformAuthorizerTokenCreator(codeName,
					AuthorizerAppid, AuthorizationCode)

			case "unauthorized":
				AuthorizerAppid := authorizeData.AuthorizerAppid
				_, _ = disableWechatAppThirdPlatformAuthorizer(codeName, AuthorizerAppid)
				// delete cache
				_, _ = wechatAppThirdPlatformAuthorizerTokenCache.Delete(
					WechatAppThirdPlatformAuthorizerKey{
						CodeName: codeName, AuthorizerAppId: AuthorizerAppid})

			}
		}
	}

	// 接收到定时推送component_verify_ticket后必须直接返回字符串success
	gokits.ResponseText(writer, "success")
}

type WechatAppThirdPlatformPreAuthCodeResponse struct {
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int    `json:"expires_in"`
}

func wechatAppThirdPlatformPreAuthCodeRequestor(codeName interface{}) (map[string]string, error) {
	cache, err := wechatAppThirdPlatformTokenCache.Value(codeName)
	if nil != err {
		return nil, err
	}
	tokenItem := cache.Data().(*WechatAppThirdPlatformToken)

	result, err := gokits.NewHttpReq(wechatAppThirdPlatformPreAuthCodeURL+tokenItem.AccessToken).
		RequestBody(gokits.Json(map[string]string{"component_appid": tokenItem.AppId})).
		Prop("Content-Type", "application/json").Post()
	gokits.LOG.Trace("Request WechatAppThirdPlatformPreAuthCode Response:(%s) %s", codeName, result)
	if nil != err {
		return nil, err
	}

	response := gokits.UnJson(result, new(WechatAppThirdPlatformPreAuthCodeResponse)).(*WechatAppThirdPlatformPreAuthCodeResponse)
	if nil == response || 0 == len(response.PreAuthCode) {
		return nil, &UnexpectedError{Message: "Request WechatAppThirdPlatformPreAuthCode Failed: " + result}
	}
	return map[string]string{
		"APP_ID":        tokenItem.AppId,
		"PRE_AUTH_CODE": response.PreAuthCode,
		"EXPIRES_IN":    gokits.StrFromInt(response.ExpiresIn)}, nil
}

// /app-authorize-component-scan/{codeName:string}
const appAuthorizeComponentScanPath = "/app-authorize-component-scan/"
const scanRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/app-authorize-component-scan/")) + "/app-authorize-redirect/%s" + "%s";
    location.replace(
        "https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri)
    );
</script></head></html>
`

func appAuthorizeComponentScan(writer http.ResponseWriter, request *http.Request) {
	codeName := trimPrefixPath(request, appAuthorizeComponentScanPath)
	if 0 == len(codeName) {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
		return
	}

	response, err := wechatAppThirdPlatformPreAuthCodeRequestor(codeName)
	if nil != err {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
		return
	}
	appId := response["APP_ID"]
	preAuthCode := response["PRE_AUTH_CODE"]

	redirectQuery := request.URL.RawQuery
	if 0 != len(redirectQuery) {
		redirectQuery = "?" + redirectQuery
	}

	gokits.ResponseHtml(writer, fmt.Sprintf(scanRedirectPageHtmlFormat, codeName, redirectQuery, appId, preAuthCode))
}

// /app-authorize-component-link/{codeName:string}
const appAuthorizeComponentLinkPath = "/app-authorize-component-link/"
const linkRedirectPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/app-authorize-component-link/")) + "/app-authorize-redirect/%s" + "%s";
    location.replace(
        "https://mp.weixin.qq.com/safe/bindcomponent?action=bindcomponent&no_scan=1&component_appid=%s&pre_auth_code=%s&redirect_uri=" + encodeURIComponent(redirect_uri) + "#wechat_redirect"
    );
</script></head></html>
`

func appAuthorizeComponentLink(writer http.ResponseWriter, request *http.Request) {
	codeName := trimPrefixPath(request, appAuthorizeComponentLinkPath)
	if 0 == len(codeName) {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
		return
	}

	response, err := wechatAppThirdPlatformPreAuthCodeRequestor(codeName)
	if nil != err {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
		return
	}
	appId := response["APP_ID"]
	preAuthCode := response["PRE_AUTH_CODE"]

	redirectQuery := request.URL.RawQuery
	if 0 != len(redirectQuery) {
		redirectQuery = "?" + redirectQuery
	}

	gokits.ResponseHtml(writer, fmt.Sprintf(linkRedirectPageHtmlFormat, codeName, redirectQuery, appId, preAuthCode))
}

// /authorize-redirect/{codeName:string}
const appAuthorizeRedirectPath = "/app-authorize-redirect/"
const appAuthorizedPageHtmlFormat = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style><script>var p="%s";0!=p.length&&location.replace(p);</script></head><body><div class="tips">授权成功</div></body></html>
`

func appAuthorizeRedirect(writer http.ResponseWriter, request *http.Request) {
	codeName := trimPrefixPath(request, appAuthorizeRedirectPath)
	if 0 == len(codeName) {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Empty"}))
		return
	}

	cache, err := wechatAppThirdPlatformConfigCache.Value(codeName)
	if nil != err {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "CodeName is Illegal"}))
		return
	}
	config := cache.Data().(*WechatAppThirdPlatformConfig)
	redirectUrl := config.RedirectURL
	redirectQuery := request.URL.RawQuery

	if 0 != len(redirectUrl) && 0 != len(redirectQuery) {
		redirectUrl = redirectUrl + "?" + redirectQuery
	}

	gokits.ResponseHtml(writer, fmt.Sprintf(appAuthorizedPageHtmlFormat, redirectUrl))
}

// /query-wechat-app-authorizer-token/{codeName:string}/{authorizerAppId:string}
const queryWechatAppAuthorizerTokenPath = "/query-wechat-app-authorizer-token/"

func queryWechatAppAuthorizerToken(writer http.ResponseWriter, request *http.Request) {
	pathParams := trimPrefixPath(request, queryWechatAppAuthorizerTokenPath)
	if 0 == len(pathParams) {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
		return
	}
	ids := strings.Split(pathParams, "/")
	if 2 != len(ids) {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/authorizerAppId"}))
		return
	}

	codeName := ids[0]
	authorizerAppId := ids[1]
	cache, err := wechatAppThirdPlatformAuthorizerTokenCache.Value(
		WechatAppThirdPlatformAuthorizerKey{CodeName: codeName, AuthorizerAppId: authorizerAppId})
	if nil != err {
		gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
		return
	}
	token := cache.Data().(*WechatAppThirdPlatformAuthorizerToken)
	gokits.ResponseJson(writer, gokits.Json(map[string]string{
		"appId":           token.AppId,
		"authorizerAppId": token.AuthorizerAppId,
		"token":           token.AuthorizerAccessToken}))
}
