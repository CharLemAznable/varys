package main

import (
	"github.com/CharLemAznable/gokits"
	"time"
)

var wechatCorpConfigCache *gokits.CacheTable
var wechatCorpTokenCache *gokits.CacheTable

func wechatCorpTokenInitialize() {
	wechatCorpConfigCache = gokits.CacheExpireAfterWrite("wechatCorpConfig")
	wechatCorpConfigCache.SetDataLoader(wechatCorpTokenConfigLoader)
	wechatCorpTokenCache = gokits.CacheExpireAfterWrite("wechatCorpToken")
	wechatCorpTokenCache.SetDataLoader(wechatCorpTokenLoader)
}

type WechatCorpConfig struct {
	CorpId     string
	CorpSecret string
}

func wechatCorpTokenConfigLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
	return configLoader(
		"WechatCorpConfig",
		queryWechatCorpConfigSQL,
		wechatCorpConfigLifeSpan,
		func(resultItem map[string]string) interface{} {
			config := new(WechatCorpConfig)
			config.CorpId = resultItem["CORP_ID"]
			config.CorpSecret = resultItem["CORP_SECRET"]
			if 0 == len(config.CorpId) || 0 == len(config.CorpSecret) {
				return nil
			}
			return config
		},
		codeName, args...)
}

type WechatCorpToken struct {
	CorpId      string
	AccessToken string
}

func wechatCorpTokenBuilder(resultItem map[string]string) interface{} {
	tokenItem := new(WechatCorpToken)
	tokenItem.CorpId = resultItem["CORP_ID"]
	tokenItem.AccessToken = resultItem["ACCESS_TOKEN"]
	return tokenItem
}

type WechatCorpTokenResponse struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func wechatCorpTokenRequestor(codeName interface{}) (map[string]string, error) {
	cache, err := wechatCorpConfigCache.Value(codeName)
	if nil != err {
		return nil, err
	}
	config := cache.Data().(*WechatCorpConfig)

	result, err := gokits.NewHttpReq(wechatCorpTokenURL).Params(
		"corpid", config.CorpId, "corpsecret", config.CorpSecret).
		Prop("Content-Type",
			"application/x-www-form-urlencoded").Get()
	gokits.LOG.Trace("Request WechatCorpToken Response:(%s) %s", codeName, result)
	if nil != err {
		return nil, err
	}

	response := gokits.UnJson(result, new(WechatCorpTokenResponse)).(*WechatCorpTokenResponse)
	if nil == response || 0 != response.Errcode || 0 == len(response.AccessToken) {
		return nil, &UnexpectedError{Message: "Request Corp access_token Failed: " + result}
	}

	// 过期时间增量: token实际有效时长
	expireTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
	return map[string]string{
		"CORP_ID":      config.CorpId,
		"ACCESS_TOKEN": response.AccessToken,
		"EXPIRE_TIME":  gokits.StrFromInt64(expireTime)}, nil
}

func wechatCorpTokenSQLParamBuilder(resultItem map[string]string, codeName interface{}) []interface{} {
	expireTime, _ := gokits.Int64FromStr(resultItem["EXPIRE_TIME"])
	return []interface{}{resultItem["ACCESS_TOKEN"], expireTime, codeName}
}

func wechatCorpTokenLoader(codeName interface{}, args ...interface{}) (*gokits.CacheItem, error) {
	return tokenLoaderStrict(
		"WechatCorpToken",
		queryWechatCorpTokenSQL,
		createWechatCorpTokenSQL,
		updateWechatCorpTokenSQL,
		wechatCorpTokenMaxLifeSpan,
		wechatCorpTokenExpireCriticalSpan,
		wechatCorpTokenBuilder,
		wechatCorpTokenRequestor,
		wechatCorpTokenSQLParamBuilder,
		codeName, args...)
}
