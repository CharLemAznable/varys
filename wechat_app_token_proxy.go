package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var wechatAppProxy *httputil.ReverseProxy
var wechatAppMpLoginProxy *httputil.ReverseProxy

func wechatAppProxyInitialize() {
    baseURL, err := url.Parse(wechatAppProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatAppProxyURL)
    }
    wechatAppProxy = gokits.ReverseProxy(baseURL)
}

func wechatAppMpLoginProxyInitialize() {
    baseURL, err := url.Parse(wechatAppMpLoginProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatAppMpLoginProxyURL)
    }
    wechatAppMpLoginProxy = gokits.ReverseProxy(baseURL)
}
