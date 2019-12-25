package main

import (
    "compress/gzip"
    "github.com/CharLemAznable/gokits"
    "io"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "time"
)

func urlConfigLoader(configStr string, loader func(configURL string)) {
    gokits.If(0 != len(configStr), func() {
        loader(configStr)
    })
}

func lifeSpanConfigLoader(configStr string, loader func(configVal time.Duration)) {
    gokits.If(0 != len(configStr), func() {
        lifeSpan, err := gokits.Int64FromStr(configStr)
        if nil == err {
            loader(time.Duration(lifeSpan))
        }
    })
}

func trimPrefixPath(request *http.Request, subPath string) string {
    return strings.TrimPrefix(request.URL.Path, gokits.PathJoin(appConfig.ContextPath, subPath))
}

func reverseProxy(target *url.URL) *httputil.ReverseProxy {
    targetQuery := target.RawQuery
    director := func(req *http.Request) {
        req.Host = target.Host // Different from the default NewSingleHostReverseProxy()

        req.URL.Scheme = target.Scheme
        req.URL.Host = target.Host
        req.URL.Path = gokits.PathJoin(target.Path, req.URL.Path)
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

func dumpRequest(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        // Save a copy of this request for debugging.
        requestDump, err := httputil.DumpRequest(request, true)
        if err != nil {
            _ = gokits.LOG.Error(err)
        }
        gokits.LOG.Debug(string(requestDump))
        handlerFunc(writer, request)
    }
}

type GzipResponseWriter struct {
    io.Writer
    http.ResponseWriter
}

func (w GzipResponseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}

func gzipHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
            handlerFunc(writer, request)
            return
        }
        writer.Header().Set("Content-Encoding", "gzip")
        gz := gzip.NewWriter(writer)
        defer func() { _ = gz.Close() }()
        gzr := GzipResponseWriter{Writer: gz, ResponseWriter: writer}
        handlerFunc(gzr, request)
    }
}
