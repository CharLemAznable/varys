package base

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

    ClusterNodeAddresses []string // notify all cluster node for cache delete event

    DriverName      string
    DataSourceName  string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxIdleTime gokits.Duration
    ConnMaxLifetime gokits.Duration
}

var DB *sqlx.DB

var configFile = ""
var config = &Config{}

func init() {
    testing.Init()
    flag.StringVar(&configFile, "configFile",
        "config.toml", "config file path")
    flag.Parse()
    LoadConfig(configFile, config)
    fixConfig()
}

func LoadConfig(configFile string, config interface{}) {
    if _, err := toml.DecodeFile(configFile, config); err != nil {
        golog.Errorf("config file decode error: %s", err.Error())
    }
}

func fixConfig() {
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
    clusterNodeAddresses := make([]string, 0)
    for _, addr := range config.ClusterNodeAddresses {
        address := strings.TrimSpace(addr)
        if strings.HasSuffix(address, "/") {
            address = address[:len(address)-1]
        }
        clusterNodeAddresses = append(clusterNodeAddresses, address)
    }
    config.ClusterNodeAddresses = clusterNodeAddresses

    gokits.GlobalHttpServerConfig = (*gokits.HttpServerConfig)(unsafe.Pointer(config))

    golog.SetLevel(config.LogLevel)
    golog.Infof("config: %+v", *config)
}

func InitSqlxDB() {
    _db, err := sqlx.Open(config.DriverName, config.DataSourceName)
    if err != nil {
        golog.Errorf("open sqlx.DB error: %s", err.Error())
        return
    }

    _db.SetMaxOpenConns(config.MaxOpenConns)
    _db.SetMaxIdleConns(config.MaxIdleConns)
    _db.SetConnMaxIdleTime(config.ConnMaxIdleTime.Duration)
    _db.SetConnMaxLifetime(config.ConnMaxLifetime.Duration)

    if err = _db.Ping(); err != nil {
        golog.Errorf("connect DB error: %s", err.Error())
        return
    }

    _db.MapperFunc(func(s string) string { return s })
    golog.Infof("DB: %+v", _db)
    DB = _db
}

func ServerAddr() string {
    return ":" + gokits.StrFromInt(config.Port)
}

func PublishToClusterNodes(consumer func(address string)) {
    if nil == consumer {
        return
    }
    for _, address := range config.ClusterNodeAddresses {
        go consumer(address)
    }
}
