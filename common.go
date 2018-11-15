package varys

import (
    "encoding/json"
    "strconv"
)

func IntFromStr(str string) (int, error) {
    return strconv.Atoi(str)
}

func Int64FromStr(str string) (int64, error) {
    return strconv.ParseInt(str, 10, 64)
}

func StrFromInt(i int) string {
    return strconv.Itoa(i)
}

func StrFromInt64(i int64) string {
    return strconv.FormatInt(i,10)
}

func Json(v interface{}) string {
    bytes, err := json.Marshal(v)
    if nil != err {
        return ""
    }
    return string(bytes)
}

func Condition(cond bool, trueVal, falseVal interface{}) interface{} {
    if cond {
        return trueVal
    }
    return falseVal
}

func ConditionFunc(cond bool, trueFunc, falseFunc func() interface{}) interface{} {
    if cond {
        return trueFunc()
    }
    return falseFunc()
}

func If(cond bool, trueFunc func()) {
    if cond {
        trueFunc()
    }
}

func Unless(cond bool, falseFunc func()) {
    if !cond {
        falseFunc()
    }
}
