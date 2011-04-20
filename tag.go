package template

import (
	"io"
	"reflect"
)

type TagFunc func(p *parser) Node

var tags = map[string]TagFunc{
	"cycle":     parseCycle,
	"firstof":   parseFirstof,
	"for":       parseFor,
	"if":        parseIf,
	"ifchanged": parseIfChanged,
	"set":       parseSet,
	"with":      parseWith,
}

type cycleTag struct {
	args  []Node
	state Variable
}

func parseCycle(p *parser) Node {
	args := make([]Node, 0, 2)
	for p.tok != tokBlockTagEnd {
		v := p.parseExpr()
		args = append(args, v)
	}
	if len(args) < 1 {
		p.Error("the cycle tag requires at least one parameter")
	}
	state := p.s.Anonymous(intValue(0))
	return &cycleTag{args, state}
}

func (c cycleTag) Render(wr io.Writer, s Stack) {
	i := c.state.Int(s)
	c.args[i].Render(wr, s)
	i++
	if int(i) >= len(c.args) {
		i = 0
	}
	c.state.Set(intValue(i), s)
}

type firstofTag []Value

func parseFirstof(p *parser) Node {
	tag := make(firstofTag, 0, 2)
	for p.tok != tokBlockTagEnd {
		v := p.parseExpr()
		tag = append(tag, v)
	}
	return tag
}

func (f firstofTag) Render(wr io.Writer, s Stack) {
	for _, val := range f {
		if val.Bool(s) {
			val.Render(wr, s)
			return
		}
	}
}

type forTag struct {
	v          Variable
	collection Value
	init       Node
	body       Node
	elseNode   Node
}

func parseFor(p *parser) Node {
	p.s.Push()
	name := p.Expect(tokIdent)
	v := p.s.Insert(name)
	p.ExpectWord("in")
	collection := p.parseExpr()
	switch collection.(type) {
	case intValue, floatValue:
		p.Error("numeric literals are not iterable")
	}
	p.Expect(tokBlockTagEnd)
	tok, body := p.ParseUntil("else", "endfor")
	var elseNode Node
	if tok == "else" {
		p.Expect(tokBlockTagEnd)
		tok, elseNode = p.ParseUntil("endfor")
	}
	if tok != "endfor" {
		p.Error("unterminated for tag")
	}
	return &forTag{v, collection, p.s.Pop(), body, elseNode}
}

// TODO: this needs reworking. We need a good way to set Variables on the stack.
func (f *forTag) Render(wr io.Writer, s Stack) {
	f.init.Render(wr, s)
	v := f.collection.Reflect(s)
	v = reflect.Indirect(v)
	n := 0
	switch v.Kind() {
	case reflect.String:
		v := f.collection.String(s)
		n = len(v)
		for _, c := range v {
			f.v.Set(stringValue(c), s)
			f.body.Render(wr, s)
		}
	case reflect.Array, reflect.Slice:
		n = v.Len()
		for i := 0; i < n; i++ {
			f.v.Set(refToVal(v.Index(i)), s)
			f.body.Render(wr, s)
		}
	case reflect.Chan:
		for {
			x, ok := v.TryRecv()
			if !ok {
				break
			}
			f.v.Set(refToVal(x), s)
			f.body.Render(wr, s)
			n++
		}
	case reflect.Map:
		n = v.Len()
		for _, k := range v.MapKeys() {
			f.v.Set(refToVal(v.MapIndex(k)), s)
			f.body.Render(wr, s)
		}
	case reflect.Struct:
		n = v.NumField()
		for i := 0; i < n; i++ {
			f.v.Set(refToVal(v.Field(i)), s)
			f.body.Render(wr, s)
		}
	}
	if n == 0 && f.elseNode != nil {
		f.elseNode.Render(wr, s)
	}
}

type ifChangedTag struct {
	vals      []Value
	last      []Variable
	ifNodes   NodeList
	elseNodes NodeList
}

func parseIfChanged(p *parser) Node {
	args := make([]Value, 0, 2)
	for p.tok != tokBlockTagEnd {
		v := p.parseExpr()
		args = append(args, v)
	}
	p.Expect(tokBlockTagEnd)
	vars := make([]Variable, len(args))
	for i := range vars {
		// use a value that can never otherwise occur so we always detect a
		// change the first time
		vars[i] = p.s.Anonymous(nilValue(1))
	}
	tok, ifNodes := p.ParseUntil("else", "endifchanged")
	var elseNodes NodeList
	if tok == "else" {
		p.Expect(tokBlockTagEnd)
		tok, elseNodes = p.ParseUntil("endifchanged")
	}
	if tok != "endifchanged" {
		p.Error("unterminated ifchanged tag")
	}
	return &ifChangedTag{args, vars, ifNodes, elseNodes}
}

func (t *ifChangedTag) Render(wr io.Writer, s Stack) {
	changed := false
	for i, v := range t.last {
		new := t.vals[i].Eval(s)
		// TODO: new != old can panic depending on the concrete values
		if !changed && new != v.Eval(s) {
			changed = true
		}
		v.Set(new, s)
	}
	if changed {
		t.ifNodes.Render(wr, s)
	} else if t.elseNodes != nil {
		t.elseNodes.Render(wr, s)
	}
}

type setTag struct {
	v Variable
	e Value
}

func parseSet(p *parser) Node {
	name := p.Expect(tokIdent)
	v := p.s.Insert(name)
	e := p.parseExpr()
	return &setTag{v, e}
}

func (t *setTag) Render(wr io.Writer, s Stack) {
	t.v.Set(t.e.Eval(s), s)
}

type with NodeList

func parseWith(p *parser) Node {
	p.s.Push()
	p.Expect(tokBlockTagEnd)
	tok, nodes := p.ParseUntil("endwith")
	if tok != "endwith" {
		p.Error("unterminated with tag")
	}
	init := p.s.Pop()
	nodes = append(nodes, nil)
	copy(nodes[1:], nodes)
	nodes[0] = init
	return with(nodes)
}

func (w with) Render(wr io.Writer, s Stack) { NodeList(w).Render(wr, s) }
