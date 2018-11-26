package varys

func UpdateWechatThirdPlatformTicket(codeName, ticket string) (int64, error) {
    count, err := db.Sql(replaceWechatThirdPlatformTicketSQL).Params(ticket, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func QueryWechatThirdPlatformTicket(codeName string) (string, error) {
    resultMap, err := db.Sql(queryWechatThirdPlatformTicketSQL).Params(codeName).Query()
    if nil != err {
        return "", err
    }
    if 1 != len(resultMap) {
        return "", &UnexpectedError{Message: "Query WechatThirdPlatform Ticket Failed"}
    }
    return resultMap[0]["TICKET"], nil
}

func EnableWechatThirdPlatformAuthorizer(
    codeName, authorizerAppid, authorizationCode, preAuthCode string) (int64, error) {

    count, err := db.Sql(enableWechatThirdPlatformAuthorizerSQL).
        Params(authorizerAppid, authorizationCode, preAuthCode, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func DisableWechatThirdPlatformAuthorizer(
    codeName, authorizerAppid string) (int64, error) {

    count, err := db.Sql(disableWechatThirdPlatformAuthorizerSQL).
        Params(authorizerAppid, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}
