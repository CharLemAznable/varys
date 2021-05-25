package base

import (
    "github.com/CharLemAznable/gokits"
    "net/http"
    "strings"
)

func TrimPrefixPath(request *http.Request, subPath string) string {
    return strings.TrimPrefix(request.URL.Path, gokits.PathJoin(config.ContextPath, subPath))
}
