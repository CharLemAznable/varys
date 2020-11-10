package main

func updateWechatTpTicket(codeName, ticket string) (int64, error) {
    count, err := db.New().Sql(updateWechatTpTicketSQL).Params(ticket, codeName).Execute()
    if nil != err {
        return 0, err
    }
    return count, nil
}

func queryWechatTpTicket(codeName string) (string, error) {
    resultMap, err := db.New().Sql(queryWechatTpTicketSQL).Params(codeName).Query()
    if nil != err {
        return "", err
    }
    if 1 != len(resultMap) {
        return "", &UnexpectedError{Message: "Query WechatTpTicket Failed"}
    }
    return resultMap[0]["TICKET"], nil
}
