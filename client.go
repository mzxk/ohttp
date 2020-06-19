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
)

//HTTPStr 1
type HTTPStr struct {
	req    *http.Request
	client http.Client
	err    error
}

//HTTP 1
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
	url := u + connect + params.Encode()
	req, err := http.NewRequest("GET", url, nil)

	return &HTTPStr{req: req, client: c, err: err}
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

//String Just Use for DEBUG ! Do not Used in code
func (t *HTTPRespone) String() string {
	if t.Byte == nil {
		return "replyNil"
	}
	return string(t.Byte)
}
