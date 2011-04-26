package template

import (
	"bytes"
	"io"
	"os"
	"reflect"
)

type TagFunc func(p *Parser) Node

var tags = map[string]TagFunc{
	"block":     parseBlock,
	"cycle":     parseCycle,
	"extends":   parseExtends,
	"firstof":   parseFirstof,
	"for":       parseFor,
	"if":        parseIf,
	"ifchanged": parseIfChanged,
	"include":   parseInclude,
	"override":  parseOverride,
	"set":       parseSet,
	"with":      parseWith,
}

type blockTag struct {
	name  Variable
	nodes NodeList
}

func parseBlock(p *Parser) Node {
	name := p.Expect(TokIdent)
	p.Expect(TokTagEnd)
	tok, nodes := p.ParseUntil("endblock")
	if tok != "endblock" {
		p.Error("unterminated block tag")
	}
	nameVar := p.Scope().Lookup("@" + name)
	return &blockTag{nameVar, nodes}
}

func (b *blockTag) Render(wr io.Writer, c *Context) {
	val := b.name.Eval(c)
	if val.Bool() {
		val.Render(wr, c)
		return
	}
	b.nodes.Render(wr, c)
}

type cycleTag struct {
	args  []Expr
	state Variable
}

func parseCycle(p *Parser) Node {
	args := make([]Expr, 0, 2)
	for p.Current() != TokTagEnd {
		v := p.ParseExpr()
		args = append(args, v)
	}
	if len(args) < 1 {
		p.Error("the cycle tag requires at least one parameter")
	}
	state := p.Scope().Anonymous(intValue(0))
	return &cycleTag{args, state}
}

func (t cycleTag) Render(wr io.Writer, c *Context) {
	i := t.state.Eval(c).Int()
	t.args[i].Eval(c).Render(wr, c)
	i++
	if int(i) >= len(t.args) {
		i = 0
	}
	t.state.Set(intValue(i), c)
}

type extendsTag struct {
	parent Expr
	nodes NodeList
}

func parseExtends(p *Parser) Node {
	parent := p.ParseExpr()
	p.Expect(TokTagEnd)
	tok, nodes := p.ParseUntil("endextends")
	if tok != "endextends" {
		p.Error("unterminated extends tag")
	}
	return &extendsTag{parent, nodes}
}

type nilWriter int

func (w nilWriter) Write(p []byte) (int, os.Error) { return len(p), nil }

func (e *extendsTag) Render(wr io.Writer, c *Context) {
	parentValue := e.parent.Eval(c)
	node, ok := parentValue.Reflect().Interface().(*Template)
	if !ok {
		// must be a string
		filename := parentValue.String()
		var err os.Error
		node, err = ParseFile(filename)
		if err != nil {
			return
		}
	}
	w := nilWriter(0)
	e.nodes.Render(w, c)
	node.Render(wr, c)
}

type firstofTag []Expr

func parseFirstof(p *Parser) Node {
	tag := make(firstofTag, 0, 2)
	for p.Current() != TokTagEnd {
		v := p.ParseExpr()
		tag = append(tag, v)
	}
	return tag
}

func (f firstofTag) Render(wr io.Writer, c *Context) {
	for _, expr := range f {
		if val := expr.Eval(c); val.Bool() {
			val.Render(wr, c)
			return
		}
	}
}

type forTag struct {
	v          Variable
	collection Expr
	init       Node
	body       Node
	elseNode   Node
}

func parseFor(p *Parser) Node {
	scope := p.Scope()
	scope.Push()
	name := p.Expect(TokIdent)
	v := scope.Insert(name)
	p.ExpectWord("in")
	collection := p.ParseExpr()
	p.Expect(TokTagEnd)
	tok, body := p.ParseUntil("else", "endfor")
	var elseNode Node
	if tok == "else" {
		p.Expect(TokTagEnd)
		tok, elseNode = p.ParseUntil("endfor")
	}
	if tok != "endfor" {
		p.Error("unterminated for tag")
	}
	return &forTag{v, collection, scope.Pop(), body, elseNode}
}

