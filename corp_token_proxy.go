package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var wechatCorpProxy *httputil.ReverseProxy

func wechatCorpProxyInitialize() {
    baseURL, err := url.Parse(wechatCorpProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultWechatCorpProxyURL)
    }
    wechatCorpProxy = gokits.ReverseProxy(baseURL)
}
