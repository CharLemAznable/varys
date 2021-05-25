package apptp

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var authProxy *httputil.ReverseProxy
var authMpLoginProxy *httputil.ReverseProxy

func authProxyInitialize() {
    baseURL, err := url.Parse(authProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultAuthProxyURL)
    }
    authProxy = gokits.ReverseProxy(baseURL)
}

func authMpLoginProxyInitialize() {
    baseURL, err := url.Parse(authMpLoginProxyURL)
    if err != nil {
        baseURL, _ = url.Parse(DefaultAuthMpLoginProxyURL)
    }
    authMpLoginProxy = gokits.ReverseProxy(baseURL)
}
