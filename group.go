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
