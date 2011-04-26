package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type Parser struct {
	l   *lexer
	tok Token
	lit []byte
	s   *scope
}

func (p *Parser) Error(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (p *Parser) Current() Token {
	return p.tok
}

func (p *Parser) Next() {
	p.tok, p.lit = p.l.scan()
}

func (p *Parser) Expect(tok Token) string {
	if p.tok != tok {
		p.Error("expected %s, got %s", tok, p.tok)
	}
	lit := p.lit
	p.Next()
	return string(lit)
}

func (p *Parser) ExpectWord(word string) {
	if p.tok != TokIdent || string(p.lit) != word {
		p.Error("expected ident %s, got Token %s, %s", word, p.tok, p.lit)
	}
	p.Next()
}

// parse until one of the following tags
func (p *Parser) ParseUntil(tags ...string) (string, NodeList) {
	r := make(NodeList, 0, 10)
	for p.tok != TokEof {
		switch p.tok {
		case TokText:
			r = append(r, printLit(p.lit))
			p.Next()
		case TokTagStart:
			p.Next()
			lit := string(p.lit)
			for _, t := range tags {
				if t == lit {
					p.Next()
					return t, r
				}
			}
			r = append(r, p.parseBlockTag())
		case TokVarStart:
			r = append(r, p.parseVarTag())
		default:
			p.Error("unexpected Token %s", p.tok)
		}
	}
	return "", r
}

func (p *Parser) parseBlockTag() Node {
	if tag, ok := tags[p.Expect(TokIdent)]; ok {
		node := tag(p)
		p.Expect(TokTagEnd)
		return node
	}
	p.Error("tag isn't registered")
	return nil
}

func (p *Parser) parseVarTag() Node {
	p.Expect(TokVarStart)
	e := p.ParseExpr()
	p.Expect(TokVarEnd)
	return varTag{e}
}

func (p *Parser) ParseExpr() Expr {
	e := p.parseVal()
	a := p.parseAttrs()
	if len(a) > 0 {
		e = &attrExpr{e, a}
	}
	f := p.parseFilters()
	if len(f) > 0 {
		e = &filterExpr{e, f}
	}
	return e
}

func (p *Parser) parseVal() Expr {
	var ret Expr
	switch p.tok {
	case TokInt:
		i, err := strconv.Atoi64(string(p.lit))
		if err != nil {
			p.Error("internal int error: %s", err)
		}
		ret = constExpr{intValue(i)}
		p.Next()
	case TokFloat:
		f, err := strconv.Atof64(string(p.lit))
		if err != nil {
			p.Error("Internal float error: %s", err)
		}
		ret = constExpr{floatValue(f)}
		p.Next()
	case TokString:
		ret = constExpr{stringValue(p.lit)}
		p.Next()
	case TokIdent:
		ret = p.parseVar()
	default:
		p.Error("Unexpected Token %s", p.tok)
	}
	return ret
}

func (p *Parser) parseVar() Variable {
	return p.s.Lookup(p.Expect(TokIdent))
}

func (p *Parser) parseAttrs() []string {
	var attrs []string
	for p.tok == TokDot {
		p.Expect(TokDot)
		attrs = append(attrs, string(p.lit))
		if p.tok == TokInt {
			p.Next()
		} else {
			p.Expect(TokIdent)
		}
	}
	return attrs
}

func (p *Parser) parseFilters() []*filter {
	var f []*filter
	for p.tok == TokFilter {
		p.Next()
		rf, ok := filters[string(p.lit)]
		if !ok {
			p.Error("filter does not exist")
		}
		p.Expect(TokIdent)
		var val Expr
		args := false
		switch rf.arg {
		case ReqArg:
			args = true
		case OptArg:
			args = p.tok == TokArgument
		case NoArg:
			if p.tok == TokArgument {
				p.Error("filter accepts no arguments")
			}
		}
		if args {
			p.Expect(TokArgument)
			val = p.ParseExpr()
		}
		f = append(f, &filter{rf.f, val})
	}
	return f
}

func Parse(s []byte) (*Template, os.Error) {
	t := new(Template)
	l := &lexer{src: s}
	l.init()
	p := &Parser{l: l, s: newScope()}

	p.Next()
	_, t.nodes = p.ParseUntil()
	t.scope = p.s

	return t, nil
}

func MustParse(s []byte) *Template {
	t, err := Parse(s)
	if err != nil {
		panic("template.MustParse error: " + err.String())
	}
	return t
}

func ParseString(s string) (*Template, os.Error) {
	return Parse([]byte(s))
}

func MustParseString(s string) *Template {
	t, err := ParseString(s)
	if err != nil {
		panic("template.MustParseString error: " + err.String())
	}
	return t
}

func ParseFile(name string) (*Template, os.Error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}
