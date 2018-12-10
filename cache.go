package varys

import (
    "github.com/CharLemAznable/gcache"
    . "github.com/CharLemAznable/goutils"
    log "github.com/CharLemAznable/log4go"
    "time"
)

func configLoader(
    name string,
    sql string,
    lifeSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    key interface{},
    args ...interface{}) (*gcache.CacheItem, error) {

    resultMap, err := db.Sql(sql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, &UnexpectedError{Message:
        "Require " + name + " Config with key: " + key.(string)} // require config
    }
    log.Trace("Query %s:(%s) %s", name, key, resultMap)

    config := builder(resultMap[0])
    log.Info("Load %s Cache:(%s) %s", name, key, Json(config))
    return gcache.NewCacheItem(key, lifeSpan, config), nil
}

func tokenLoader(
    name string,
    querySql string,
    createSql string,
    updateSql string,
    uncompleteSql string,
    completeSql string,
    lifeSpan time.Duration,
    lifeSpanTemp time.Duration,
    builder func(resultItem map[string]string) interface{},
    requestor func(key interface{}) (map[string]string, error),
    completeParamBuilder func(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{},
    key interface{},
    args ...interface{}) (*gcache.CacheItem, error) {

    resultMap, err := db.Sql(querySql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        log.Info("Try to request %s:(%s)", name, key)
        count, err := db.Sql(createSql).Params(key).Execute()
        if nil == err && count > 0 {
            tokenItem, err := requestUpdater(name, uncompleteSql, completeSql, lifeSpan,
                builder, requestor, completeParamBuilder, key, args...)
            if nil != err {
                return nil, err
            }
            log.Info("Request %s:(%s) %s", name, key, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem), nil
        }
        log.Warn("Give up request %s:(%s), wait for next cache Query", name, key)
        return nil, &UnexpectedError{Message: "Query " + name + " Later"}
    }
    log.Trace("Query %s:(%s) %s", name, key, resultMap)

    resultItem := resultMap[0]
    updated := resultItem["UPDATED"]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err {
        return nil, err
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        log.Info("Try to request and update %s:(%s)", name, key)
        count, err := db.Sql(updateSql).Params(key).Execute()
        if nil == err && count > 0 {
            tokenItem, err := requestUpdater(name, uncompleteSql, completeSql, lifeSpan,
                builder, requestor, completeParamBuilder, key, args...)
            if nil != err {
                return nil, err
            }
            log.Info("Request %s:(%s) %s", name, key, Json(tokenItem))
            return gcache.NewCacheItem(key, lifeSpan, tokenItem), nil
        }
        log.Warn("Give up request and update %s:(%s), use Query result Temporarily", name, key)
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := Condition(isExpired, lifeSpanTemp, lifeSpan).(time.Duration)
    tokenItem := builder(resultItem)
    log.Info("Load %s Cache:(%s) %s, cache %3.1f min", name, key, Json(tokenItem), ls.Minutes())
    return gcache.NewCacheItem(key, ls, tokenItem), nil
}

func requestUpdater(
    name string,
    uncompleteSql string,
    completeSql string,
    lifeSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    requestor func(key interface{}) (map[string]string, error),
    completeParamBuilder func(resultItem map[string]string, lifeSpan time.Duration, key interface{}) []interface{},
    key interface{},
    args ...interface{}) (interface{}, error) {

    resultItem, err := requestor(key)
    if nil != err {
        db.Sql(uncompleteSql).Params(key).Execute()
        return nil, err
    }
    // // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    // expireTimeInc := expiresIn - int(lifeSpan.Seconds()*1.1)
    // count, err := db.Sql(completeSql).Params(token, expireTimeInc, key).Execute()
    count, err := db.Sql(completeSql).Params(completeParamBuilder(resultItem, lifeSpan, key)...).Execute()
    if nil != err {
        return nil, err
    }
    if count < 1 {
        return nil, &UnexpectedError{Message: "Record new " + name + " Failed"}
    }

    return builder(resultItem), nil
}
