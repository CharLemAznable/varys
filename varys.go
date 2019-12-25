package main

import (
	"github.com/CharLemAznable/gokits"
	"net/http"
)

const welcomePath = "/welcome"

func welcome(writer http.ResponseWriter, request *http.Request) {
	gokits.ResponseText(writer, `Three great men, a king, a priest, and a rich man.
Between them stands a common sellsword.
Each great man bids the sellsword kill the other two.
Who lives, who dies?
`)
}
