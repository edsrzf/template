package template

import (
	"io"
)

type ifTag struct {
	cond     Value
	ifNode   Node
	elseNode Node
}

func parseIf(p *Parser) Node {
	tag := new(ifTag)
	tag.cond = parseCondition(p)
	p.Expect(TokTagEnd)
	var tok string
	tok, tag.ifNode = p.ParseUntil("elif", "else", "endif")
	for tok != "endif" {
		switch tok {
		case "elif":
			tag.elseNode = parseIf(p)
			return tag
		case "else":
			tok, tag.elseNode = p.ParseUntil("endif")
		default:
			p.Error("unterminated if tag")
		}
	}
	return tag
}

func (i *ifTag) Render(wr io.Writer, s Stack) {
	if i.cond.Bool(s) {
		i.ifNode.Render(wr, s)
	} else if i.elseNode != nil {
		i.elseNode.Render(wr, s)
	}
}

func parseCondition(p *Parser) Value {
	return p.ParseExpr()
}

type equal struct {
	left, right *expr
}

func (e *equal) eval(s Stack) bool {
	// TODO: Make sure types are comparable
	l := e.left.String(s)
	r := e.right.String(s)
	return l == r
}

type nequal struct {
	left, right *expr
}

func (n *nequal) eval(s Stack) bool {
	// TODO: Make sure types are comparable
	l := n.left.String(s)
	r := n.right.String(s)
	return l != r
}

type not struct {
	x Value
}

func (n *not) eval(s Stack) bool {
	return !n.x.Bool(s)
}

type and struct {
	left, right Value
}

func (a *and) eval(s Stack) bool {
	return a.left.Bool(s) && a.right.Bool(s)
}

type or struct {
	left, right Value
}

func (o *or) eval(s Stack) bool {
	return o.left.Bool(s) || o.right.Bool(s)
}
