package varys

import (
    . "github.com/CharLemAznable/gokits"
    "net/http"
    "net/http/httputil"
    "net/url"
)

var wechatCorpProxy *httputil.ReverseProxy

func wechatCorpProxyInitialize() {
    baseURL, err := url.Parse(wechatCorpProxyURL)
    if err != nil {
        baseURL, _ = url.Parse("https://qyapi.weixin.qq.com/cgi-bin/")
    }
    wechatCorpProxy = httputil.NewSingleHostReverseProxy(baseURL)
    wechatCorpProxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
        LOG.Info(request.URL)
        writer.Write([]byte(Json(map[string]string{"error": err.Error()})))
    }
}
