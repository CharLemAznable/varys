package app

import (
    "github.com/CharLemAznable/varys/base"
    "github.com/kataras/golog"
    "net/http"
)

func Run() {
    base.InitSqlxDB()
    base.Load()

    mux := http.NewServeMux()
    base.Handle(mux)

    server := http.Server{Addr: base.ServerAddr(), Handler: mux}
    if err := server.ListenAndServe(); err != nil {
        golog.Errorf("Start server Error: %s", err.Error())
    }
}
