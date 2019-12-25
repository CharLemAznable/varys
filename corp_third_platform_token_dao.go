package main

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
