package varys

func updateWechatAppThirdPlatformTicket(codeName, ticket string) (int64, error) {
    count, err := db.New().Sql(updateWechatAppThirdPlatformTicketSQL).Params(ticket, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func queryWechatAppThirdPlatformTicket(codeName string) (string, error) {
    resultMap, err := db.New().Sql(queryWechatAppThirdPlatformTicketSQL).Params(codeName).Query()
    if nil != err {
        return "", err
    }
    if 1 != len(resultMap) {
        return "", &UnexpectedError{Message: "Query WechatAppThirdPlatformTicket Failed"}
    }
    return resultMap[0]["TICKET"], nil
}

func enableWechatAppThirdPlatformAuthorizer(
    codeName, authorizerAppid, authorizationCode, preAuthCode string) (int64, error) {

    count, err := db.New().Sql(enableWechatAppThirdPlatformAuthorizerSQL).
        Params(authorizerAppid, authorizationCode, preAuthCode, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func disableWechatAppThirdPlatformAuthorizer(
    codeName, authorizerAppid string) (int64, error) {

    count, err := db.New().Sql(disableWechatAppThirdPlatformAuthorizerSQL).
        Params(authorizerAppid, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func updateWechatCorpThirdPlatformTicket(codeName, ticket string) (int64, error) {
    count, err := db.New().Sql(replaceWechatCorpThirdPlatformTicketSQL).Params(ticket, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func queryWechatCorpThirdPlatformTicket(codeName string) (string, error) {
    resultMap, err := db.New().Sql(queryWechatCorpThirdPlatformTicketSQL).Params(codeName).Query()
    if nil != err {
        return "", err
    }
    if 1 != len(resultMap) {
        return "", &UnexpectedError{Message: "Query WechatCorpThirdPlatform Ticket Failed"}
    }
    return resultMap[0]["TICKET"], nil
}

func enableWechatCorpThirdPlatformAuthorizer(
    codeName, corpId, permanentCode string) (int64, error) {

    count, err := db.New().Sql(enableWechatCorpThirdPlatformAuthorizerSQL).
        Params(corpId, permanentCode, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func disableWechatCorpThirdPlatformAuthorizer(
    codeName, corpId string) (int64, error) {

    count, err := db.New().Sql(disableWechatCorpThirdPlatformAuthorizerSQL).
        Params(corpId, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}
