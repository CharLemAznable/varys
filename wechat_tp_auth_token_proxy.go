package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var wechatTpAuthProxy *httputil.ReverseProxy
var wechatTpAuthMpLoginProxy *httputil.ReverseProxy

func wechatTpAuthProxyInitialize() {
    baseURL, err := url.Parse(wechatTpAuthProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatTpAuthProxyURL)
    }
    wechatTpAuthProxy = gokits.ReverseProxy(baseURL)
}

func wechatTpAuthMpLoginProxyInitialize() {
    baseURL, err := url.Parse(wechatTpAuthMpLoginProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatTpAuthMpLoginProxyURL)
    }
    wechatTpAuthMpLoginProxy = gokits.ReverseProxy(baseURL)
}
