package template

import (
	"io"
	"reflect"
	"strconv"
	"utf8"
)

type Context map[string]interface{}

type Stack []Value

type scope struct {
	levels []map[string]Variable
	// the greatest number of Variables this scope and its children can hold
	maxLen int
}

func newScope() *scope {
	return &scope{levels: []map[string]Variable{{}}}
}

func (s *scope) top() map[string]Variable {
	return s.levels[len(s.levels)-1]
}

// Push creates a new scope level
func (s *scope) Push() {
	s.levels = append(s.levels, map[string]Variable{})
}

// Pop removes the top scope
func (s *scope) Pop() {
	if len(s.levels) == 1 {
		return
	}
	if s.maxLen < s.len() {
		s.maxLen = s.len()
	}
	s.levels = s.levels[:len(s.levels)-1]
}

func (s *scope) len() int {
	l := 0
	for _, level := range s.levels {
		l += len(level)
	}
	return l
}

// Lookup returns the Variable for the given name from the most specific
// possible scope.
// If the name cannot be found, it is inserted into the broadest scope and
// the new Variable is returned.
func (s *scope) Lookup(name string) Variable {
	l := len(s.levels)
	for i := l - 1; i >= 0; i-- {
		if v, ok := s.levels[i][name]; ok {
			return v
		}
	}
	v := Variable(s.maxLen)
	s.levels[0][name] = v
	s.maxLen++
	return v
}

// Insert creates a new Variable at the top scope and returns it.
// If the Variable already exists in that scope, the existing Variable is
// returned.
func (s *scope) Insert(name string) Variable {
	l := len(s.levels)
	v, ok := s.levels[l-1][name]
	if ok {
		return v
	}
	v = Variable(s.len())
	s.levels[l-1][name] = v
	s.maxLen++
	return v
}

// A Node represents a part of the Template, such as a tag or a block of text.
type Node interface {
	// Render evaluates the node with the given Stack and writes the result to
	// wr.
	// Render should be reentrant. If the Node needs to store state, it should
	// allocate a Variable on the stack during parsing and use that Variable.
	// The parameter s should only be used when calling a Valuer's Value method.
	// Nodes should not access s directly.
	Render(wr io.Writer, s Stack)
}

type NodeList []Node

func (l NodeList) Render(wr io.Writer, s Stack) {
	for _, r := range l {
		r.Render(wr, s)
	}
}

type printLit []byte

func (p printLit) Render(wr io.Writer, s Stack) { wr.Write([]byte(p)) }

type Template struct {
	scope *scope
	nodes NodeList
}

// expr represents a Value with possible attributes and filters.
// Attributes work like this:
// For the expression "a.b"
// - If a is a map[string]T, this is treated as a["b"]
// - If a is a struct or pointer to a struct, this is treated as a.b
// - If a is a map[numeric]T, slice, array, or pointer to an array, this is treated as a[b]
// - If the above all fail, this is treated as a method call a.b()
type expr struct {
	v       Valuer
	attrs   []string
	filters []*filter
}

// Returns a true or false Value for an expression.
// false Values include:
// - nil
// - A slice, map, channel, array, or pointer to an array with zero len
// - The bool Value false
// - An empty string
// - Zero of any numeric type
func (e *expr) eval(s Stack) bool {
	v := e.Value(s)
	return valueAsBool(v)
}

func (e *expr) Value(s Stack) Value {
	val := e.v.Value(s)

	// apply attributes
	var ret Value
	switch val := val.(type) {
	case bool, float32, float64, complex64, complex128, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr, nil:
		ret = val
	case string:
		if len(e.attrs) > 0 {
			idx, err := strconv.Atoi(e.attrs[0])
			if err != nil {
				return nil
			}
			var n, i, c int
			for i, c = range val {
				if n == idx {
					break
				}
				n++
			}
			ret = val[i : i+utf8.RuneLen(c)]
			break
		}
		ret = val
	case reflect.Value:
		ret = getVal(val, e.attrs)
	}

	// apply filters
	for _, f := range e.filters {
		ret = f.f(ret, s, f.args)
	}
	return ret
}

func (e *expr) Render(wr io.Writer, s Stack) { renderValuer(e, wr, s) }

func (t *Template) Execute(wr io.Writer, c Context) {
	s := make(Stack, t.scope.maxLen)
	if c != nil {
		for k, v := range t.scope.top() {
			if val, ok := c[k]; ok {
				switch val.(type) {
				case bool, float32, float64, complex64, complex128, int, int8, int16,
					int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, string, nil:
				default:
					val = reflect.NewValue(val)
				}
				s[v] = val
			}
		}
	}
	t.nodes.Render(wr, s)
}
