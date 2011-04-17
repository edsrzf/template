package template

import (
	"io"
)

type ifTag struct {
	cond     condition
	ifNode   Renderer
	elseNode Renderer
}

func (i *ifTag) Render(wr io.Writer, s Stack) {
	if i.cond.eval(s) {
		i.ifNode.Render(wr, s)
	} else {
		i.elseNode.Render(wr, s)
	}
}

type condition interface {
	eval(s Stack) bool
}

type equal struct {
	left, right *variable
}

func (e *equal) eval(s Stack) bool {
	// TODO: Make sure types are comparable
	l := e.left.value(s)
	r := e.right.value(s)
	return l == r
}

type nequal struct {
	left, right *variable
}

func (n *nequal) eval(s Stack) bool {
	// TODO: Make sure types are comparable
	l := n.left.value(s)
	r := n.right.value(s)
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
