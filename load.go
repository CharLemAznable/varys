package varys

import (
    . "github.com/CharLemAznable/goutils"
    "github.com/CharLemAznable/gql"
    log "github.com/CharLemAznable/log4go"
    "os"
    "time"
)

var db *gql.Gql

func load() {
    log.LoadConfiguration("logback.xml")

    // init db config
    gql.LoadConfigFile("gql.yaml")
    _db, err := gql.Default()
    if nil != err {
        log.Error("Missing db config: Default in gql.yaml")
        os.Exit(-1)
    }
    db = _db

    // query app config -> map[string]string
    configs, err := db.Sql(`
SELECT C.CONFIG_NAME ,C.CONFIG_VALUE
  FROM APP_CONFIG C
 WHERE C.ENABLED = 1
`).Query()
    if nil != err {
        log.Error("Query Configuration Err: %s", err.Error())
        os.Exit(-1)
    }
    configMap := make(map[string]string)
    for _, config := range configs {
        name := config["CONFIG_NAME"]
        value := config["CONFIG_VALUE"]
        if 0 != len(name) && 0 != len(value) {
            configMap[name] = value
        }
    }

    wechatAPITokenInitialize(configMap)
    wechatThirdPlatformAuthorizerTokenInitialize(configMap)
    wechatCorpTokenInitialize(configMap)
    wechatCorpThirdPlatformAuthorizerTokenInitialize(configMap)
}

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
