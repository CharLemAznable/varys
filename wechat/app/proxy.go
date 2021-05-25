package app

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var proxy *httputil.ReverseProxy
var mpLoginProxy *httputil.ReverseProxy

func proxyInitialize() {
    baseURL, err := url.Parse(proxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultProxyURL)
    }
    proxy = gokits.ReverseProxy(baseURL)
}

func mpLoginProxyInitialize() {
    baseURL, err := url.Parse(mpLoginProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultMpLoginProxyURL)
    }
    mpLoginProxy = gokits.ReverseProxy(baseURL)
}
