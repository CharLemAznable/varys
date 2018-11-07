package main

import (
    "github.com/CharLemAznable/gcache"
    "github.com/CharLemAznable/gql"
    "log"
    "os"
    "time"
)

var db *gql.Gql

func init() {
    // init db config
    gql.LoadConfigFile("gql.yaml")
    _db, err := gql.Default()
    if nil != err {
        log.Println("Missing db config: Default in gql.yaml")
        os.Exit(-1)
    }
    db = _db

    // query app config -> map[string]string
    configs, err := db.Sql(queryConfigurationSQL).Query()
    if nil != err {
        log.Println("Query Configuration Err: ", err)
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

    // load app config
    {
        _path := configMap["path"]
        If(0 != len(_path), func() { path = _path })
    }
    {
        _port := configMap["port"]
        If(0 != len(_port), func() { port = _port })
    }
    {
        _wechatAPITokenURL := configMap["wechatAPITokenURL"]
        If(0 != len(_wechatAPITokenURL), func() { wechatAPITokenURL = _wechatAPITokenURL })
    }
    {
        wechatAPITokenConfigLifeSpanString := configMap["wechatAPITokenConfigLifeSpan"]
        If(0 != len(wechatAPITokenConfigLifeSpanString), func() {
            lifeSpan, err := Int64FromStr(wechatAPITokenConfigLifeSpanString)
            if nil == err {
                wechatAPITokenConfigLifeSpan = time.Minute * time.Duration(lifeSpan)
            }
        })
    }
    {
        wechatAPITokenLifeSpanString := configMap["wechatAPITokenLifeSpan"]
        If(0 != len(wechatAPITokenLifeSpanString), func() {
            lifeSpan, err := Int64FromStr(wechatAPITokenLifeSpanString)
            if nil == err {
                wechatAPITokenLifeSpan = time.Minute * time.Duration(lifeSpan)
            }
        })
    }
    {
        wechatAPITokenTempLifeSpanString := configMap["wechatAPITokenTempLifeSpan"]
        If(0 != len(wechatAPITokenTempLifeSpanString), func() {
            lifeSpan, err := Int64FromStr(wechatAPITokenTempLifeSpanString)
            if nil == err {
                wechatAPITokenTempLifeSpan = time.Minute * time.Duration(lifeSpan)
            }
        })
    }

    // build cache loader
    wechatAPITokenConfigCache = gcache.CacheExpireAfterWrite("WechatAPITokenConfig")
    wechatAPITokenConfigCache.SetDataLoader(wechatAPITokenConfigLoader)
    wechatAPITokenCache = gcache.CacheExpireAfterWrite("wechatAPIToken")
    wechatAPITokenCache.SetDataLoader(wechatAPITokenLoader)
}
