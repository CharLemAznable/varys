package varys

import (
    . "github.com/CharLemAznable/gokits"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "time"
)

func urlConfigLoader(configStr string, loader func(configURL string)) {
    If(0 != len(configStr), func() {
        loader(configStr)
    })
}

func lifeSpanConfigLoader(configStr string, loader func(configVal time.Duration)) {
    If(0 != len(configStr), func() {
        lifeSpan, err := Int64FromStr(configStr)
        if nil == err {
            loader(time.Duration(lifeSpan))
        }
    })
}

func trimPrefixPath(request *http.Request, subPath string) string {
    return strings.TrimPrefix(request.URL.Path, JoinPathComponent(_path, subPath))
}

func reverseProxy(target *url.URL) *httputil.ReverseProxy {
    targetQuery := target.RawQuery
    director := func(req *http.Request) {
        req.Host = target.Host // Different from the default NewSingleHostReverseProxy()

        req.URL.Scheme = target.Scheme
        req.URL.Host = target.Host
        req.URL.Path = JoinPathComponent(target.Path, req.URL.Path)
        if targetQuery == "" || req.URL.RawQuery == "" {
            req.URL.RawQuery = targetQuery + req.URL.RawQuery
        } else {
            req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
        }
        if _, ok := req.Header["User-Agent"]; !ok {
            req.Header.Set("User-Agent", "")
        }
    }
    return &httputil.ReverseProxy{Director: director}
}
