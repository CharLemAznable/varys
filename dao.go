package main

type WechatAPITokenConfig struct {
    AppId     string
    AppSecret string
}

type WechatAPIToken struct {
    AppId       string
    AccessToken string
}

func GetWechatAPITokenConfig(appId string) (*WechatAPITokenConfig, error) {
    config, err := wechatAPITokenConfigCache.Value(appId)
    if nil != err {
        return nil, err
    }
    return config.Data().(*WechatAPITokenConfig), nil
}

func GetWechatAPIToken(appId string) (*WechatAPIToken, error) {
    token, err := wechatAPITokenCache.Value(appId)
    if nil != err {
        return nil, err
    }
    return token.Data().(*WechatAPIToken), nil
}
