package app

import (
    "bytes"
    "crypto/sha256"
    "fmt"
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "io"
    "io/ioutil"
    "net/http"
    "strings"
    "time"
)

func init() {
    RegisterHandler(func(mux *http.ServeMux) {
        gokits.HandleFunc(mux, fengniaoAppAuthPath, fengniaoAppAuth)
        gokits.HandleFunc(mux, fengniaoAppAuthCallbackPath, Post(fengniaoAppAuthCallback))
        gokits.HandleFunc(mux, queryFengniaoAppTokenPath, queryFengniaoAppToken)
        gokits.HandleFunc(mux, proxyFengniaoAppPath, proxyFengniaoApp, gokits.GzipResponseDisabled)
        gokits.HandleFunc(mux, fengniaoAppCallbackPath, fengniaoAppCallback)
    })
}

// /fengniao-app-auth/{codeName:string}
const fengniaoAppAuthPath = "/fengniao-app-auth/"
const fengniaoAppAuthPageHtmlFormat = `
<html><head><script>
    location.replace(
        "https://open.ele.me/app-auth?app_id=%s&dev_id=%s"
    );
</script></head></html>
`

func fengniaoAppAuth(writer http.ResponseWriter, request *http.Request) {
    codeName := TrimPrefixPath(request, fengniaoAppAuthPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*FengniaoAppConfig)

    gokits.ResponseHtml(writer, fmt.Sprintf(
        fengniaoAppAuthPageHtmlFormat, config.AppId, config.DevId))
}

// /fengniao-app-auth-callback/{codeName:string}
const fengniaoAppAuthCallbackPath = "/fengniao-app-auth-callback/"

type AuthCallbackRequest struct {
    Code       string `json:"code"`
    MerchantId string `json:"merchant_id"`
    Scope      string `json:"scope"`
    State      string `json:"state"`
}

func fengniaoAppAuthCallback(writer http.ResponseWriter, request *http.Request) {
    codeName := TrimPrefixPath(request, fengniaoAppAuthCallbackPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*FengniaoAppConfig)
    callbackUrl := config.CallbackUrl

    body, err := gokits.RequestBody(request)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    callback := gokits.UnJson(body, new(AuthCallbackRequest)).(*AuthCallbackRequest)

    go func() {
        tokenCreator(codeName, config, callback)

        if "" != callbackUrl {
            callbackData := map[string]interface{}{
                "callback_business_type": "authNotify",
                "param": map[string]string{
                    "merchant_id": callback.MerchantId}}
            rsp, err := gokits.NewHttpReq(callbackUrl).
                RequestBody(gokits.Json(callbackData)).
                Prop("Content-Type", "application/json").Post()
            if nil != err {
                golog.Errorf("Callback Error: %s", err.Error())
            }
            golog.Debugf("Callback Response: %s", rsp)
        }
    }()
    gokits.ResponseText(writer, "Success")
}

// /query-fengniao-app-token/{codeName:string}/{merchantId:string}
const queryFengniaoAppTokenPath = "/query-fengniao-app-token/"

func queryFengniaoAppToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, queryFengniaoAppTokenPath)
    if "" == pathParams {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/merchantId"}))
        return
    }

    codeName := ids[0]
    merchantId := ids[1]
    tokenCacheData, err := tokenCache.Value(
        FengniaoAppTokenKey{CodeName: codeName, MerchantId: merchantId})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := tokenCacheData.Data().(*FengniaoAppToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}

// /proxy-fengniao-app/{codeName:string}/{merchantId:string}/...
const proxyFengniaoAppPath = "/proxy-fengniao-app/"

func proxyFengniaoApp(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, proxyFengniaoAppPath)
    if "" == pathParams {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    splits := strings.SplitN(pathParams, "/", 3)

    codeName := splits[0]
    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*FengniaoAppConfig)

    merchantId := splits[1]
    cache, err := tokenCache.Value(
        FengniaoAppTokenKey{CodeName: codeName, MerchantId: merchantId})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    tokenCacheData := cache.Data().(*FengniaoAppToken)
    appId := tokenCacheData.AppId
    accessToken := tokenCacheData.AccessToken

    actualPath := splits[2]
    if "" == actualPath {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "proxy PATH is Empty"}))
        return
    }

    data, err := gokits.RequestBody(request) // 向代理发起的请求仅需提供业务参数
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }

    params := make(map[string]string)
    params["app_id"] = appId
    timestamp := gokits.StrFromInt64(time.Now().UnixNano() / 1e6)
    params["timestamp"] = timestamp
    params["access_token"] = accessToken
    params["merchant_id"] = merchantId
    params["business_data"] = data
    params["version"] = "1.0"
    plainText := config.AppSecret + "access_token=" + accessToken +
        "&app_id=" + appId + "&business_data=" + data +
        "&merchant_id=" + merchantId + "&timestamp=" + timestamp + "&version=1.0"
    signature := fmt.Sprintf("%x", sha256.Sum256([]byte(plainText)))
    params["signature"] = signature
    body := bytes.NewBuffer([]byte(gokits.Json(params))) // 代理请求体

    // 重写请求体, ref: http/request.go(line:838)
    req := request
    req.URL.Path = actualPath
    req.Body = ioutil.NopCloser(body)
    req.ContentLength = int64(body.Len())
    bodyBuf := body.Bytes()
    req.GetBody = func() (io.ReadCloser, error) {
        r := bytes.NewReader(bodyBuf)
        return ioutil.NopCloser(r), nil
    }
    proxy.ServeHTTP(writer, req)
}

type CallbackRequest struct {
    AppId        string `json:"app_id"`
    Timestamp    string `json:"timestamp"`
    Signature    string `json:"signature"`
    BusinessData string `json:"business_data"`
}

// /fengniao-app-callback/{codeName:string}
const fengniaoAppCallbackPath = "/fengniao-app-callback/"

func fengniaoAppCallback(writer http.ResponseWriter, request *http.Request) {
    codeName := TrimPrefixPath(request, fengniaoAppCallbackPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*FengniaoAppConfig)
    appId := config.AppId
    callbackUrl := config.CallbackUrl

    body, err := gokits.RequestBody(request)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "read request body error"}))
        return
    }

    callbackRequest := gokits.UnJson(body, new(CallbackRequest)).(*CallbackRequest)
    businessData := callbackRequest.BusinessData
    if appId != callbackRequest.AppId {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "app_id mismatch"}))
        return
    }
    plainText := config.AppSecret + "app_id=" + appId +
        "&business_data=" + businessData + "&timestamp=" + callbackRequest.Timestamp
    signature := fmt.Sprintf("%x", sha256.Sum256([]byte(plainText)))
    if signature != callbackRequest.Signature {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "signature mismatch"}))
        return
    }

    go func() {
        if "" != callbackUrl {
            rsp, err := gokits.NewHttpReq(callbackUrl).
                RequestBody(businessData).
                Prop("Content-Type", "application/json").Post()
            if nil != err {
                golog.Errorf("Callback Error: %s", err.Error())
            }
            golog.Debugf("Callback Response: %s", rsp)
        }
    }()

    gokits.ResponseJson(writer, gokits.Json(map[string]string{"result": "OK"}))
}
