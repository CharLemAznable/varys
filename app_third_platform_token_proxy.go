package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var wechatAppThirdPlatformProxy *httputil.ReverseProxy

func wechatAppThirdPlatformProxyInitialize() {
    baseURL, err := url.Parse(wechatAppThirdPlatformProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatAppThirdPlatformProxyURL)
    }
    wechatAppThirdPlatformProxy = gokits.ReverseProxy(baseURL)
}
