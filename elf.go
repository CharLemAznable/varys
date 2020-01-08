package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http"
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
