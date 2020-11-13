package main

import (
    "github.com/CharLemAznable/gokits"
    "net/http"
    "strings"
)

func trimPrefixPath(request *http.Request, subPath string) string {
    return strings.TrimPrefix(request.URL.Path, gokits.PathJoin(globalConfig.ContextPath, subPath))
}
