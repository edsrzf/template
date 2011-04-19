package template

import (
	"io"
)

type ifTag struct {
	cond     Value
	ifNode   Node
	elseNode Node
}

func parseIf(p *parser) Node {
	tag := new(ifTag)
	tag.cond = parseCondition(p)
	p.Expect(tokBlockTagEnd)
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

type condition interface {
	eval(s Stack) bool
}

func parseCondition(p *parser) Value {
	return p.parseExpr()
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
	inner condition
}

func (n *not) eval(s Stack) bool {
	return !n.inner.eval(s)
}

type and struct {
	left, right condition
}

func (a *and) eval(s Stack) bool {
	return a.left.eval(s) && a.right.eval(s)
}

type or struct {
	left, right condition
}

func (o *or) eval(s Stack) bool {
	return o.left.eval(s) || o.right.eval(s)
}
