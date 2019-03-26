package varys

import (
    . "github.com/CharLemAznable/gokits"
    "os"
)

var db *Gql

func load() {
    LOG.LoadConfiguration("logback.xml")

    // init db config
    LoadGqlConfigFile("gql.yaml")
    _db, err := DefaultGql()
    if nil != err {
        _ = LOG.Error("Missing db config: Default in gql.yaml")
        os.Exit(-1)
    }
    db = _db

    // query app config -> map[string]string
    configs, err := db.New().Sql(`
SELECT C.CONFIG_NAME ,C.CONFIG_VALUE
  FROM APP_CONFIG C
 WHERE C.ENABLED = 1
`).Query()
    if nil != err {
        _ = LOG.Error("Query Configuration Err: %s", err.Error())
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

    wechatAppTokenLoad(configMap)
    wechatAppThirdPlatformAuthorizerTokenLoad(configMap)
    wechatCorpTokenLoad(configMap)
    wechatCorpThirdPlatformAuthorizerTokenLoad(configMap)
}
