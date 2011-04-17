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
	"firstof": firstofTag,
}

type firstof []valuer

func firstofTag(p *parser) Renderer {
	p.expect(tokIdent)
	var f firstof
	for p.tok != tokBlockTagEnd {
		v := p.parseVar()
		f = append(f, v)
	}
	p.expect(tokBlockTagEnd)
	return f
}

func (f firstof) Render(wr io.Writer, context Context) {
	var v value
	var b bool
	for _, val := range f {
		v = val.value(context)
		b, _ = valueAsBool(v)
		if b {
			str, _ := valueAsString(v)
			wr.Write([]byte(str))
			return
		}
	}
}
