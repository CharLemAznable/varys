package main

func enableWechatTpAuth(codeName, authorizerAppid, authorizationCode, preAuthCode string) (int64, error) {
    count, err := db.New().Sql(enableWechatTpAuthSQL).
        Params(authorizerAppid, authorizationCode, preAuthCode, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func disableWechatTpAuth(codeName, authorizerAppid string) (int64, error) {
    count, err := db.New().Sql(disableWechatTpAuthSQL).
        Params(authorizerAppid, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}
