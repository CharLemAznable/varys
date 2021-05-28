package app

import (
    "github.com/CharLemAznable/gokits"
    "net/http/httputil"
    "net/url"
)

var developerProxy *httputil.ReverseProxy
var merchantProxy *httputil.ReverseProxy
var fileProxy *httputil.ReverseProxy

func proxyInitialize() {
    developerURL, err := url.Parse(developerProxyURL)
    if err != nil {
        developerURL, _ = url.Parse(DefaultDeveloperProxyURL)
    }
    developerProxy = gokits.ReverseProxy(developerURL)

    merchantURL, err := url.Parse(merchantProxyURL)
    if err != nil {
        merchantURL, _ = url.Parse(DefaultMerchantProxyURL)
    }
    merchantProxy = gokits.ReverseProxy(merchantURL)

    fileURL, err := url.Parse(fileProxyURL)
    if err != nil {
        fileURL, _ = url.Parse(DefaultFileProxyURL)
    }
    fileProxy = gokits.ReverseProxy(fileURL)
}
