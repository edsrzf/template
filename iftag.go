package template

import (
	"io"
)

type ifTag struct {
	cond     Expr
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

func (i *ifTag) Render(wr io.Writer, c *Context) {
	if i.cond.Eval(c).Bool() {
		i.ifNode.Render(wr, c)
	} else if i.elseNode != nil {
		i.elseNode.Render(wr, c)
	}
}

func parseCondition(p *Parser) Expr {
	return p.ParseExpr()
}

type equal struct {
	left, right Expr
}

func (e *equal) eval(c *Context) bool {
	// TODO: Make sure types are comparable
	l := e.left.Eval(c).String()
	r := e.right.Eval(c).String()
	return l == r
}

type nequal struct {
	left, right Expr
}

func (n *nequal) eval(c *Context) bool {
	// TODO: Make sure types are comparable
	l := n.left.Eval(c).String()
	r := n.right.Eval(c).String()
	return l != r
}

type not struct {
	x Value
}

func (n *not) eval(c *Context) bool {
	return !n.x.Bool()
}

type and struct {
	left, right Value
}

func (a *and) eval(c *Context) bool {
	return a.left.Bool() && a.right.Bool()
}

type or struct {
	left, right Value
}

func (o *or) eval(c *Context) bool {
	return o.left.Bool() || o.right.Bool()
}
