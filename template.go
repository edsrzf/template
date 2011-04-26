package template

import (
	"io"
	"reflect"
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
				stack[v] = refToVal(reflect.ValueOf(val))
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

type varTag struct {
	e Expr
}

func (v varTag) Render(wr io.Writer, c *Context) { v.e.Eval(c).Render(wr, c) }

type Template struct {
	scope *scope
	nodes NodeList
}

func (t *Template) Execute(wr io.Writer, vars map[string]interface{}) {
	c := newContext(t.scope, vars)
	t.render(wr, c)
}

func (t *Template) Render(wr io.Writer, c *Context) {
	// we have to create a new Context that matches this template's
	// stack layout.
	c = newContext(t.scope, c.vars)
	t.render(wr, c)
}

func (t *Template) render(wr io.Writer, c *Context) {
	t.scope.levels[0].init.Render(wr, c)
	t.nodes.Render(wr, c)
}
