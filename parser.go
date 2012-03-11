package template

import (
	"fmt"
	"io/ioutil"
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

func (p *Parser) Scope() *scope {
	return p.s
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
	e := p.parseBinaryExpr(1)
	if f := p.parseFilters(); len(f) > 0 {
		e = &filterExpr{e, f}
	}
	return e
}

func (p *Parser) parseOperand() Expr {
	var ret Expr
	switch p.tok {
	case TokInt:
		i, err := strconv.ParseInt(string(p.lit), 10, 64)
		if err != nil {
			p.Error("internal int error: %s", err)
		}
		ret = constExpr{intValue(i)}
		p.Next()
	case TokFloat:
		f, err := strconv.ParseFloat(string(p.lit), 64)
		if err != nil {
			p.Error("internal float error: %s", err)
		}
		ret = constExpr{floatValue(f)}
		p.Next()
	case TokString:
		ret = constExpr{stringValue(p.lit)}
		p.Next()
	case TokIdent:
		ret = p.parseVar()
	default:
		p.Error("unexpected Token %s", p.tok)
	}
	return ret
}

func (p *Parser) parsePrimaryExpr() Expr {
	x := p.parseOperand()
L:
	for {
		switch p.tok {
		case TokDot:
			p.Next()
			attr := string(p.lit)
			if p.tok == TokInt {
				p.Next()
			} else {
				p.Expect(TokIdent)
			}
			x = &attrExpr{x, attr}
		default:
			break L
		}
	}
	return x
}

func (p *Parser) parseUnaryExpr() Expr {
	switch p.tok {
	case TokAdd, TokSub, TokNot:
		op := p.tok
		p.Next()
		x := p.parseUnaryExpr()
		return &unaryExpr{op, x}
	}
	return p.parsePrimaryExpr()
}

func (p *Parser) parseBinaryExpr(prec1 int) Expr {
	x := p.parseUnaryExpr()
	for prec := p.tok.Precedence(); prec >= prec1; prec-- {
		for p.tok.Precedence() == prec {
			op := p.tok
			p.Next()
			y := p.parseBinaryExpr(prec + 1)
			x = &binaryExpr{op, x, y}
		}
	}
	return x
}

func (p *Parser) parseVar() Variable {
	return p.s.Lookup(p.Expect(TokIdent))
}

func (p *Parser) parseFilters() []*filter {
	var f []*filter
	for p.tok == TokBar {
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
			args = p.tok == TokColon
		case NoArg:
			if p.tok == TokColon {
				p.Error("filter accepts no arguments")
			}
		}
		if args {
			p.Expect(TokColon)
			val = p.ParseExpr()
		}
		f = append(f, &filter{rf.f, val})
	}
	return f
}

func Parse(s []byte) (*Template, error) {
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
		panic("template.MustParse error: " + err.Error())
	}
	return t
}

func ParseString(s string) (*Template, error) {
	return Parse([]byte(s))
}

func MustParseString(s string) *Template {
	t, err := ParseString(s)
	if err != nil {
		panic("template.MustParseString error: " + err.Error())
	}
	return t
}

func ParseFile(name string) (*Template, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}
