package template

import (
	"io"
	"reflect"
)

type TagFunc func(p *parser) Renderer

var tags = map[string]TagFunc{
	"firstof": parseFirstof,
	"for":     parseFor,
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
	for _, val := range f {
		v = val.value(s)
		if valueAsBool(v) {
			str := valueAsString(v)
			wr.Write([]byte(str))
			return
		}
	}
}

type forTag struct {
	v          *variable
	collection valuer
	r          Renderer
	elseNode   Renderer
}

func parseFor(p *parser) Renderer {
	name := p.Expect(tokIdent)
	v := &variable{v: p.s.Lookup(name)}
	p.ExpectWord("in")
	collection := p.parseVar()
	switch collection.(type) {
	case intLit, floatLit:
		p.Error("numeric literals are not iterable")
	}
	p.Expect(tokBlockTagEnd)
	tok, r := p.ParseUntil("else", "endfor")
	var elseNode Renderer
	if tok == "else" {
		p.Expect(tokBlockTagEnd)
		tok, elseNode = p.ParseUntil("endfor")
	}
	if tok != "endfor" {
		p.Error("unterminated for tag")
	}
	p.Expect(tokBlockTagEnd)
	return &forTag{v, collection, r, elseNode}
}

// TODO: this needs reworking. We need a good way to set variables on the stack.
func (f *forTag) Render(wr io.Writer, s Stack) {
	v := f.collection.value(s)
	n := 0
	switch v := v.(type) {
	case string:
		n = len(v)
		for _, c := range v {
			s[f.v.v] = string(c)
			f.r.Render(wr, s)
		}
	case reflect.Value:
		v = reflect.Indirect(v)
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			n = v.Len()
			for i := 0; i < n; i++ {
				s[f.v.v] = refToVal(v.Index(i))
				f.r.Render(wr, s)
			}
		case reflect.Chan:
			for {
				x, ok := v.TryRecv()
				if !ok {
					break
				}
				s[f.v.v] = refToVal(x)
				f.r.Render(wr, s)
				n++
			}
		case reflect.Map:
			n = v.Len()
			for _, k := range v.MapKeys() {
				s[f.v.v] = refToVal(v.MapIndex(k))
				f.r.Render(wr, s)
			}
		case reflect.Struct:
			n = v.NumField()
			for i := 0; i < n; i++ {
				s[f.v.v] = refToVal(v.Field(i))
				f.r.Render(wr, s)
			}
		}
	}
	if n == 0 && f.elseNode != nil {
		f.elseNode.Render(wr, s)
	}
}
