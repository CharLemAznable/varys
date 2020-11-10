package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var wechatTpProxy *httputil.ReverseProxy

func wechatTpProxyInitialize() {
    baseURL, err := url.Parse(wechatTpProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatTpProxyURL)
    }
    wechatTpProxy = gokits.ReverseProxy(baseURL)
}
