package app

import (
    "bytes"
    "crypto/md5"
    "fmt"
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "io"
    "io/ioutil"
    "mime/multipart"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "time"
)

func init() {
    RegisterHandler(func(mux *http.ServeMux) {
        gokits.HandleFunc(mux, shansongAppAuthPath, shansongAppAuth)
        gokits.HandleFunc(mux, shansongAppAuthRedirectPath, shansongAppAuthRedirect)
        gokits.HandleFunc(mux, queryShansongAppTokenPath, queryShansongAppToken)
        gokits.HandleFunc(mux, proxyShansongAppDeveloperPath, Post(proxyShansongAppDeveloper), gokits.GzipResponseDisabled)
        gokits.HandleFunc(mux, proxyShansongAppMerchantPath, Post(proxyShansongAppMerchant), gokits.GzipResponseDisabled)
        gokits.HandleFunc(mux, proxyShansongAppFilePath, Post(proxyShansongAppFile), gokits.GzipResponseDisabled)
        gokits.HandleFunc(mux, shansongAppCallbackPath, Post(shansongAppCallback))
    })
}

// /shansong-app-auth/{codeName:string}/{merchantCode:string}
const shansongAppAuthPath = "/shansong-app-auth/"
const shansongAppAuthPageHtmlFormat = `
<html><head><script>
    redirect_uri = location.href.substring(0, location.href.indexOf("/shansong-app-auth/")) + "/shansong-app-auth-redirect/%s";
    location.replace(
        "%s?response_type=code&client_id=%s&scope=shop_open_api&redirect_uri=" + encodeURIComponent(redirect_uri) + "&state=%s"
    );
</script></head></html>
`

func shansongAppAuth(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, shansongAppAuthPath)
    if "" == pathParams {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/merchantCode"}))
        return
    }
    codeName := ids[0]
    merchantCode := ids[1]

    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*ShansongAppConfig)

    gokits.ResponseHtml(writer, fmt.Sprintf(
        shansongAppAuthPageHtmlFormat, codeName, authBaseURL, config.AppId, merchantCode))
}

// /shansong-app-auth-redirect/{codeName:string}
const shansongAppAuthRedirectPath = "/shansong-app-auth-redirect/"
const shansongAppAuthRedirectPageHtmlFormat = `
<html><head><title>index</title><style type="text/css">
    body{max-width:640px;margin:0 auto;font-size:14px;-webkit-text-size-adjust:none;-moz-text-size-adjust:none;-ms-text-size-adjust:none;text-size-adjust:none}
    .tips{margin-top:40px;text-align:center;color:green}
</style><script>var p="%s";0!=p.length&&location.replace(p);</script></head><body><div class="tips">授权成功</div></body></html>
`

func shansongAppAuthRedirect(writer http.ResponseWriter, request *http.Request) {
    codeName := TrimPrefixPath(request, shansongAppAuthRedirectPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*ShansongAppConfig)
    appId := config.AppId
    redirectUrl := config.RedirectURL

    redirectQuery := request.URL.RawQuery
    authCode := request.URL.Query().Get("code")
    merchantCode := request.URL.Query().Get("state")
    if "" == authCode || "" == merchantCode {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Shop Unauthorized"}))
        return
    }

    go tokenCreator(codeName, appId, merchantCode, authCode)

    if "" != redirectUrl && "" != redirectQuery {
        redirectUrl = redirectUrl + "?" + redirectQuery
    }
    gokits.ResponseHtml(writer, fmt.Sprintf(shansongAppAuthRedirectPageHtmlFormat, redirectUrl))
}

// /query-shansong-app-token/{codeName:string}/{merchantCode:string}
const queryShansongAppTokenPath = "/query-shansong-app-token/"

func queryShansongAppToken(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, queryShansongAppTokenPath)
    if "" == pathParams {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    ids := strings.Split(pathParams, "/")
    if 2 != len(ids) {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Missing param codeName/merchantCode"}))
        return
    }

    codeName := ids[0]
    merchantCode := ids[1]
    tokenCacheData, err := tokenCache.Value(
        ShansongAppTokenKey{CodeName: codeName, MerchantCode: merchantCode})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    token := tokenCacheData.Data().(*ShansongAppToken)
    gokits.ResponseJson(writer, gokits.Json(token))
}

// /proxy-shansong-app-developer/{codeName:string}/{merchantCode:string}/...
const proxyShansongAppDeveloperPath = "/proxy-shansong-app-developer/"

func proxyShansongAppDeveloper(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, proxyShansongAppDeveloperPath)
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
    config := configCacheData.Data().(*ShansongAppConfig)
    appId := config.AppId
    appSecret := config.AppSecret

    merchantCode := splits[1]
    cache, err := tokenCache.Value(
        ShansongAppTokenKey{CodeName: codeName, MerchantCode: merchantCode})
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    tokenCacheData := cache.Data().(*ShansongAppToken)
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

    plainData := gokits.Condition("" != data, "data"+data, "").(string)
    timestamp := gokits.StrFromInt64(time.Now().UnixNano() / 1e6)
    plainText := appSecret + "accessToken" + accessToken +
        "clientId" + appId + plainData + "timestamp" + timestamp
    signature := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))

    values := url.Values{}
    values.Add("accessToken", accessToken)
    values.Add("clientId", appId)
    values.Add("timestamp", timestamp)
    values.Add("sign", signature)
    if "" != data {
        values.Add("data", data)
    }
    proxy(writer, request, developerProxy,
        actualPath, bytes.NewBuffer([]byte(values.Encode())))
}

