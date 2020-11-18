package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var fengniaoAppProxy *httputil.ReverseProxy

func fengniaoAppProxyInitialize() {
    baseURL, err := url.Parse(fengniaoAppProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultFengniaoAppProxyURL)
    }
    fengniaoAppProxy = gokits.ReverseProxy(baseURL)
}
