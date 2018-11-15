package gql

import (
    "database/sql"
    "log"
)

type Gql struct {
    conn   *sql.DB
    sql    string
    params []interface{}
}

func New(connName string) (*Gql, error) {
    connection := Connection(connName)
    if nil == connection {
        return nil, &UnknownConnectionName{Name: connName}
    }
    gql := new(Gql)
    gql.conn = connection
    return gql, nil
}

func Default() (*Gql, error) {
    return New("Default")
}

func (gql *Gql) Sql(sql string) *Gql {
    gql.sql = sql
    return gql
}

func (gql *Gql) Params(params ... interface{}) *Gql {
    gql.params = params
    return gql
}

func (gql *Gql) Query() ([]map[string]string, error) {
    err := gql.conn.Ping()
    if err != nil {
        log.Println(err)
        return nil, err
    }

    stmt, err := gql.conn.Prepare(gql.sql)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    rows, err := stmt.Query(gql.params...)
    defer rows.Close()
    if err != nil {
        log.Println(err)
        return nil, err
    }

    columns, err := rows.Columns()
    if err != nil {
        log.Println(err)
        return nil, err
    }

    values := make([]sql.RawBytes, len(columns))
    scanArgs := make([]interface{}, len(columns))
    for i := range values {
        scanArgs[i] = &values[i]
    }

    list := make([]map[string]string, 0)
    for rows.Next() {
        record := make(map[string]string)
        // 将行数据保存到record字典
        err = rows.Scan(scanArgs...)
        if err != nil {
            log.Println(err)
            return nil, err
        }
        for i, col := range values {
            if col != nil {
                record[columns[i]] = string(col)
            }
        }
        list = append(list, record)
    }
    return list, nil
}

func (gql *Gql) Execute() (int64, error) {
    err := gql.conn.Ping()
    if err != nil {
        log.Println(err)
        return 0, err
    }

    stmt, err := gql.conn.Prepare(gql.sql)
    if err != nil {
        log.Println(err)
        return 0, err
    }

    result, err := stmt.Exec(gql.params...)
    if err != nil {
        log.Println(err)
        return 0, err
    }

    return result.RowsAffected()
}
