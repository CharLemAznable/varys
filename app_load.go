package main

import (
    "flag"
    "github.com/BurntSushi/toml"
    "github.com/CharLemAznable/gokits"
    "github.com/CharLemAznable/sqlx"
    "github.com/kataras/golog"
    "strings"
    "testing"
    "unsafe"
)

type Config struct {
    gokits.HttpServerConfig

    LogLevel string

    DriverName      string
    DataSourceName  string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxIdleTime gokits.Duration
    ConnMaxLifetime gokits.Duration

    ClusterNodeAddresses string // notify all cluster node for cache delete event

    WechatAppTokenURL          string
    WechatAppConfigLifeSpan    gokits.Duration
    WechatAppTokenLifeSpan     gokits.Duration
    WechatAppTokenTempLifeSpan gokits.Duration
    WechatAppProxyURL          string
    WechatMpProxyURL           string
    WechatMpLoginProxyURL      string

    WechatTpTokenURL          string
    WechatTpConfigLifeSpan    gokits.Duration
    WechatTpCryptorLifeSpan   gokits.Duration
    WechatTpTokenLifeSpan     gokits.Duration
    WechatTpTokenTempLifeSpan gokits.Duration
    WechatTpProxyURL          string

    WechatTpPreAuthCodeURL        string
    WechatTpQueryAuthURL          string
    WechatTpRefreshAuthURL        string
    WechatTpAuthTokenLifeSpan     gokits.Duration
    WechatTpAuthTokenTempLifeSpan gokits.Duration

    WechatCorpTokenURL                string
    WechatCorpConfigLifeSpan          gokits.Duration
    WechatCorpTokenMaxLifeSpan        gokits.Duration
    WechatCorpTokenExpireCriticalSpan gokits.Duration
    WechatCorpProxyURL                string

    WechatCorpTpTokenURL                string
    WechatCorpTpConfigLifeSpan          gokits.Duration
    WechatCorpTpCryptorLifeSpan         gokits.Duration
    WechatCorpTpTokenMaxLifeSpan        gokits.Duration
    WechatCorpTpTokenExpireCriticalSpan gokits.Duration

    WechatCorpTpPreAuthCodeURL              string
    WechatCorpTpPermanentCodeURL            string
    WechatCorpTpAuthTokenURL                string
    WechatCorpTpPermanentCodeLifeSpan       gokits.Duration
    WechatCorpTpAuthTokenMaxLifeSpan        gokits.Duration
    WechatCorpTpAuthTokenExpireCriticalSpan gokits.Duration

    ToutiaoAppTokenURL          string
    ToutiaoAppConfigLifeSpan    gokits.Duration
    ToutiaoAppTokenLifeSpan     gokits.Duration
    ToutiaoAppTokenTempLifeSpan gokits.Duration

    FengniaoAppTokenURL          string
    FengniaoAppConfigLifeSpan    gokits.Duration
    FengniaoAppTokenLifeSpan     gokits.Duration
    FengniaoAppTokenTempLifeSpan gokits.Duration
    FengniaoAppProxyURL          string
    FengniaoCallbackAddress      string
}

var globalConfig = &Config{}
var db *sqlx.DB
var clusterNodeAddresses = make([]string, 0)

func init() {
    testing.Init()
    configFile := ""
    flag.StringVar(&configFile, "configFile",
        "config.toml", "config file path")
    flag.Parse()
    if _, err := toml.DecodeFile(configFile, globalConfig); err != nil {
        golog.Errorf("config file decode error: %s", err.Error())
    }

    fixedConfig(globalConfig)
    db = loadSqlxDB(globalConfig)
    fetchClusterNodes(globalConfig)

    wechatAppTokenLoad(globalConfig)
    wechatTpTokenLoad(globalConfig)
    wechatTpAuthTokenLoad(globalConfig)
    wechatCorpTokenLoad(globalConfig)
    wechatCorpTpTokenLoad(globalConfig)
    wechatCorpTpAuthTokenLoad(globalConfig)
    toutiaoAppTokenLoad(globalConfig)
    fengniaoAppTokenLoad(globalConfig)
}

func fixedConfig(config *Config) {
    gokits.If(0 == config.Port, func() {
        config.Port = 4236
    })
    gokits.If("" != config.ContextPath, func() {
        gokits.Unless(strings.HasPrefix(config.ContextPath, "/"),
            func() { config.ContextPath = "/" + config.ContextPath })
        gokits.If(strings.HasSuffix(config.ContextPath, "/"),
            func() { config.ContextPath = config.ContextPath[:len(config.ContextPath)-1] })
    })
    gokits.If("" == config.LogLevel, func() {
        config.LogLevel = "info"
    })

    gokits.GlobalHttpServerConfig = (*gokits.HttpServerConfig)(unsafe.Pointer(config))

    golog.SetLevel(config.LogLevel)
    golog.Infof("config: %+v", *config)
}

func loadSqlxDB(config *Config) *sqlx.DB {
    db, err := sqlx.Open(config.DriverName, config.DataSourceName)
    if err != nil {
        golog.Errorf("open sqlx.DB error: %s", err.Error())
        return nil
    }

    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime.Duration)
    db.SetConnMaxLifetime(config.ConnMaxLifetime.Duration)

    if err = db.Ping(); err != nil {
        golog.Errorf("connect DB error: %s", err.Error())
        return nil
    }

    db.MapperFunc(func(s string) string { return s })
    golog.Infof("DB: %+v", db)
    return db
}

func fetchClusterNodes(config *Config) {
    gokits.If("" != config.ClusterNodeAddresses, func() {
        addrSlice := strings.Split(config.ClusterNodeAddresses, ",")
        for _, addr := range addrSlice {
            address := strings.TrimSpace(addr)
            if strings.HasSuffix(address, "/") {
                address = address[:len(address)-1]
            }
            clusterNodeAddresses = append(clusterNodeAddresses, address)
        }
    })
}

func publishToClusterNodes(consumer func(address string)) {
    if nil == consumer {
        return
    }
    for _, address := range clusterNodeAddresses {
        go consumer(address)
    }
}