// /proxy-shansong-app-merchant/{codeName:string}/...
const proxyShansongAppMerchantPath = "/proxy-shansong-app-merchant/"

func proxyShansongAppMerchant(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, proxyShansongAppMerchantPath)
    if "" == pathParams {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    splits := strings.SplitN(pathParams, "/", 2)

    codeName := splits[0]
    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*ShansongAppConfig)
    appId := config.AppId
    appSecret := config.AppSecret

    actualPath := splits[1]
    if "" == actualPath {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "proxy PATH is Empty"}))
        return
    }

    data, err := gokits.RequestBody(request) // 向代理发起的请求仅需提供业务参数
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }

    plainData := gokits.Condition("" != data, "data"+data, "").(string)
    timestamp := gokits.StrFromInt64(time.Now().UnixNano() / 1e6)
    plainText := appSecret + "clientId" + appId + plainData + "timestamp" + timestamp
    signature := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))

    values := url.Values{}
    values.Add("clientId", appId)
    values.Add("timestamp", timestamp)
    values.Add("sign", signature)
    if "" != data {
        values.Add("data", data)
    }
    proxy(writer, request, merchantProxy,
        actualPath, bytes.NewBuffer([]byte(values.Encode())))
}

// /proxy-shansong-app-file/{codeName:string}/...
const proxyShansongAppFilePath = "/proxy-shansong-app-file/"

func proxyShansongAppFile(writer http.ResponseWriter, request *http.Request) {
    pathParams := TrimPrefixPath(request, proxyShansongAppFilePath)
    if "" == pathParams {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "Path Params is Empty"}))
        return
    }
    splits := strings.SplitN(pathParams, "/", 2)

    codeName := splits[0]
    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*ShansongAppConfig)
    appId := config.AppId
    appSecret := config.AppSecret

    actualPath := splits[1]
    if "" == actualPath {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "proxy PATH is Empty"}))
        return
    }

    fileName, fileData, err := parseFile(request)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }

    body := &bytes.Buffer{}
    partWriter := multipart.NewWriter(body)
    part, err := partWriter.CreateFormFile("file", fileName)
    if err != nil {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    _, err = part.Write(fileData)

    timestamp := gokits.StrFromInt64(time.Now().UnixNano() / 1e6)
    plainText := appSecret + "clientId" + appId + "timestamp" + timestamp
    signature := fmt.Sprintf("%x", md5.Sum([]byte(plainText)))

    _ = partWriter.WriteField("clientId", appId)
    _ = partWriter.WriteField("timestamp", timestamp)
    _ = partWriter.WriteField("sign", signature)

    err = partWriter.Close()
    if err != nil {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": err.Error()}))
        return
    }
    proxy(writer, request, fileProxy, actualPath, body)
}

func proxy(writer http.ResponseWriter, request *http.Request,
    proxy *httputil.ReverseProxy, actualPath string, actualBody *bytes.Buffer) {

    // 重写请求体, ref: http/request.go(line:838)
    req := request
    req.URL.Path = actualPath
    req.Body = ioutil.NopCloser(actualBody) // 代理请求体
    req.ContentLength = int64(actualBody.Len())
    bodyBuf := actualBody.Bytes()
    req.GetBody = func() (io.ReadCloser, error) {
        r := bytes.NewReader(bodyBuf)
        return ioutil.NopCloser(r), nil
    }
    proxy.ServeHTTP(writer, req)
}

func parseFile(request *http.Request) (string, []byte, error) {
    err := request.ParseMultipartForm(1024)
    if nil != err {
        return "", nil, err
    }

    fileHeader := request.MultipartForm.File["file"][0]
    fileName := fileHeader.Filename

    file, err := fileHeader.Open()
    if nil != err {
        return "", nil, err
    }
    defer func() { _ = file.Close() }()

    fileData, err := ioutil.ReadAll(file)
    if nil != err {
        return "", nil, err
    }

    return fileName, fileData, nil
}

// /shansong-app-callback/{codeName:string}
const shansongAppCallbackPath = "/shansong-app-callback/"

func shansongAppCallback(writer http.ResponseWriter, request *http.Request) {
    codeName := TrimPrefixPath(request, shansongAppCallbackPath)
    if "" == codeName {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Empty"}))
        return
    }

    configCacheData, err := configCache.Value(codeName)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "codeName is Illegal"}))
        return
    }
    config := configCacheData.Data().(*ShansongAppConfig)
    callbackUrl := config.CallbackURL

    body, err := gokits.RequestBody(request)
    if nil != err {
        gokits.ResponseJson(writer, gokits.Json(map[string]string{"error": "read request body error"}))
        return
    }

    go func() {
        if "" != callbackUrl {
            rsp, err := gokits.NewHttpReq(callbackUrl).
                RequestBody(body).Prop("Content-Type", "application/json").Post()
            if nil != err {
                golog.Errorf("Callback Error: %s", err.Error())
            }
            golog.Debugf("Callback Response: %s", rsp)
        }
    }()

    gokits.ResponseJson(writer, gokits.Json(
        map[string]interface{}{"status": 200, "msg": ""}))
}
