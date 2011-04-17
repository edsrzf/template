package template

import (
	"io"
)

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

type TagFunc func(p *parser) Renderer

var tags = map[string]TagFunc{
	"firstof": parseFirstof,
	"if":      parseIf,
}

type firstofTag []valuer

func parseFirstof(p *parser) Renderer {
	var f firstofTag
	for p.tok != tokBlockTagEnd {
		v := p.parseVar()
		f = append(f, v)
	}
	p.Expect(tokBlockTagEnd)
	return f
}

func (f firstofTag) Render(wr io.Writer, s Stack) {
	var v value
	var b bool
	for _, val := range f {
		v = val.value(s)
		b, _ = valueAsBool(v)
		if b {
			str, _ := valueAsString(v)
			wr.Write([]byte(str))
			return
		}
	}
}
