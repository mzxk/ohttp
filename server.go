package ohttp

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

//Server http-server struct
type Server struct {
	router     map[string]func(map[string]string) (interface{}, error)
	routerAuth map[string]bool
	header     map[string]string
	limit      int64
}

//Access HeaderAccess for Allow Origin if empty set *
func (t *Server) Access(s string) {
	if s == "" {
		s = "*"
	}
	t.HeaderSet("Access-Control-Allow-Origin", s)
}

//HeaderSet setHeader default
func (t *Server) HeaderSet(k, v string) {
	t.header[k] = v
}

//Group 1
func (t *Server) Group(s string) *Group {
	return &Group{t: t, s: s}
}

//New handle a new server
func New() *Server {
	t := &Server{
		router:     map[string]func(map[string]string) (interface{}, error){},
		routerAuth: map[string]bool{},
		header:     map[string]string{},
	}
	http.HandleFunc("/", t.handle)
	return t
}

//NewWithSession handle a new server and init session
func NewWithSession(add, pwd string) *Server {
	t := New()
	initSession(add, pwd)
	return t
}

//Add addrouter make sure that has "/"
func (t *Server) Add(s string, f func(map[string]string) (interface{}, error)) {
	t.router[s] = f
}

//AddAuth addrouter make sure that has "/"
func (t *Server) AddAuth(s string, f func(map[string]string) (interface{}, error)) {
	t.router[s] = f
	t.routerAuth[s] = true
}

//Run @params :port
func (t *Server) Run(port string) {
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}

func (t *Server) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//Add Header
	for k, v := range t.header {
		w.Header().Set(k, v)
	}

	if f, ok := t.router[r.URL.Path]; ok {
		m := parse(r)
		if t.routerAuth[r.URL.Path] {
			bsonid, err := t.checkSign(r, m)
			if err != nil {
				doRespond(w, nil, err)
				return
			}
			m["bsonid"] = bsonid
		}
		result, err := f(m)
		doRespond(w, result, err)
		return
	}
	doRespond(w, nil, errors.New("UnknowMethod"))
}
func (t *Server) checkSign(r *http.Request, m map[string]string) (string, error) {
	tm := s2i(m["nonce"]) - time.Now().Unix()
	if tm > 3 || tm < -3 {
		return "", errors.New("SignTimeWrong")
	}
	sign := r.Header.Get("sign")
	ids, value := loadSession(m["key"])
	if value == "" {
		return "", errors.New("SignExpire")
	}
	if sha(r.RequestURI+value) == sign {
		updateSession(m["key"])
		return ids, nil
	}
	return "", errors.New("SignError")
}
func s2i(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		i = 0
	}
	return i
}
