package varys

func UpdateWechatThirdPlatformTicket(appId, ticket string) (int64, error) {
    count, err := db.Sql(replaceWechatThirdPlatformTicketSQL).Params(appId, ticket).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func QueryWechatThirdPlatformTicket(appId string) (string, error) {
    resultMap, err := db.Sql(queryWechatThirdPlatformTicketSQL).Params(appId).Query()
    if nil != err {
        return "", err
    }
    if 1 != len(resultMap) {
        return "", &UnexpectedError{Message: "Query WechatThirdPlatform Ticket Failed"}
    }
    return resultMap[0]["TICKET"], nil
}

func EnableWechatThirdPlatformAuthorizer(
    appId, authorizerAppid, authorizationCode, preAuthCode string) (int64, error) {

    count, err := db.Sql(enableWechatThirdPlatformAuthorizerSQL).
        Params(appId, authorizerAppid, authorizationCode, preAuthCode).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func DisableWechatThirdPlatformAuthorizer(appId, authorizerAppid string) (int64, error) {

    count, err := db.Sql(disableWechatThirdPlatformAuthorizerSQL).
        Params(appId, authorizerAppid).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}
