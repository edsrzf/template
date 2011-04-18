package template

import (
	"io"
	"reflect"
)

type TagFunc func(p *parser) Node

var tags = map[string]TagFunc{
	"firstof": parseFirstof,
	"for":     parseFor,
	"if":      parseIf,
	"set":     parseSet,
	"with":    parseWith,
}

type firstofTag []valuer

func parseFirstof(p *parser) Node {
	var f firstofTag
	for p.tok != tokBlockTagEnd {
		v := p.parseExpr()
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
	v          variable
	collection valuer
	r          Node
	elseNode   Node
}

func parseFor(p *parser) Node {
	p.s.Push()
	defer p.s.Pop()
	name := p.Expect(tokIdent)
	v := p.s.Insert(name)
	p.ExpectWord("in")
	collection := p.parseExpr()
	switch collection.(type) {
	case intLit, floatLit:
		p.Error("numeric literals are not iterable")
	}
	p.Expect(tokBlockTagEnd)
	tok, r := p.ParseUntil("else", "endfor")
	var elseNode Node
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
			f.v.set(string(c), s)
			f.r.Render(wr, s)
		}
	case reflect.Value:
		v = reflect.Indirect(v)
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			n = v.Len()
			for i := 0; i < n; i++ {
				f.v.set(refToVal(v.Index(i)), s)
				f.r.Render(wr, s)
			}
		case reflect.Chan:
			for {
				x, ok := v.TryRecv()
				if !ok {
					break
				}
				f.v.set(refToVal(x), s)
				f.r.Render(wr, s)
				n++
			}
		case reflect.Map:
			n = v.Len()
			for _, k := range v.MapKeys() {
				f.v.set(refToVal(v.MapIndex(k)), s)
				f.r.Render(wr, s)
			}
		case reflect.Struct:
			n = v.NumField()
			for i := 0; i < n; i++ {
				f.v.set(refToVal(v.Field(i)), s)
				f.r.Render(wr, s)
			}
		}
	}
	if n == 0 && f.elseNode != nil {
		f.elseNode.Render(wr, s)
	}
}

type setTag struct {
	v variable
	e valuer
}

func parseSet(p *parser) Node {
	name := p.Expect(tokIdent)
	v := p.s.Insert(name)
	e := p.parseExpr()
	p.Expect(tokBlockTagEnd)
	return &setTag{v, e}
}

func (t *setTag) Render(wr io.Writer, s Stack) {
	t.v.set(t.e.value(s), s)
}

type with NodeList

func parseWith(p *parser) Node {
	p.s.Push()
	defer p.s.Pop()
	p.Expect(tokBlockTagEnd)
	tok, nodes := p.ParseUntil("endwith")
	if tok != "endwith" {
		p.Error("unterminated with tag")
	}
	p.Expect(tokBlockTagEnd)
	return with(nodes)
}

func (w with) Render(wr io.Writer, s Stack) { NodeList(w).Render(wr, s) }
