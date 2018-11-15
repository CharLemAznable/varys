package gql

import (
    "database/sql"
    "github.com/CharLemAznable/yaml"
    "log"
    "time"
)

type Config struct {
    DriverName      string
    DataSourceName  string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime int // Seconds
}

var connMap = make(map[string]*sql.DB)

func Connection(name string) *sql.DB {
    return connMap[name]
}

func LoadConfigFile(filename string) {
    configFile, err := yaml.ReadFile(filename)
    if nil != err {
        log.Println(err)
        return
    }

    loadConfigYAML(configFile)
}

func LoadConfigString(yamlconf string) {
    configFile, err := yaml.ReadString(yamlconf)
    if nil != err {
        log.Println(err)
        return
    }

    loadConfigYAML(configFile)
}

func loadConfigYAML(file *yaml.File) {
    configMap, err := yaml.MapOf(file.Root, "root")
    if nil != err {
        log.Println(err)
        return
    }

    for name, node := range configMap {
        configItemMap, err := yaml.MapOf(node, name)
        if nil != err {
            log.Println(err)
            continue
        }

        driverName, err := yaml.StringOf(configItemMap["DriverName"], name+".DriverName")
        if nil != err {
            log.Println(err)
            continue
        }
        dataSourceName, err := yaml.StringOf(configItemMap["DataSourceName"], name+".DataSourceName")
        if nil != err {
            log.Println(err)
            continue
        }

        db, err := sql.Open(driverName, dataSourceName)
        if nil != err {
            log.Println(err)
            continue
        }

        maxOpenConns, err := yaml.IntOf(configItemMap["MaxOpenConns"], name+".MaxOpenConns")
        if nil == err {
            db.SetMaxOpenConns(int(maxOpenConns))
        }
        maxIdleConns, err := yaml.IntOf(configItemMap["MaxIdleConns"], name+".MaxIdleConns")
        if nil == err {
            db.SetMaxIdleConns(int(maxIdleConns))
        }
        connMaxLifetime, err := yaml.IntOf(configItemMap["ConnMaxLifetime"], name+".ConnMaxLifetime")
        if nil == err {
            db.SetConnMaxLifetime(time.Second * time.Duration(connMaxLifetime))
        }

        connMap[name] = db
    }
}
