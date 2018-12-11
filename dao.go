package varys

func updateWechatThirdPlatformTicket(codeName, ticket string) (int64, error) {
    count, err := db.Sql(replaceWechatThirdPlatformTicketSQL).Params(ticket, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func queryWechatThirdPlatformTicket(codeName string) (string, error) {
    resultMap, err := db.Sql(queryWechatThirdPlatformTicketSQL).Params(codeName).Query()
    if nil != err {
        return "", err
    }
    if 1 != len(resultMap) {
        return "", &UnexpectedError{Message: "Query WechatThirdPlatform Ticket Failed"}
    }
    return resultMap[0]["TICKET"], nil
}

func enableWechatThirdPlatformAuthorizer(
    codeName, authorizerAppid, authorizationCode, preAuthCode string) (int64, error) {

    count, err := db.Sql(enableWechatThirdPlatformAuthorizerSQL).
        Params(authorizerAppid, authorizationCode, preAuthCode, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func disableWechatThirdPlatformAuthorizer(
    codeName, authorizerAppid string) (int64, error) {

    count, err := db.Sql(disableWechatThirdPlatformAuthorizerSQL).
        Params(authorizerAppid, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func updateWechatCorpThirdPlatformTicket(codeName, ticket string) (int64, error) {
    count, err := db.Sql(replaceWechatCorpThirdPlatformTicketSQL).Params(ticket, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func queryWechatCorpThirdPlatformTicket(codeName string) (string, error) {
    resultMap, err := db.Sql(queryWechatCorpThirdPlatformTicketSQL).Params(codeName).Query()
    if nil != err {
        return "", err
    }
    if 1 != len(resultMap) {
        return "", &UnexpectedError{Message: "Query WechatCorpThirdPlatform Ticket Failed"}
    }
    return resultMap[0]["TICKET"], nil
}
