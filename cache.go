package main

import (
    "errors"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "time"
)

type ExpireTimeRecord interface {
    GetExpireTime() int64
}

type UpdatedRecord interface {
    ExpireTimeRecord
    GetUpdated() string
}

func configLoader(
    name string,
    config interface{},
    sql string,
    lifeSpan time.Duration,
    key interface{},
    args ...interface{}) (*gokits.CacheItem, error) {

    err := db.NamedGet(config, sql, map[string]interface{}{"CodeName": key})
    if nil != err {
        return nil, errors.New("Require " + name + " Config with key: " + key.(string))
    }
    golog.Infof("Load %s Cache:(%s) %+v", name, key, config)
    return gokits.NewCacheItem(key, lifeSpan, config), nil
}

// 针对具有刷新过渡期的token
// 在旧的token即将过期时, 请求获取新的token
// 在此过程中新旧token同时可用
// 所以在某一集群节点更新token时
// 其他节点可临时缓存旧的token保证平滑过渡
func tokenLoader(
    name string,
    queryDest UpdatedRecord,
    querySql string,
    queryTokenBuilder func(queryDest UpdatedRecord) interface{},
    createSql string,
    updateSql string,
    requestor func(key interface{}) (map[string]string, error),
    uncompleteSql string,
    completeSql string,
    completeArg func(response map[string]string, lifeSpan time.Duration) map[string]interface{},
    requestTokenBuilder func(response map[string]string) interface{},
    lifeSpan time.Duration,
    lifeSpanTemp time.Duration,
    key interface{},
    args ...interface{}) (*gokits.CacheItem, error) {

    err := db.NamedGet(queryDest, querySql, map[string]interface{}{"CodeName": key})
    if nil != err {
        golog.Debugf("Try to request %s:(%s)", name, key)
        count, err := db.NamedExecX(createSql, map[string]interface{}{"CodeName": key})
        if nil == err && count > 0 {
            token, err := requestUpdater(name, requestor, uncompleteSql,
                completeSql, completeArg, requestTokenBuilder, lifeSpan, key, args...)
            if nil != err {
                return nil, err
            }
            golog.Infof("Request %s:(%s) %+v", name, key, token)
            return gokits.NewCacheItem(key, lifeSpan, token), nil
        }
        golog.Warnf("Give up request %s:(%s), wait for next cache Query", name, key)
        return nil, errors.New("Query " + name + " Later")
    }
    golog.Debugf("Query %s:(%s) %+v", name, key, queryDest)

    isExpired := time.Now().Unix() > queryDest.GetExpireTime()
    isUpdated := "1" == queryDest.GetUpdated()
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        golog.Debugf("Try to request and update %s:(%s)", name, key)
        count, err := db.NamedExecX(updateSql, map[string]interface{}{"CodeName": key})
        if nil == err && count > 0 {
            token, err := requestUpdater(name, requestor, uncompleteSql,
                completeSql, completeArg, requestTokenBuilder, lifeSpan, key, args...)
            if nil != err {
                return nil, err
            }
            golog.Infof("Request and Update %s:(%s) %+v", name, key, token)
            return gokits.NewCacheItem(key, lifeSpan, token), nil
        }
        golog.Warnf("Give up request and update %s:(%s), use Query result Temporarily", name, key)
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := gokits.Condition(isExpired, lifeSpanTemp, lifeSpan).(time.Duration)
    token := queryTokenBuilder(queryDest)
    golog.Infof("Load %s Cache:(%s) %+v, cache %3.1f min", name, key, token, ls.Minutes())
    return gokits.NewCacheItem(key, ls, token), nil
}

func requestUpdater(
    name string,
    requestor func(key interface{}) (map[string]string, error),
    uncompleteSql string,
    completeSql string,
    completeArg func(response map[string]string, lifeSpan time.Duration) map[string]interface{},
    requestTokenBuilder func(response map[string]string) interface{},
    lifeSpan time.Duration,
    key interface{},
    args ...interface{}) (interface{}, error) {

    response, err := requestor(key)
    if nil != err {
        _, _ = db.NamedExec(uncompleteSql, map[string]interface{}{"CodeName": key})
        return nil, err
    }
    // // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    // expireTimeInc := expiresIn - int(lifeSpan.Seconds()*1.1)
    // count, err := db.Sql(completeSql).Params(token, expireTimeInc, key).Execute()
    arg := completeArg(response, lifeSpan)
    arg["CodeName"] = key
    _, err = db.NamedExec(completeSql, arg)
    if nil != err {
        return nil, err
    }

    return requestTokenBuilder(response), nil
}

