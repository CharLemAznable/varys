package main

import (
	"time"
)

var wechatCorpThirdPlatformTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_suite_token"
var wechatCorpThirdPlatformPreAuthCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_pre_auth_code?suite_access_token="
var wechatCorpThirdPlatformPermanentCodeURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_permanent_code?suite_access_token="
var wechatCorpThirdPlatformCorpTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/service/get_corp_token?suite_access_token="

var wechatCorpThirdPlatformConfigLifeSpan = time.Minute * 60             // config cache 60 min default
var wechatCorpThirdPlatformCryptorLifeSpan = time.Minute * 60            // cryptor cache 60 min default
var wechatCorpThirdPlatformTokenMaxLifeSpan = time.Minute * 5            // stable token cache 5 min max
var wechatCorpThirdPlatformTokenExpireCriticalSpan = time.Second * 1     // token about to expire critical time span
var wechatCorpThirdPlatformPermanentCodeLifeSpan = time.Minute * 60      // permanent_code cache 60 min default
var wechatCorpThirdPlatformCorpTokenMaxLifeSpan = time.Minute * 5        // stable token cache 5 min max
var wechatCorpThirdPlatformCorpTokenExpireCriticalSpan = time.Second * 1 // token about to expire critical time span

func wechatCorpThirdPlatformAuthorizerTokenLoad(configMap map[string]string) {
	urlConfigLoader(configMap["wechatCorpThirdPlatformTokenURL"],
		func(configURL string) {
			wechatCorpThirdPlatformTokenURL = configURL
		})
	urlConfigLoader(configMap["wechatCorpThirdPlatformPreAuthCodeURL"],
		func(configURL string) {
			wechatCorpThirdPlatformPreAuthCodeURL = configURL
		})
	urlConfigLoader(configMap["wechatCorpThirdPlatformPermanentCodeURL"],
		func(configURL string) {
			wechatCorpThirdPlatformPermanentCodeURL = configURL
		})
	urlConfigLoader(configMap["wechatCorpThirdPlatformCorpTokenURL"],
		func(configURL string) {
			wechatCorpThirdPlatformCorpTokenURL = configURL
		})

	lifeSpanConfigLoader(
		configMap["wechatCorpThirdPlatformConfigLifeSpan"],
		func(configVal time.Duration) {
			wechatCorpThirdPlatformConfigLifeSpan = configVal * time.Minute
		})
	lifeSpanConfigLoader(
		configMap["wechatCorpThirdPlatformCryptorLifeSpan"],
		func(configVal time.Duration) {
			wechatCorpThirdPlatformCryptorLifeSpan = configVal * time.Minute
		})
	lifeSpanConfigLoader(
		configMap["wechatCorpThirdPlatformTokenMaxLifeSpan"],
		func(configVal time.Duration) {
			wechatCorpThirdPlatformTokenMaxLifeSpan = configVal * time.Minute
		})
	lifeSpanConfigLoader(
		configMap["wechatCorpThirdPlatformTokenExpireCriticalSpan"],
		func(configVal time.Duration) {
			wechatCorpThirdPlatformTokenExpireCriticalSpan = configVal * time.Second
		})
	lifeSpanConfigLoader(
		configMap["wechatCorpThirdPlatformPermanentCodeLifeSpan"],
		func(configVal time.Duration) {
			wechatCorpThirdPlatformPermanentCodeLifeSpan = configVal * time.Minute
		})
	lifeSpanConfigLoader(
		configMap["wechatCorpThirdPlatformCorpTokenMaxLifeSpan"],
		func(configVal time.Duration) {
			wechatCorpThirdPlatformCorpTokenMaxLifeSpan = configVal * time.Minute
		})
	lifeSpanConfigLoader(
		configMap["wechatCorpThirdPlatformCorpTokenExpireCriticalSpan"],
		func(configVal time.Duration) {
			wechatCorpThirdPlatformCorpTokenExpireCriticalSpan = configVal * time.Second
		})

	wechatCorpThirdPlatformAuthorizerTokenInitialize()
}
