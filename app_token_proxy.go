package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var wechatAppProxy *httputil.ReverseProxy

func wechatAppProxyInitialize() {
    baseURL, err := url.Parse(wechantAppProxyURL)
    if err != nil {
        baseURL, _ = url.Parse("https://api.weixin.qq.com/cgi-bin/")
    }
    wechatAppProxy = gokits.ReverseProxy(baseURL)
}
