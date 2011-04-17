package template

import (
	"io"
	"reflect"
)

type Context map[string]interface{}

type Stack []value

type scope map[string]int

func (s scope) Lookup(name string) int {
	v, ok := s[name]
	if ok {
		return v
	}
	v = len(s)
	s[name] = v
	return v
}

type Renderer interface {
	Render(wr io.Writer, s Stack)
}

type RenderList []Renderer

func (l RenderList) Render(wr io.Writer, s Stack) {
	for _, r := range l {
		r.Render(wr, s)
	}
}

type printLit struct {
	content string
}

func (p *printLit) Render(wr io.Writer, s Stack) {
	wr.Write([]byte(p.content))
}

type printVar struct {
	val *expr
}

func (p *printVar) Render(wr io.Writer, s Stack) {
	v := p.val.value(s)
	str := valueAsString(v)
	wr.Write([]byte(str))
}

type Template struct {
	scope scope
	nodes RenderList
}

func (t *Template) Execute(wr io.Writer, c Context) {
	s := make(Stack, len(t.scope))
	if c != nil {
		for k, v := range t.scope {
			if val, ok := c[k]; ok {
				switch val.(type) {
				case bool, float32, float64, complex64, complex128, int, int8, int16,
					int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, string, nil:
				default:
					val = reflect.NewValue(val)
				}
				s[v] = val
			}
		}
	}
	t.nodes.Render(wr, s)
}
