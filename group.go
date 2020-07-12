/*
只是为了更快的输入路由，如？
先为/v1创建一个group
之后只要v1.add("/app1")  ("/app2")
就可以创建两条新的陆游 /v1/app1  /v1/app2
*/
package ohttp

//Group easy to write path
type Group struct {
	t *Server
	s string
}

//Add Group Add router make sure that has "/"
func (t *Group) Add(s string, f func(map[string]string) (interface{}, error)) {
	t.t.Add(t.s+s, f)
}

//AddAuth Group Add router make sure that has "/"
func (t *Group) AddAuth(s string, f func(map[string]string) (interface{}, error)) {
	t.t.AddAuth(t.s+s, f)
}

//Group group
func (t *Group) Group(s string) *Group {
	result := &Group{t: t.t, s: t.s + s}
	return result
}
