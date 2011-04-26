package template

import (
	"reflect"
	"strconv"
	"utf8"
)

// Expr represents an expression that can be evaluated at runtime.
type Expr interface {
	Eval(c *Context) Value
}

type constExpr struct {
	v Value
}

func (e constExpr) Eval(c *Context) Value { return e.v }

type attrExpr struct {
	x     Expr
	attrs []string
}

func (e *attrExpr) Eval(c *Context) Value {
	val := e.x.Eval(c)
	ref := val.Reflect()

	// apply attributes
	k := ref.Kind()
	if reflect.Bool <= k && k <= reflect.Complex128 {
		return val
	} else if k == reflect.String {
		if len(e.attrs) > 0 {
			str := val.String()
			idx, err := strconv.Atoi(e.attrs[0])
			if err != nil {
				return nilValue(0)
			}
			var n, i, c int
			for i, c = range str {
				if n == idx {
					break
				}
				n++
			}
			return stringValue(str[i : i+utf8.RuneLen(c)])
		} else {
			return val
		}
	}
	return getVal(ref, e.attrs)
}

type filterExpr struct {
	x       Expr
	filters []*filter
}

func (e *filterExpr) Eval(c *Context) Value {
	val := e.x.Eval(c)
	// apply filters
	for _, f := range e.filters {
		val = f.f(constExpr{val}, c, f.args)
	}
	return val
}

type equalExpr struct {
	left, right Expr
}

func (e *equalExpr) Eval(c *Context) bool {
	// TODO: Make sure types are comparable
	l := e.left.Eval(c)
	r := e.right.Eval(c)
	return l == r
}

type nequalExpr struct {
	left, right Expr
}

func (n *nequalExpr) Eval(c *Context) bool {
	// TODO: Make sure types are comparable
	l := n.left.Eval(c)
	r := n.right.Eval(c)
	return l != r
}

type notExpr struct {
	x Expr
}

func (n *notExpr) Eval(c *Context) bool {
	return !n.x.Eval(c).Bool()
}

type andExpr struct {
	left, right Expr
}

func (a *andExpr) Eval(c *Context) bool {
	return a.left.Eval(c).Bool() && a.right.Eval(c).Bool()
}

type orExpr struct {
	left, right Expr
}

func (o *orExpr) Eval(c *Context) bool {
	return o.left.Eval(c).Bool() || o.right.Eval(c).Bool()
}
