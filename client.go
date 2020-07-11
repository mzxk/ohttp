package ohttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//HTTPStr 1
type HTTPStr struct {
	req    *http.Request
	client http.Client
	err    error
	url    string
}

//HTTP 一个简单的调用http client的方式，第一个参数是地址，第二个参数是？后面的参数
func HTTP(u string, p map[string]interface{}) *HTTPStr {
	c := http.Client{}
	params := url.Values{}

	if p != nil && len(p) > 0 {
		for k, v := range p {
			params.Set(k, fmt.Sprint(v))
		}
	}
	connect := "?"
	if strings.Contains(u, "?") {
		connect = "&"
	}
	urls := u + connect + params.Encode()
	req, err := http.NewRequest("GET", urls, nil)

	return &HTTPStr{req: req, client: c, err: err, url: urls}
}

//HTTPSign 这是一个封装好的进行签名的函数，他会把key和nonce写入请求，并且在header里添加sign
func HTTPSign(u string, p map[string]interface{}, key, value string) *HTTPStr {
	c := http.Client{}
	params := url.Values{}

	if p != nil && len(p) > 0 {
		for k, v := range p {
			params.Set(k, fmt.Sprint(v))
		}
	}
	params.Set("key", key)
	params.Set("nonce", fmt.Sprint(time.Now().Unix()))
	connect := "?"
	if strings.Contains(u, "?") {
		connect = "&"
	}
	urls := u + connect + params.Encode()
	req, err := http.NewRequest("GET", urls, nil)
	t := &HTTPStr{req: req, client: c, err: err, url: urls}
	sign := sha(t.req.URL.RequestURI() + value)
	t.Header(map[string]string{"sign": sign})
	return t
}

//Proxy 给请求加上代理，接受https://xxxx.com:1111的格式,其他的接受不了
func (t *HTTPStr) Proxy(p string) *HTTPStr {
	ii := url.URL{}
	proxy, err := ii.Parse(p)
	t.err = err
	trans := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	t.client.Transport = trans
	return t
}

//Header Write header
func (t *HTTPStr) Header(hd map[string]string) *HTTPStr {
	for k, v := range hd {
		t.req.Header.Set(k, v)
	}
	return t
}

//Auth SetBasicAuth
func (t *HTTPStr) Auth(k, v string) *HTTPStr {
	t.req.SetBasicAuth(k, v)
	return t
}

//Get 结构
func (t *HTTPStr) Get() (*HTTPRespone, error) {
	if t.err != nil {
		return nil, t.err
	}
	rsp, err := t.client.Do(t.req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return &HTTPRespone{URL: t.req.URL.String(), Byte: body}, err
}

//Post POST
//input body[]byte
func (t *HTTPStr) Post(date []byte) (*HTTPRespone, error) {
	if t.err != nil {
		return nil, t.err
	}
	t.req.Method = "POST"
	t.req.Body = ioutil.NopCloser(bytes.NewReader(date))
	rsp, err := t.client.Do(t.req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return &HTTPRespone{URL: t.req.URL.String(), Byte: body}, err
}

//HTTPRespone HTTP的返回值结构
type HTTPRespone struct {
	URL  string
	Byte []byte
}

//JSON unmarshal to json
//input interface or struct &
func (t *HTTPRespone) JSON(result interface{}) error {
	if t.Byte == nil {
		return errors.New("replyNil")
	}
	return json.Unmarshal(t.Byte, &result)
}

//JSONSelf 和sign一样，只是自己网站使用的东西，这将解析成本网站正常的结构
func (t *HTTPRespone) JSONSelf(result interface{}) error {
	if t.Byte == nil {
		return errors.New("replyNil")
	}
	var rlt struct {
		M string
		R json.RawMessage
		T int64
	}
	err := json.Unmarshal(t.Byte, &rlt)
	if err != nil {
		return err
	}
	if rlt.M != "ok" {
		return errors.New(rlt.M)
	}
	return json.Unmarshal(rlt.R, &result)
}

//String Just Use for DEBUG ! Do not Used in code
func (t *HTTPRespone) String() string {
	if t.Byte == nil {
		return "replyNil"
	}
	return string(t.Byte)
}
