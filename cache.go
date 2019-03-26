package varys

import (
    "fmt"
    . "github.com/CharLemAznable/gokits"
    "time"
)

func configLoader(
    name string,
    sql string,
    lifeSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    key interface{},
    args ...interface{}) (*CacheItem, error) {

    resultMap, err := db.New().Sql(sql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        return nil, &UnexpectedError{Message:
        "Require " + name + " Config with key: " + key.(string)} // require config
    }
    LOG.Trace("Query %s:(%s) %s", name, key, resultMap)

    config := builder(resultMap[0])
    LOG.Info("Load %s Cache:(%s) %s", name, key, Json(config))
    return NewCacheItem(key, lifeSpan, config), nil
}

// 针对具有刷新过渡期的token
// 在旧的token即将过期时, 请求获取新的token
// 在此过程中新旧token同时可用
// 所以在某一集群节点更新token时
// 其他节点可临时缓存旧的token保证平滑过渡
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
    args ...interface{}) (*CacheItem, error) {

    resultMap, err := db.New().Sql(querySql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        LOG.Info("Try to request %s:(%s)", name, key)
        count, err := db.New().Sql(createSql).Params(key).Execute()
        if nil == err && count > 0 {
            tokenItem, err := requestUpdater(name, uncompleteSql, completeSql, lifeSpan,
                builder, requestor, completeParamBuilder, key, args...)
            if nil != err {
                return nil, err
            }
            LOG.Info("Request %s:(%s) %s", name, key, Json(tokenItem))
            return NewCacheItem(key, lifeSpan, tokenItem), nil
        }
        _ = LOG.Warn("Give up request %s:(%s), wait for next cache Query", name, key)
        return nil, &UnexpectedError{Message: "Query " + name + " Later"}
    }
    LOG.Trace("Query %s:(%s) %s", name, key, resultMap)

    resultItem := resultMap[0]
    updated := resultItem["UPDATED"]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"])
    if nil != err {
        return nil, err
    }
    isExpired := time.Now().Unix() > expireTime
    isUpdated := "1" == updated
    if isExpired && isUpdated { // 已过期 && 是最新记录 -> 触发更新
        LOG.Info("Try to request and update %s:(%s)", name, key)
        count, err := db.New().Sql(updateSql).Params(key).Execute()
        if nil == err && count > 0 {
            tokenItem, err := requestUpdater(name, uncompleteSql, completeSql, lifeSpan,
                builder, requestor, completeParamBuilder, key, args...)
            if nil != err {
                return nil, err
            }
            LOG.Info("Request and Update %s:(%s) %s", name, key, Json(tokenItem))
            return NewCacheItem(key, lifeSpan, tokenItem), nil
        }
        _ = LOG.Warn("Give up request and update %s:(%s), use Query result Temporarily", name, key)
    }

    // 未过期 || (已非最新记录 || 更新为旧记录失败)
    // 未过期 -> 正常缓存查询到的token
    // (已非最新记录 || 更新为旧记录失败) 表示已有其他集群节点开始了更新操作
    // 已过期(已开始更新) -> 临时缓存查询到的token
    ls := Condition(isExpired, lifeSpanTemp, lifeSpan).(time.Duration)
    tokenItem := builder(resultItem)
    LOG.Info("Load %s Cache:(%s) %s, cache %3.1f min", name, key, Json(tokenItem), ls.Minutes())
    return NewCacheItem(key, ls, tokenItem), nil
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
        _, _ = db.New().Sql(uncompleteSql).Params(key).Execute()
        return nil, err
    }
    // // 过期时间增量: token实际有效时长 - token缓存时长 * 缓存提前更新系数(1.1)
    // expireTimeInc := expiresIn - int(lifeSpan.Seconds()*1.1)
    // count, err := db.Sql(completeSql).Params(token, expireTimeInc, key).Execute()
    count, err := db.New().Sql(completeSql).Params(completeParamBuilder(resultItem, lifeSpan, key)...).Execute()
    if nil != err {
        return nil, err
    }
    if count < 1 {
        return nil, &UnexpectedError{Message: "Record new " + name + " Failed"}
    }

    return builder(resultItem), nil
}

