package varys

import (
    "encoding/json"
    "encoding/xml"
    "github.com/CharLemAznable/httpreq"
    "github.com/CharLemAznable/wechataes"
    "io/ioutil"
    "log"
    "net/http"
)

var wechatAPITokenURL = "https://api.weixin.qq.com/cgi-bin/token"

type WechatAPITokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatAPITokenRequestor(appId interface{}) (string, int, error) {
    cache, err := wechatAPITokenConfigCache.Value(appId)
    if nil != err {
        return "", 0, err
    }
    config := cache.Data().(*WechatAPITokenConfig)

    result, err := httpreq.New(wechatAPITokenURL).Params(
        "grant_type", "client_credential",
        "appid", config.AppId, "secret", config.AppSecret).
        Prop("Content-Type",
            "application/x-www-form-urlencoded").Get()
    if nil != err {
        return "", 0, err
    }

    response := new(WechatAPITokenResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", 0, err
    }
    return response.AccessToken, response.ExpiresIn, nil
}

type WechatAuthorizeData struct {
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

func parseWechatAuthorizeData(appId string, request *http.Request) (*WechatAuthorizeData, error) {
    cache, err := wechatThirdPlatformCryptorCache.Value(appId)
    if nil != err {
        log.Println("GetWechatThirdPlatformCryptor error:", err.Error())
        return nil, err
    }
    cryptor := cache.Data().(*wechataes.WechatCryptor)

    body, err := ioutil.ReadAll(request.Body)
    if nil != err {
        log.Println("Request read Body error:", err.Error())
        return nil, err
    }

    err = request.ParseForm()
    if nil != err {
        log.Println("Request ParseForm error:", err.Error())
        return nil, err
    }

    params := request.Form
    msgSign := params.Get("msg_signature")
    timeStamp := params.Get("timestamp")
    nonce := params.Get("nonce")
    decryptMsg, err := cryptor.DecryptMsg(msgSign, timeStamp, nonce, string(body))
    if nil != err {
        log.Println("WechatCryptor DecryptMsg error:", err.Error())
        return nil, err
    }

    log.Println("微信推送消息(明文):", decryptMsg)
    authorizeData := new(WechatAuthorizeData)
    err = xml.Unmarshal([]byte(decryptMsg), authorizeData)
    if nil != err {
        log.Println("Unmarshal DecryptMsg error:", err.Error())
        return nil, err
    }

    return authorizeData, nil
}

var wechatThirdPlatformTokenURL = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"

type WechatThirdPlatformTokenResponse struct {
    ComponentAccessToken string `json:"component_access_token"`
    ExpiresIn            int    `json:"expires_in"`
}

func wechatThirdPlatformTokenRequestor(appId interface{}) (string, int, error) {
    cache, err := wechatThirdPlatformConfigCache.Value(appId)
    if nil != err {
        return "", 0, err
    }
    config := cache.Data().(*WechatThirdPlatformConfig)

    ticket, err := QueryWechatThirdPlatformTicket(appId.(string))
    if nil != err {
        return "", 0, err
    }

    result, err := httpreq.New(wechatThirdPlatformTokenURL).
        RequestBody(Json(map[string]string{
            "component_appid":         config.AppId,
            "component_appsecret":     config.AppSecret,
            "component_verify_ticket": ticket})).
        Prop("Content-Type", "application/json").Get()
    if nil != err {
        return "", 0, err
    }

    response := new(WechatThirdPlatformTokenResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", 0, err
    }
    return response.ComponentAccessToken, response.ExpiresIn, nil
}

var wechatThirdPlatformPreAuthCodeURL = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token="

type WechatThirdPlatformPreAuthCodeResponse struct {
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

func wechatThirdPlatformPreAuthCodeRequestor(appId interface{}) (string, int, error) {
    cache, err := wechatThirdPlatformTokenCache.Value(appId)
    if nil != err {
        return "", 0, err
    }
    tokenItem := cache.Data().(*WechatThirdPlatformToken)

    result, err := httpreq.New(wechatThirdPlatformPreAuthCodeURL + tokenItem.ComponentAccessToken).
        RequestBody(Json(map[string]string{"component_appid": appId.(string)})).
        Prop("Content-Type", "application/json").Get()
    if nil != err {
        return "", 0, err
    }

    response := new(WechatThirdPlatformPreAuthCodeResponse)
    err = json.Unmarshal([]byte(result), response)
    if nil != err {
        return "", 0, err
    }
    return response.PreAuthCode, response.ExpiresIn, nil
}
