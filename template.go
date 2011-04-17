package template

import (
	"io"
	"os"
	"strconv"
)

type Context map[string]interface{}

type Stack []value

type Renderer interface {
	Render(wr io.Writer, context Context)
}

type printLit struct {
	content string
}

func (p *printLit) Render(wr io.Writer, context Context) {
	wr.Write([]byte(p.content))
}

type printVar struct {
	val *expr
}

func (p *printVar) Render(wr io.Writer, context Context) {
	v := p.val.value(context)
	s, _ := valueAsString(v)
	wr.Write([]byte(s))
}

type Template struct {
	nodes []Renderer
}

func (t *Template) push(n Renderer) {
	t.nodes = append(t.nodes, n)
}

func (t *Template) Execute(wr io.Writer, context Context) {
	for _, node := range t.nodes {
		node.Render(wr, context)
	}
}

type parser struct {
	l   *lexer
	tok token
	lit string
	s   scope
}

func (p *parser) next() {
	p.tok, p.lit = p.l.scan()
}

func (p *parser) expect(tok token) {
	if p.tok != tok {
		panic("expected " + tokStrings[tok] + ", got " + tokStrings[p.tok])
	}
	p.next()
}

func (p *parser) parseBlockTag() Renderer {
	p.expect(tokBlockTagStart)
	if p.tok != tokIdent {
		// Use expect error handling
		p.expect(tokIdent)
	}
	tag, ok := tags[p.lit]
	if !ok {
		panic("Tag isn't registered")
	}
	return tag(p)
}

func (p *parser) parseVarTag() Renderer {
	p.expect(tokVarTagStart)
	v := p.parseVar()
	f := p.parseFilters()
	p.expect(tokVarTagEnd)
	return &printVar{&expr{v, f}}
}

func (p *parser) parseVar() valuer {
	var ret valuer
	switch p.tok {
	case tokInt:
		i, err := strconv.Atoi64(p.lit)
		if err != nil {
			panic("Internal int error: " + err.String())
		}
		ret = intLit(i)
		p.next()
	case tokFloat:
		f, err := strconv.Atof64(p.lit)
		if err != nil {
			panic("Internal float error: " + err.String())
		}
		ret = floatLit(f)
		p.next()
	case tokString:
		ret = stringLit(p.lit)
		p.next()
	case tokIdent:
		ret = p.parseAttrVar()
	default:
		panic("Unexpected token " + tokStrings[p.tok])
	}
	return ret
}

func (p *parser) parseAttrVar() variable {
	v := make(variable, 0, 2)
	p.s.Lookup(p.lit)
	v = append(v, p.lit)
	p.expect(tokIdent)
	for p.tok == tokDot {
		p.expect(tokDot)
		v = append(v, p.lit)
		if p.tok == tokInt {
			p.expect(tokInt)
		} else {
			p.expect(tokIdent)
		}
	}
	return v
}

func (p *parser) parseFilters() []*filter {
	f := make([]*filter, 0, 2)
	for p.tok == tokFilter {
		p.next()
		rf, ok := filters[p.lit]
		if !ok {
			panic("Filter does not exist")
		}
		p.expect(tokIdent)
		var val valuer
		args := false
		switch rf.arg {
		case ReqArg:
			args = true
		case OptArg:
			args = p.tok == tokArgument
		case NoArg:
			if p.tok == tokArgument {
				panic("Filter accepts no arguments")
			}
		}
		if args {
			p.expect(tokArgument)
			val = p.parseVar()
		}
		f = append(f, &filter{rf.f, val})
	}
	return f
}

func Parse(s string) (*Template, os.Error) {
	t := new(Template)
	l := &lexer{src: []byte(s)}
	l.init()
	p := &parser{l: l}

	p.next()
	for p.tok != tokEof {
		switch p.tok {
		case tokText:
			t.push(&printLit{p.lit})
			p.next()
		case tokBlockTagStart:
			t.push(p.parseBlockTag())
		case tokVarTagStart:
			t.push(p.parseVarTag())
		}
	}

	return t, nil
}