// 针对没有刷新过渡期的token
// 旧的token固定过期时间, 在有效期内无法获取新的token
// 超出有效期后, 可请求获取新的token
// 所以在旧的token即将过期前的一小段时间内, 停止缓存token并请求更新
// 在某一集群节点更新token后, 其他节点触发入库失败并查询已更新的token进行缓存
// 由于有效期内token不会更新, 所以重复请求不会覆盖token
func tokenLoaderStrict(
    name string,
    querySql string,
    createSql string,
    updateSql string,
    maxLifeSpan time.Duration,
    expireCriticalSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    requestor func(key interface{}) (map[string]string, error),
    sqlParamBuilder func(resultItem map[string]string, key interface{}) []interface{},
    key interface{},
    args ...interface{}) (*CacheItem, error) {

    resultMap, err := db.New().Sql(querySql).Params(key).Query()
    if nil != err || 1 != len(resultMap) {
        LOG.Info("Try to request %s:(%s)", name, key)

        tokenItem, err := requestUpdaterStrict(name, querySql, createSql,
            expireCriticalSpan, builder, requestor, sqlParamBuilder, key, args...)
        if nil != err {
            return nil, err
        }
        // 请求成功, 缓存最大缓存时长
        LOG.Info("Request %s:(%s) %s", name, key, Json(tokenItem))
        return NewCacheItem(key, maxLifeSpan, tokenItem), nil
    }
    LOG.Trace("Query %s:(%s) %s", name, key, resultMap)

    resultItem := resultMap[0]
    expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
    if nil != err {
        return nil, err
    }
    effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
    // 即将过期 -> 触发更新
    if effectiveSpan <= expireCriticalSpan {
        time.Sleep(expireCriticalSpan) // 休眠后再请求最新的access_token
        LOG.Info("Try to request and update %s:(%s)", name, key)

        tokenItem, err := requestUpdaterStrict(name, querySql, updateSql,
            expireCriticalSpan, builder, requestor, sqlParamBuilder, key, args...)
        if nil != err {
            return nil, err
        }
        // 请求更新成功, 缓存最大缓存时长
        LOG.Info("Request and Update %s:(%s) %s", name, key, Json(tokenItem))
        return NewCacheItem(key, maxLifeSpan, tokenItem), nil
    }

    // token有效期少于最大缓存时长, 则仅缓存剩余有效期时长
    ls := Condition(effectiveSpan > maxLifeSpan, maxLifeSpan, effectiveSpan).(time.Duration)
    tokenItem := builder(resultItem)
    LOG.Info("Load %s Cache:(%s) %s, cache %3.1f min", name, key, Json(tokenItem), ls.Minutes())
    return NewCacheItem(key, ls, tokenItem), nil
}

func requestUpdaterStrict(
    name string,
    querySql string,
    completeSql string,
    expireCriticalSpan time.Duration,
    builder func(resultItem map[string]string) interface{},
    requestor func(key interface{}) (map[string]string, error),
    sqlParamBuilder func(resultItem map[string]string, key interface{}) []interface{},
    key interface{},
    args ...interface{}) (interface{}, error) {

    resultItem, err := requestor(key)
    if nil != err {
        _ = LOG.Warn("Request %s Failed:(%s) %s", name, key, err.Error())
        return nil, err
    }

    count, err := db.New().Sql(completeSql).Params(sqlParamBuilder(resultItem, key)...).Execute()
    if nil != err || count < 1 { // 记录入库失败, 则查询记录并返回
        resultMap, err := db.New().Sql(querySql).Params(key).Query()
        if nil != err || 1 != len(resultMap) {
            return nil, DefaultIfNil(err, &UnexpectedError{
                Message: fmt.Sprintf("Query %s:(%s) Failed", name, key)}).(error)
        }

        resultItem := resultMap[0]
        expireTime, err := Int64FromStr(resultItem["EXPIRE_TIME"]) // in second
        if nil != err {
            return nil, err
        }
        effectiveSpan := time.Duration(expireTime-time.Now().Unix()) * time.Second
        if effectiveSpan <= expireCriticalSpan {
            return nil, &UnexpectedError{Message:
                fmt.Sprintf("Query %s:(%s) expireTime Failed", name, key)}
        }

        return builder(resultItem), nil
    }

    return builder(resultItem), nil
}
