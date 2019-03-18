package varys

import (
    "net/http/httputil"
    "net/url"
)

var wechatCorpProxy *httputil.ReverseProxy

func wechatCorpProxyInitialize() {
    baseURL, err := url.Parse(wechatCorpProxyURL)
    if err != nil {
        baseURL, _ = url.Parse("https://api.weixin.qq.com/cgi-bin/")
    }
    wechatCorpProxy = httputil.NewSingleHostReverseProxy(baseURL)
}
