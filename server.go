package ohttp

import (
	"errors"
	"fmt"
	"net/http"
)

//Server http-server struct
type Server struct {
	router     map[string]func(map[string]string) (interface{}, error)
	routerAuth map[string]bool
	header     map[string]string
}

//Group easy to write path
type Group struct {
	t *Server
	s string
}

//Add Group Add router make sure that has "/"
func (t *Group) Add(s string, f func(map[string]string) (interface{}, error)) {
	t.t.Add(t.s+s, f)
}

//Group group
func (t *Group) Group(s string) *Group {
	result := &Group{t: t.t, s: t.s + s}
	return result
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
		router: map[string]func(map[string]string) (interface{}, error){}, routerAuth: map[string]bool{},
		header: map[string]string{},
	}
	http.HandleFunc("/", t.handle)
	return t
}

//Add addrouter make sure that has "/"
func (t *Server) Add(s string, f func(map[string]string) (interface{}, error)) {
	t.router[s] = f
}

//Run @params :port
func (t *Server) Run(port string) {
	http.ListenAndServe(port, nil)
}

func (t *Server) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//Add Header
	for k, v := range t.header {
		w.Header().Set(k, v)
	}
	m := parse(r)
	if t.routerAuth[r.URL.Path] {
		//TODO check auth
	}
	if f, ok := t.router[r.URL.Path]; ok {
		p := parse(r)
		result, err := f(p)
		doRespond(w, result, err)
	} else {
		doRespond(w, nil, errors.New("UnknowMethod"))
	}
	fmt.Println(m)
	fmt.Println(r.URL.Path)
	fmt.Println(r.RequestURI)
}
