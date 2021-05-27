package app

import (
    . "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "net/http"
)

func Run() {
    InitSqlxDB()
    Load()

    server := http.Server{Addr: ServerAddr(),
        Handler: Handle(http.NewServeMux())}
    if err := server.ListenAndServe(); err != nil {
        golog.Errorf("Start server Error: %s", err.Error())
    }
}
