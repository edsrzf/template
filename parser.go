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
func (p *parser) ParseUntil(tags ...string) (string, RenderList) {
	r := make(RenderList, 0, 10)
	for p.tok != tokEof {
		switch p.tok {
		case tokText:
			r = append(r, &printLit{p.lit})
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

func (p *parser) parseBlockTag() Renderer {
	if tag, ok := tags[p.Expect(tokIdent)]; ok {
		return tag(p)
	}
	p.Error("tag isn't registered")
	return nil
}

func (p *parser) parseVarTag() Renderer {
	p.Expect(tokVarTagStart)
	v := p.parseVar()
	f := p.parseFilters()
	p.Expect(tokVarTagEnd)
	return &printVar{&expr{v, f}}
}

func (p *parser) parseVar() valuer {
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
		ret = p.parseAttrVar()
	default:
		p.Error("Unexpected token %s", tokStrings[p.tok])
	}
	return ret
}

func (p *parser) parseAttrVar() *variable {
	var v variable
	v.v = p.s.Lookup(p.Expect(tokIdent))
	for p.tok == tokDot {
		p.Expect(tokDot)
		v.attrs = append(v.attrs, string(p.lit))
		if p.tok == tokInt {
			p.Expect(tokInt)
		} else {
			p.Expect(tokIdent)
		}
	}
	return &v
}

func (p *parser) parseFilters() []*filter {
	f := make([]*filter, 0, 2)
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
			val = p.parseVar()
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
