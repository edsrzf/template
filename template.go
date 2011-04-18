package template

import (
	"io"
	"reflect"
)

type Context map[string]interface{}

type Stack []value

type scope struct {
	levels []map[string]int
	// the greatest number of variables this scope and its children can hold
	maxLen int
}

func newScope() *scope {
	return &scope{levels: []map[string]int{{}}}
}

func (s *scope) top() map[string]int {
	return s.levels[len(s.levels)-1]
}

// Push creates a new scope level
func (s *scope) Push() {
	s.levels = append(s.levels, map[string]int{})
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

// Lookup returns the variable number for the given variable name from the
// most specific possible scope.
// If the name cannot be found, it is inserted into the broadest scope and
// the new variable number is returned.
func (s *scope) Lookup(name string) int {
	l := len(s.levels)
	for i := l - 1; i >= 0; i-- {
		if v, ok := s.levels[i][name]; ok {
			return v
		}
	}
	v := s.maxLen
	//println("inserting", name, "at level 0 as", v)
	s.levels[0][name] = v
	s.maxLen++
	return v
}

// Insert creates a new variable at the top scope and returns the
// new variable number.
// If the variable already exists in that scope, the old variable number is
// returned.
func (s *scope) Insert(name string) int {
	l := len(s.levels)
	v, ok := s.levels[l-1][name]
	if ok {
		return v
	}
	v = s.len()
	//println("inserting", name, "at level", l-1, "as", v)
	s.levels[l-1][name] = v
	s.maxLen++
	return v
}

type Node interface {
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

// expr represents a value with possible filters
type expr struct {
	v       valuer
	filters []*filter
}

func (e *expr) Render(wr io.Writer, s Stack) {
	v := e.value(s)
	str := valueAsString(v)
	wr.Write([]byte(str))
}

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