// TODO: this needs reworking. We need a good way to set Variables on the stack.
func (f *forTag) Render(wr io.Writer, c *Context) {
	f.init.Render(wr, c)
	colVal := f.collection.Eval(c)
	v := colVal.Reflect()
	v = reflect.Indirect(v)
	n := 0
	switch v.Kind() {
	case reflect.String:
		v := colVal.String()
		n = len(v)
		for _, ch := range v {
			f.v.Set(stringValue(ch), c)
			f.body.Render(wr, c)
		}
	case reflect.Array, reflect.Slice:
		n = v.Len()
		for i := 0; i < n; i++ {
			f.v.Set(refToVal(v.Index(i)), c)
			f.body.Render(wr, c)
		}
	case reflect.Chan:
		for {
			x, ok := v.TryRecv()
			if !ok {
				break
			}
			f.v.Set(refToVal(x), c)
			f.body.Render(wr, c)
			n++
		}
	case reflect.Map:
		n = v.Len()
		for _, k := range v.MapKeys() {
			f.v.Set(refToVal(v.MapIndex(k)), c)
			f.body.Render(wr, c)
		}
	case reflect.Struct:
		n = v.NumField()
		for i := 0; i < n; i++ {
			f.v.Set(refToVal(v.Field(i)), c)
			f.body.Render(wr, c)
		}
	}
	if n == 0 && f.elseNode != nil {
		f.elseNode.Render(wr, c)
	}
}

type ifChangedTag struct {
	vals      []Expr
	last      []Variable
	ifNodes   NodeList
	elseNodes NodeList
}

func parseIfChanged(p *Parser) Node {
	args := make([]Expr, 0, 2)
	for p.Current() != TokTagEnd {
		v := p.ParseExpr()
		args = append(args, v)
	}
	p.Expect(TokTagEnd)
	vars := make([]Variable, len(args))
	for i := range vars {
		// use a value that can never otherwise occur so we always detect a
		// change the first time
		vars[i] = p.Scope().Anonymous(nilValue(1))
	}
	tok, ifNodes := p.ParseUntil("else", "endifchanged")
	var elseNodes NodeList
	if tok == "else" {
		p.Expect(TokTagEnd)
		tok, elseNodes = p.ParseUntil("endifchanged")
	}
	if tok != "endifchanged" {
		p.Error("unterminated ifchanged tag")
	}
	return &ifChangedTag{args, vars, ifNodes, elseNodes}
}

func (t *ifChangedTag) Render(wr io.Writer, c *Context) {
	changed := false
	for i, v := range t.last {
		new := t.vals[i].Eval(c)
		// TODO: new != old can panic depending on the concrete values
		if !changed && new != v.Eval(c) {
			changed = true
		}
		v.Set(new, c)
	}
	if changed {
		t.ifNodes.Render(wr, c)
	} else if t.elseNodes != nil {
		t.elseNodes.Render(wr, c)
	}
}

type includeTag struct {
	e Expr
}

func parseInclude(p *Parser) Node {
	expr := p.ParseExpr()
	return includeTag{expr}
}

func (i includeTag) Render(wr io.Writer, c *Context) {
	val := i.e.Eval(c)
	node, ok := val.Reflect().Interface().(*Template)
	if !ok {
		// must be a string
		filename := val.String()
		var err os.Error
		node, err = ParseFile(filename)
		if err != nil {
			return
		}
	}
	node.Render(wr, c)
}

type setTag struct {
	v Variable
	e Expr
}

type overrideTag struct {
	name  string
	nameVar Variable
	nodes NodeList
}

func parseOverride(p *Parser) Node {
	name := p.Expect(TokIdent)
	p.Expect(TokTagEnd)
	tok, nodes := p.ParseUntil("endoverride")
	if tok != "endoverride" {
		p.Error("unterminated block tag")
	}
	name = "@" + name
	nameVar := p.Scope().Lookup(name)
	return &overrideTag{name, nameVar, nodes}
}

func (o *overrideTag) Render(wr io.Writer, c *Context) {
	var buf bytes.Buffer
	o.nodes.Render(&buf, c)
	str := buf.String()
	o.nameVar.Set(stringValue(str), c)
	c.vars[o.name] = str
}

func parseSet(p *Parser) Node {
	name := p.Expect(TokIdent)
	v := p.Scope().Insert(name)
	e := p.ParseExpr()
	return &setTag{v, e}
}

func (t *setTag) Render(wr io.Writer, c *Context) {
	t.v.Set(t.e.Eval(c), c)
}

type with NodeList

func parseWith(p *Parser) Node {
	scope := p.Scope()
	scope.Push()
	p.Expect(TokTagEnd)
	tok, nodes := p.ParseUntil("endwith")
	if tok != "endwith" {
		p.Error("unterminated with tag")
	}
	init := scope.Pop()
	nodes = append(nodes, nil)
	copy(nodes[1:], nodes)
	nodes[0] = init
	return with(nodes)
}

func (w with) Render(wr io.Writer, c *Context) { NodeList(w).Render(wr, c) }
