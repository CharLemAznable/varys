package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/CharLemAznable/gokits"
	"os"
	"strings"
)

type AppConfig struct {
	Port        int
	ContextPath string
	ConnectName string
}

var appConfig AppConfig
var _configFile string
var db *gokits.Gql

func init() {
	gokits.LOG.LoadConfiguration("logback.xml")

	flag.StringVar(&_configFile, "configFile", "appConfig.toml", "config file path")
	flag.Parse()

	if _, err := toml.DecodeFile(_configFile, &appConfig); err != nil {
		gokits.LOG.Crashf("config file decode error: %s", err.Error())
	}

	gokits.If(0 == appConfig.Port, func() {
		appConfig.Port = 4236
	})
	gokits.If(0 != len(appConfig.ContextPath), func() {
		gokits.Unless(strings.HasPrefix(appConfig.ContextPath, "/"),
			func() { appConfig.ContextPath = "/" + appConfig.ContextPath })
		gokits.If(strings.HasSuffix(appConfig.ContextPath, "/"),
			func() { appConfig.ContextPath = appConfig.ContextPath[:len(appConfig.ContextPath)-1] })
	})
	gokits.If(0 == len(appConfig.ConnectName), func() {
		appConfig.ConnectName = "Default"
	})

	gokits.LOG.Debug("appConfig: %s", gokits.Json(appConfig))

	// init db config
	gokits.LoadGqlConfigFile("gql.yaml")
	_db, err := gokits.NewGql(appConfig.ConnectName)
	if nil != err {
		_ = gokits.LOG.Error("Missing db config: %s in gql.yaml", appConfig.ConnectName)
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
		_ = gokits.LOG.Error("Query Configuration Err: %s", err.Error())
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
