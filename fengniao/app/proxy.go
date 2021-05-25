package app

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var proxy *httputil.ReverseProxy

func proxyInitialize() {
    baseURL, err := url.Parse(proxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultProxyURL)
    }
    proxy = gokits.ReverseProxy(baseURL)
}