// 针对没有刷新过渡期的token
// 旧的token固定过期时间, 在有效期内无法获取新的token
// 超出有效期后, 可请求获取新的token
// 所以在旧的token即将过期前的一小段时间内, 停止缓存token并请求更新
// 在某一集群节点更新token后, 其他节点触发入库失败并查询已更新的token进行缓存
// 由于有效期内token不会更新, 所以重复请求不会覆盖token
func tokenLoaderStrict(
    name string,
    queryDest ExpireTimeRecord,
    querySql string,
    queryTokenBuilder func(queryDest ExpireTimeRecord) interface{},
    requestor func(key interface{}) (map[string]string, error),
    createSql string,
    updateSql string,
    createUpdateArg func(response map[string]string) map[string]interface{},
    expireCriticalSpan time.Duration,
    requestTokenBuilder func(response map[string]string) interface{},
    maxLifeSpan time.Duration,
    key interface{},
    args ...interface{}) (*gokits.CacheItem, error) {

    err := db.NamedGet(queryDest, querySql, map[string]interface{}{"CodeName": key})
    if nil != err {
        golog.Debugf("Try to request %s:(%s)", name, key)
        token, err := requestUpdaterStrict(name, requestor, createSql, createUpdateArg,
            queryDest, querySql, queryTokenBuilder, expireCriticalSpan, requestTokenBuilder, key, args...)
        if nil != err {
            return nil, err
        }
        // 请求成功, 缓存最大缓存时长
        golog.Infof("Request %s:(%s) %+v", name, key, token)
        return gokits.NewCacheItem(key, maxLifeSpan, token), nil
    }
    golog.Debugf("Query %s:(%s) %+v", name, key, queryDest)

    effectiveSpan := time.Duration(queryDest.GetExpireTime()-time.Now().Unix()) * time.Second // in second
    // 即将过期 -> 触发更新
    if effectiveSpan <= expireCriticalSpan {
        time.Sleep(expireCriticalSpan) // 休眠后再请求最新的access_token
        golog.Debugf("Try to request and update %s:(%s)", name, key)
        token, err := requestUpdaterStrict(name, requestor, updateSql, createUpdateArg,
            queryDest, querySql, queryTokenBuilder, expireCriticalSpan, requestTokenBuilder, key, args...)
        if nil != err {
            return nil, err
        }
        // 请求更新成功, 缓存最大缓存时长
        golog.Infof("Request and Update %s:(%s) %+v", name, key, token)
        return gokits.NewCacheItem(key, maxLifeSpan, token), nil
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长
    ls := gokits.Condition(effectiveSpan > maxLifeSpan, maxLifeSpan, effectiveSpan).(time.Duration)
    token := queryTokenBuilder(queryDest)
    golog.Infof("Load %s Cache:(%s) %+v, cache %3.1f min", name, key, token, ls.Minutes())
    return gokits.NewCacheItem(key, ls, token), nil
}

func requestUpdaterStrict(
    name string,
    requestor func(key interface{}) (map[string]string, error),
    writeSql string,
    writeArg func(response map[string]string) map[string]interface{},
    queryDest ExpireTimeRecord,
    querySql string,
    queryTokenBuilder func(queryDest ExpireTimeRecord) interface{},
    expireCriticalSpan time.Duration,
    requestTokenBuilder func(response map[string]string) interface{},
    key interface{},
    args ...interface{}) (interface{}, error) {

    response, err := requestor(key)
    if nil != err {
        golog.Warnf("Request %s Failed:(%s) %s", name, key, err.Error())
        return nil, err
    }
    arg := writeArg(response)
    arg["CodeName"] = key
    count, err := db.NamedExecX(writeSql, arg)
    if nil != err || count < 1 { // 记录入库失败, 则查询记录并返回
        err := db.NamedGet(queryDest, querySql, map[string]interface{}{"CodeName": key})
        if nil != err {
            return nil, err
        }

        effectiveSpan := time.Duration(queryDest.GetExpireTime()-time.Now().Unix()) * time.Second // in second
        if effectiveSpan <= expireCriticalSpan {
            return nil, errors.New(fmt.Sprintf("Query %s:(%s) expireTime Failed", name, key))
        }
        return queryTokenBuilder(queryDest), nil
    }
    return requestTokenBuilder(response), nil
}
