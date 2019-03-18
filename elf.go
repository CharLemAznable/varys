package varys

import (
    . "github.com/CharLemAznable/gokits"
    "net/http"
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
    return strings.TrimPrefix(request.URL.Path, _path+subPath)
}
