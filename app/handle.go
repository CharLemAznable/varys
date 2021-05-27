package app

import (
    "github.com/CharLemAznable/gokits"
    . "github.com/CharLemAznable/varys/base"
    "net/http"
)

func init() {
    RegisterHandler(func(mux *http.ServeMux) {
        gokits.HandleFunc(mux, "/", gokits.EmptyHandler, gokits.DumpRequestDisabled)
        gokits.HandleFunc(mux, "/welcome", welcome, gokits.DumpRequestDisabled)
    })
}

func welcome(writer http.ResponseWriter, request *http.Request) {
    gokits.ResponseText(writer, `Three great men, a king, a priest, and a rich man.
Between them stands a common sellsword.
Each great man bids the sellsword kill the other two.
Who lives, who dies?
`)
}
