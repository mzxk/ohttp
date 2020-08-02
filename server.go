/*
	Sign的方式：
		1：登陆，获取后端返回的key和value。
				例如：key = foo , value = bar
		2：在正常的请求上url，增加两个参数：
			nonce	时间戳 精确到秒
			key		login后后端返回的key，value之一的key
				例如：
					原始请求是http://localhost/user/changePwd?oldpwd=old&newpwd=new
					增加后变成http://localhost/user/changePwd?oldpwd=old&newpwd=new&key=foo&nonce=158754624
		3：计算sign
			使用整个requestURI 不包含域名和端口的部分 加上 value 进行sha512计算 ， 然后截取[58:98]
				例如：
					needsign:="/user/changePwd?oldpwd=old&newpwd=new&key=foo&nonce=158754624"+"bar"
					sign=sha512(needsign)[58:98]
					"bf11206c32dc27301f681da41fc3570937c1e2cd"
			然后在请求request里增加一个header，名字为key,值为上面的sign

		请注意：
			如果需要加密的字符串里含有中文字符,那么一定要在url被转码成%ab的格式后在进行签名，直接使用中文拼接进行的签名是无法通过检查的
*/
package ohttp

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/mzxk/oval"
)

//Server http-server struct
type Server struct {
	router     map[string]func(map[string]string) (interface{}, error) //路由的主要结构体
	routerAuth map[string]bool                                         //确认当前路由是否需要验证签名
	header     map[string]string                                       //写在这里的东西将会写在返回值的header里
	LimitIP    int64
	LimitID    sync.Map
}

//Access 跨域问题处理，当s为空时，默认允许所有域 ×
//如果是面对网页的前后分离的后端 必须要调用此函数
func (t *Server) Access(s string) {
	if s == "" {
		s = "*"
	}
	t.HeaderSet("Access-Control-Allow-Origin", s)
}

//HeaderSet 设置返回头
func (t *Server) HeaderSet(k, v string) {
	t.header[k] = v
}

//Group 一个简单的方式设置路由，调用group的add
func (t *Server) Group(s string) *Group {
	return &Group{t: t, s: s}
}

//SetLimitIP 设置每IP的访问限制，这是一个统一限速
func (t *Server) SetLimitIP(i int64) {
	if i <= 0 {
		i = 1200
	}
	t.LimitIP = i
}

//SetLimitID 设置每ID（既签名接口）的访问限制，如果不设置，默认为120/分钟
func (t *Server) SetLimitID(id string, i int64) {
	if i <= 0 {
		i = 120
	}
	t.LimitID.Store(id, i)
}
func (t *Server) getLimitID(id string) int64 {
	if i, ok := t.LimitID.Load(id); ok {
		return i.(int64)
	}
	return 120
}

//New handle a new server
func New() *Server {
	t := &Server{
		router:     map[string]func(map[string]string) (interface{}, error){},
		routerAuth: map[string]bool{},
		header:     map[string]string{},
		LimitIP:    1200,
	}
	http.HandleFunc("/", t.handle)
	return t
}

//NewWithSession 这将提供一个带有session的服务端，需要输入redis的地址和密码
func NewWithSession(add, pwd string) *Server {
	t := New()
	initSession(add, pwd)
	return t
}

//Add 增加一个静态路由，记得以“/”开始
func (t *Server) Add(s string, f func(map[string]string) (interface{}, error)) {
	t.router[s] = f
}

//AddAuth 增加一个需要验证签名静态路由，记得以“/”开始
func (t *Server) AddAuth(s string, f func(map[string]string) (interface{}, error)) {
	t.router[s] = f
	t.routerAuth[s] = true
}

//Run 启动坚挺，记得port的格式是：12345
func (t *Server) Run(port string) {
	if len(t.header) == 0 {
		t.Access("*")
	}
	t.HeaderSet("Access-Control-Allow-Headers", "*")
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}

//最主要的处理结构
func (t *Server) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//Add Header
	for k, v := range t.header {
		w.Header().Set(k, v)
	}

	if f, ok := t.router[r.URL.Path]; ok { //路由存在
		m := parse(r) //解析
		//每分钟1200次的IP限制
		if oval.Limited(m["ip"], 60, t.LimitIP) {
			doRespond(w, nil, errors.New("outLimit"))
			return
		}
		if t.routerAuth[r.URL.Path] { //此路由需要验证签名
			bsonid, err := t.checkSign(r, m) //验证签名
			if err != nil {                  //验证失败
				doRespond(w, nil, err)
				return
			}
			if oval.Limited(bsonid, 60, t.getLimitID(bsonid)) {
				doRespond(w, nil, errors.New("outLimit"))
				return
			}
			m["bsonid"] = bsonid //签名验证成功，把用户id写入统一输入
		}
		//下面这一条是所有代码里最重要的一条
		result, err := f(m)
		//所有我分了两条注释，他调用了此路由从输入函数并返回值
		doRespond(w, result, err)
		return
	}
	doRespond(w, nil, errors.New("UnknowMethod"))
}

//验证签名 签名在请求的header里sign保存，
//签名的方式是用户验证requestURI里key是否存在，
//验证nonce是否和服务器时间一致（2妙内），验证sha512(
//requestURI+value)[58:98]是否和sign一致
func (t *Server) checkSign(r *http.Request, m map[string]string) (string, error) {
	//验证key是否存在
	if m["key"] == "" {
		return "", errors.New("SignKeyWrong")
	}
	//验证时间是否在3妙内
	tm := s2i(m["nonce"]) - time.Now().Unix()
	if tm > 2 || tm < -2 {
		return "", errors.New("SignTimeWrong")
	}
	//判断sign是否存在并且一致
	sign := r.Header.Get("sign")
	if sign == "" {
		return "", errors.New("SignError")
	}
	ids, value := loadSession(m["key"])
	if value == "" {
		return "", errors.New("SignExpire")
	}
	if mac(r.RequestURI, value) == sign {
		//更新此key的过期时间，并且返回用户id
		updateSession(m["key"])
		return ids, nil
	}
	return "", errors.New("SignError")
}
