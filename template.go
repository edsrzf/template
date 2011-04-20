package template

import (
	"io"
	"reflect"
	"strconv"
	"utf8"
)

type Context struct {
	vars  map[string]interface{}
	stack []Value
}

func newContext(s *scope, vars map[string]interface{}) *Context {
	stack := make([]Value, s.maxLen)
	if vars != nil {
		for k, v := range s.top() {
			if val, ok := vars[k]; ok {
				stack[v] = refToVal(reflect.NewValue(val))
			}
		}
	}
	return &Context{vars, stack}
}

type scopeLevel struct {
	named map[string]Variable
	// This node initializes anonymous variables.
	// It's up to the popper of this level to make sure these run.
	init initNode
}

type scope struct {
	levels []*scopeLevel
	// the greatest number of Variables this scope and its children can hold
	maxLen int
}

func newScope() *scope {
	s := new(scope)
	level := &scopeLevel{map[string]Variable{}, map[Variable]Value{}}
	s.levels = []*scopeLevel{level}
	return s
}

func (s *scope) top() map[string]Variable {
	return s.levels[len(s.levels)-1].named
}

// Push creates a new scope level
func (s *scope) Push() {
	level := &scopeLevel{map[string]Variable{}, map[Variable]Value{}}
	s.levels = append(s.levels, level)
}

// Pop removes the top scope
func (s *scope) Pop() Node {
	if len(s.levels) == 1 {
		return initNode(nil)
	}
	if s.maxLen < s.len() {
		s.maxLen = s.len()
	}
	node := s.levels[len(s.levels)-1].init
	s.levels = s.levels[:len(s.levels)-1]
	return node
}

// len returns the number of variables that need to be allocated for the
// current stack of scopes.
func (s *scope) len() int {
	l := 0
	for _, level := range s.levels {
		l += len(level.named) + len(level.init)
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
		if v, ok := s.levels[i].named[name]; ok {
			return v
		}
	}
	v := Variable(s.maxLen)
	s.levels[0].named[name] = v
	s.maxLen++
	return v
}

// Insert creates a new Variable at the top scope and returns it.
// If the Variable already exists in that scope, the existing Variable is
// returned.
func (s *scope) Insert(name string) Variable {
	l := len(s.levels)
	v, ok := s.levels[l-1].named[name]
	if ok {
		return v
	}
	v = Variable(s.len())
	s.levels[l-1].named[name] = v
	s.maxLen++
	return v
}

// Anonymous creates a new anonymous Variable at the top scope and returns it.
// This is useful for Nodes that might Render more than once and want to
// store state between Renders.
func (s *scope) Anonymous(init Value) Variable {
	v := Variable(s.len())
	level := s.levels[len(s.levels)-1]
	level.init[v] = init
	s.maxLen++
	return v
}

type initNode map[Variable]Value

func (i initNode) Render(wr io.Writer, c *Context) {
	for v, val := range i {
		v.Set(val, c)
	}
}

// A Node represents a part of the Template, such as a tag or a block of text.
type Node interface {
	// Render evaluates the node with the given Context and writes the result to
	// wr.
	// Render should be reentrant. If the Node needs to store state, it should
	// allocate a Variable on the stack during parsing and use that Variable.
	Render(wr io.Writer, c *Context)
}

type NodeList []Node

func (l NodeList) Render(wr io.Writer, c *Context) {
	for _, r := range l {
		r.Render(wr, c)
	}
}

type printLit []byte

func (p printLit) Render(wr io.Writer, c *Context) { wr.Write([]byte(p)) }

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
	v       Value
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
func (e *expr) Eval(c *Context) Value {
	val := e.v.Reflect(c)

	// apply attributes
	var ret Value
	k := val.Kind()
	if reflect.Bool <= k && k <= reflect.Complex128 {
		ret = e.v
	} else if k == reflect.String {
		if len(e.attrs) > 0 {
			str := e.v.String(c)
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
			ret = stringValue(str[i : i+utf8.RuneLen(c)])
		} else {
			ret = e.v
		}
	} else {
		ret = getVal(val, e.attrs)
	}

	// apply filters
	for _, f := range e.filters {
		ret = f.f(ret, c, f.args)
	}
	return ret
}

func (e *expr) Bool(c *Context) bool             { return e.Eval(c).Bool(c) }
func (e *expr) Int(c *Context) int64             { return e.Eval(c).Int(c) }
func (e *expr) String(c *Context) string         { return e.Eval(c).String(c) }
func (e *expr) Uint(c *Context) uint64           { return e.Eval(c).Uint(c) }
func (e *expr) Reflect(c *Context) reflect.Value { return e.Eval(c).Reflect(c) }

func (e *expr) Render(wr io.Writer, c *Context) { renderValue(e, wr, c) }

func (t *Template) Execute(wr io.Writer, vars map[string]interface{}) {
	c := newContext(t.scope, vars)
	t.Render(wr, c)
}

func (t *Template) Render(wr io.Writer, c *Context) {
	t.scope.levels[0].init.Render(wr, c)
	t.nodes.Render(wr, c)
}
