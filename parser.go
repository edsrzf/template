package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type parser struct {
	l   *lexer
	tok token
	lit []byte
	s   *scope
}

func (p *parser) Error(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (p *parser) next() {
	p.tok, p.lit = p.l.scan()
}

func (p *parser) Expect(tok token) string {
	if p.tok != tok {
		p.Error("expected %s, got %s", tokStrings[tok], tokStrings[p.tok])
	}
	lit := p.lit
	p.next()
	return string(lit)
}

func (p *parser) ExpectWord(word string) {
	if p.tok != tokIdent || string(p.lit) != word {
		p.Error("expected ident %s, got token %s, %s", word, tokStrings[p.tok], p.lit)
	}
	p.next()
}

// parse until one of the following tags
func (p *parser) ParseUntil(tags ...string) (string, NodeList) {
	r := make(NodeList, 0, 10)
	for p.tok != tokEof {
		switch p.tok {
		case tokText:
			r = append(r, printLit(p.lit))
			p.next()
		case tokBlockTagStart:
			p.next()
			lit := string(p.lit)
			for _, t := range tags {
				if t == lit {
					p.next()
					return t, r
				}
			}
			r = append(r, p.parseBlockTag())
		case tokVarTagStart:
			r = append(r, p.parseVarTag())
		default:
			p.Error("unexpected token %s", tokStrings[p.tok])
		}
	}
	return "", r
}

func (p *parser) parseBlockTag() Node {
	if tag, ok := tags[p.Expect(tokIdent)]; ok {
		node := tag(p)
		p.Expect(tokBlockTagEnd)
		return node
	}
	p.Error("tag isn't registered")
	return nil
}

func (p *parser) parseVarTag() Node {
	p.Expect(tokVarTagStart)
	e := p.parseExpr()
	p.Expect(tokVarTagEnd)
	return e
}

func (p *parser) parseExpr() valuer {
	v := p.parseVal()
	a := p.parseAttrs()
	f := p.parseFilters()
	if len(a) == 0 && len(f) == 0 {
		return v
	}
	return &expr{v, a, f}
}

func (p *parser) parseVal() valuer {
	var ret valuer
	switch p.tok {
	case tokInt:
		i, err := strconv.Atoi64(string(p.lit))
		if err != nil {
			p.Error("internal int error: %s", err)
		}
		ret = intLit(i)
		p.next()
	case tokFloat:
		f, err := strconv.Atof64(string(p.lit))
		if err != nil {
			p.Error("Internal float error: %s", err)
		}
		ret = floatLit(f)
		p.next()
	case tokString:
		ret = stringLit(p.lit)
		p.next()
	case tokIdent:
		ret = p.parseVar()
	default:
		p.Error("Unexpected token %s", tokStrings[p.tok])
	}
	return ret
}

func (p *parser) parseVar() variable {
	return p.s.Lookup(p.Expect(tokIdent))
}

func (p *parser) parseAttrs() []string {
	var attrs []string
	for p.tok == tokDot {
		p.Expect(tokDot)
		attrs = append(attrs, string(p.lit))
		if p.tok == tokInt {
			p.next()
		} else {
			p.Expect(tokIdent)
		}
	}
	return attrs
}

func (p *parser) parseFilters() []*filter {
	var f []*filter
	for p.tok == tokFilter {
		p.next()
		rf, ok := filters[string(p.lit)]
		if !ok {
			p.Error("filter does not exist")
		}
		p.Expect(tokIdent)
		var val valuer
		args := false
		switch rf.arg {
		case ReqArg:
			args = true
		case OptArg:
			args = p.tok == tokArgument
		case NoArg:
			if p.tok == tokArgument {
				p.Error("filter accepts no arguments")
			}
		}
		if args {
			p.Expect(tokArgument)
			val = p.parseExpr()
		}
		f = append(f, &filter{rf.f, val})
	}
	return f
}

func Parse(s []byte) (*Template, os.Error) {
	t := new(Template)
	l := &lexer{src: s}
	l.init()
	p := &parser{l: l, s: newScope()}

	p.next()
	_, t.nodes = p.ParseUntil()
	t.scope = p.s

	return t, nil
}

func ParseString(s string) (*Template, os.Error) {
	return Parse([]byte(s))
}

func ParseFile(name string) (*Template, os.Error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}
