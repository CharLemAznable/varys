package main

import (
    "github.com/CharLemAznable/varys"
    _ "github.com/CharLemAznable/varys/fengniao/app"
    _ "github.com/CharLemAznable/varys/toutiao/app"
    _ "github.com/CharLemAznable/varys/wechat/app"
    _ "github.com/CharLemAznable/varys/wechat/apptp"
    _ "github.com/CharLemAznable/varys/wechat/corp"
    _ "github.com/CharLemAznable/varys/wechat/corptp"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    varys.Run()
}
