package httpreq

import (
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "strings"
)

type HttpReq struct {
    baseUrl string
    req     string
    params  map[string]string
    props   []prop
}

type prop struct {
    name  string
    value string
}

func New(baseUrl string) *HttpReq {
    httpReq := new(HttpReq)
    httpReq.baseUrl = baseUrl
    httpReq.params = make(map[string]string)
    httpReq.props = make([]prop, 0)
    return httpReq
}

func (httpReq *HttpReq) Req(req string) *HttpReq {
    httpReq.req = req
    return httpReq
}

func (httpReq *HttpReq) Prop(name string, value string) *HttpReq {
    if 0 == len(name) || 0 == len(value) {
        return httpReq
    }
    httpReq.props = append(httpReq.props, prop{name: name, value: value})
    return httpReq
}

func (httpReq *HttpReq) Cookie(value string) *HttpReq {
    if 0 == len(value) {
        return httpReq
    }
    return httpReq.Prop("Cookie", value)
}

func (httpReq *HttpReq) Params(name string, value string, more ... string) *HttpReq {
    if 0 != len(name) || 0 != len(value) {
        httpReq.params[name] = value
    }

    for i := 0; i < len(more); i += 2 {
        if i+1 >= len(more) {
            break
        }

        k, v := more[i], more[i+1]
        if 0 != len(k) || 0 != len(v) {
            httpReq.params[k] = v
        }
    }

    return httpReq
}

func (httpReq *HttpReq) ParamsMapping(params map[string]string) *HttpReq {
    if nil == params {
        return httpReq
    }

    for key, value := range params {
        if 0 != len(key) || 0 != len(value) {
            httpReq.params[key] = value
        }
    }
    return httpReq
}

func (httpReq *HttpReq) Get() (string, error) {
    request, err := httpReq.createGetRequest()
    if nil != err {
        return "", err
    }
    httpReq.commonSettings(request)
    httpReq.setHeaders(request)

    response, err := http.DefaultClient.Do(request)
    defer response.Body.Close()
    if nil != err {
        log.Printf("Get: %s, STATUS CODE = %d\n\n%s\n",
            request.URL.String(), response.StatusCode, err.Error())
        return "", err
    }

    body, err := ioutil.ReadAll(response.Body)
    if nil != err {
        return "", err
    }
    return string(body), nil
}

func (httpReq *HttpReq) Post() (string, error) {
    request, err := httpReq.createPostRequest()
    if nil != err {
        return "", err
    }
    httpReq.commonSettings(request)
    httpReq.postSettings(request)
    httpReq.setHeaders(request)

    response, err := http.DefaultClient.Do(request)
    defer response.Body.Close()
    if nil != err {
        log.Printf("Post: %s, STATUS CODE = %d\n\n%s\n",
            request.URL.String(), response.StatusCode, err.Error())
        return "", err
    }

    body, err := ioutil.ReadAll(response.Body)
    if nil != err {
        return "", err
    }
    return string(body), nil
}

func (httpReq *HttpReq) createGetRequest() (*http.Request, error) {
    values := url.Values{}
    for key, value := range httpReq.params {
        values.Add(key, value)
    }
    encoded := values.Encode()

    urlStr := httpReq.baseUrl + httpReq.req
    if len(encoded) > 0 {
        urlStr = urlStr + "?" + encoded
    }
    return http.NewRequest("GET", urlStr, nil)
}

func (httpReq *HttpReq) createPostRequest() (*http.Request, error) {
    values := url.Values{}
    for key, value := range httpReq.params {
        values.Add(key, value)
    }
    encoded := values.Encode()

    urlStr := httpReq.baseUrl + httpReq.req
    bodyReader := strings.NewReader(encoded)
    return http.NewRequest("POST", urlStr, bodyReader)
}

func (httpReq *HttpReq) commonSettings(request *http.Request) {
    request.Header.Set("Accept-Charset", "UTF-8")
}

func (httpReq *HttpReq) setHeaders(request *http.Request) {
    for _, prop := range httpReq.props {
        request.Header.Set(prop.name, prop.value)
    }
}

func (httpReq *HttpReq) postSettings(request *http.Request) {
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}
