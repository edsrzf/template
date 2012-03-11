package template

import (
	"reflect"
	"strconv"
	"unicode/utf8"
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
	x    Expr
	attr string
}

func (e *attrExpr) Eval(c *Context) Value {
	val := e.x.Eval(c)
	ref := val.Reflect()

	// apply attributes
	k := ref.Kind()
	if reflect.Bool <= k && k <= reflect.Complex128 {
		// invalid; do nothing
		return val
	} else if k == reflect.String {
		str := val.String()
		idx, err := strconv.Atoi(e.attr)
		if err != nil {
			// invalid; do nothing
			return val
		}
		var n, i int
		var c rune
		for i, c = range str {
			if n == idx {
				break
			}
			n++
		}
		return stringValue(str[i : i+utf8.RuneLen(c)])
	}

	ref = lookup(ref, e.attr)
	if ref.Kind() == reflect.Invalid {
		return nilValue(0)
	}
	return refToVal(ref)
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

type unaryExpr struct {
	op Token
	x  Expr
}

func (u *unaryExpr) Eval(c *Context) Value {
	val := u.x.Eval(c)
	switch u.op {
	case TokAdd:
		return val
	case TokSub:
		return intValue(-val.Int())
	case TokNot:
		return boolValue(!val.Bool())
	}
	panic("unreachable")
}

type binaryExpr struct {
	op          Token
	left, right Expr
}

func (b *binaryExpr) Eval(c *Context) Value {
	l := b.left.Eval(c)
	r := b.right.Eval(c)
	switch b.op {
	case TokAdd:
		// TODO: Floats will be truncated; other results might be unexpected
		return intValue(l.Int() + r.Int())
	case TokSub:
		return intValue(l.Int() - r.Int())
	case TokMul:
		return intValue(l.Int() * r.Int())
	case TokDiv:
		divisor := r.Int()
		if divisor == 0 {
			return stringValue("NaN")
		}
		return intValue(l.Int() / divisor)
	case TokRem:
		return intValue(l.Int() % r.Int())
	case TokAnd:
		return boolValue(l.Bool() && r.Bool())
	case TokOr:
		return boolValue(l.Bool() || r.Bool())
	case TokEqual:
		// TODO: Make sure types are comparable
		return boolValue(l == r)
	case TokNotEq:
		// TODO: Make sure types are comparable
		return boolValue(l != r)
	case TokLess:
		return boolValue(l.Int() < r.Int())
	case TokLessEq:
		return boolValue(l.Int() <= r.Int())
	case TokGreater:
		return boolValue(l.Int() > r.Int())
	case TokGreaterEq:
		return boolValue(l.Int() >= r.Int())
	}
	panic("unreachable")
}
