/*
提供了一个简易的http调用包
可以使用get,post,增加header等
同时提供自用的签名封装
一个简单的例子：
	ohttp.HTTP("https://baidu.com/s",map[string]string{"wd":"123"}).Proxy("http://127.0.0.1:8080").Get()
	这里调用了一个简单的get，最终url为https://www.baidu.com/s?wd=123,同时使用了本地的Proxy
*/
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
//其他详情请参考session注释
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

//Header 这里给请求加上header
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

//Body 向请求的Body写入内容
func (t *HTTPStr) Body(date []byte) *HTTPStr {
	t.req.Body = ioutil.NopCloser(bytes.NewReader(date))
	return t
}

//Method 写入请求的方法GET,POST,PUT 等
func (t *HTTPStr) Method(method string) *HTTPStr {
	t.req.Method = method
	return t
}

//Do 最终进行请求，并返回
func (t *HTTPStr) Do() (*HTTPRespone, error) {
	if t.err != nil { //首先判断之前的过程中，是否存在错误
		return nil, t.err
	}
	rsp, err := t.client.Do(t.req)
	if err != nil { //请求后是否存在错误
		return nil, err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body) //读取返回body时是否存在错误
	if err != nil {
		return nil, err
	}
	return &HTTPRespone{URL: t.req.URL.String(), Byte: body}, err
}

//Get 进行get的返回,易用性封装
func (t *HTTPStr) Get() (*HTTPRespone, error) {
	return t.Method("GET").Do()
}

//Post POST 进行POST的返回 易用性封装
func (t *HTTPStr) Post(date []byte) (*HTTPRespone, error) {
	return t.Method("POST").Body(date).Do()
}

//HTTPRespone HTTP的返回值结构
//URL 总请求url
//Byte 获取的返回值
type HTTPRespone struct {
	URL  string
	Byte []byte
}

//JSON 这将用传入的result指针来解析body。
func (t *HTTPRespone) JSON(result interface{}) error {
	if t.Byte == nil {
		return errors.New("replyNil")
	}
	return json.Unmarshal(t.Byte, &result)
}

//JSONSelf 和sign一样，只是自己网站使用的东西，这将解析成本网站正常的结构（参考下面rlt struct的结构
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
