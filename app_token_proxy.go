package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var wechatAppProxy *httputil.ReverseProxy
var wechatMpProxy *httputil.ReverseProxy

func wechatAppProxyInitialize() {
    baseURL, err := url.Parse(wechatAppProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatAppProxyURL)
    }
    wechatAppProxy = gokits.ReverseProxy(baseURL)
}

func wechatMpProxyInitialize() {
    baseURL, err := url.Parse(wechatMpProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatMpProxyURL)
    }
    wechatMpProxy = gokits.ReverseProxy(baseURL)
}
