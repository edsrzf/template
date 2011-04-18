package template

import (
	"bytes"
	"unicode"
	"utf8"
)

type token int

const (
	tokIllegal token = iota
	tokEof

	tokIdent  // for
	tokInt    // 12345
	tokFloat  // 123.45e2
	tokString // 'abc' or "abc"

	tokText // text not inside a tag

	tokBlockTagStart   // {%
	tokBlockTagEnd     // %}
	tokVarTagStart     // {{
	tokVarTagEnd       // }}

	tokDot      // .
	tokFilter   // |
	tokArgument // :
)

var tokStrings = map[token]string{
	tokIllegal:         "illegal",
	tokEof:             "eof",
	tokIdent:           "ident",
	tokInt:             "int",
	tokFloat:           "float",
	tokString:          "string",
	tokText:            "text",
	tokBlockTagStart:   "{%",
	tokBlockTagEnd:     "%}",
	tokVarTagStart:     "{{",
	tokVarTagEnd:       "}}",
	tokDot:             ".",
	tokFilter:          "|",
	tokArgument:        ":",
}

type lexer struct {
	src       []byte
	offset    int
	ch        int
	width     int
	insideTag bool
}

func (l *lexer) init() {
	l.ch, l.width = utf8.DecodeRune(l.src)
}

func (l *lexer) next() {
	l.offset += l.width
	if l.offset < len(l.src) {
		r, w := int(l.src[l.offset]), 1
		if r >= utf8.RuneSelf {
			r, w = utf8.DecodeRune(l.src[l.offset:])
		}
		l.width = w
		l.ch = r
	} else {
		l.width = 0
		l.ch = -1
	}
}

func (l *lexer) scan() (token, []byte) {
	if !l.insideTag && l.ch != '{' {
		lit := l.scanText()
		return tokText, lit
	}
	l.insideTag = true
	// Another divergence from Django is that we allow spaces between all tokens. So
	// {{ var . 1 }}
	// is a valid template and does what you would expect. Django doesn't allow this.
	l.consumeWhitespace()

	pos := l.offset
	tok := tokIllegal

	switch ch := l.ch; {
	case unicode.IsLetter(ch), l.ch == '_':
		tok = l.scanIdent()
	case unicode.IsDigit(ch):
		tok = l.scanNumber()
	case ch == '-', ch == '+':
		tok = l.scanNumber()
	case ch == '|':
		tok = tokFilter
		l.next()
	case ch == '.':
		tok = tokDot
		l.next()
	case ch == ':':
		tok = tokArgument
		l.next()
	default:
		l.next()
		switch ch {
		case -1:
			tok = tokEof
		case '{':
			switch l.ch {
			case '%':
				tok = tokBlockTagStart
			case '{':
				tok = tokVarTagStart
			case '#':
				// start of a comment; scan until the end
				l.next()
				for {
					for l.ch != '#' {
						l.next()
					}
					l.next()
					if l.ch == '}' {
						l.next()
						break
					}
				}
				// TODO: probably shouldn't recurse
				return l.scan()
			}
		case '%':
			if l.ch != '}' {
				goto illegal
			}
			l.insideTag = false
			tok = tokBlockTagEnd
		case '}':
			if l.ch != '}' {
				goto illegal
			}
			l.insideTag = false
			tok = tokVarTagEnd
		case '\'', '"':
			tok = l.scanString(byte(ch))
			return tok, l.src[pos+1 : l.offset-1]
		default:
		illegal:
			panic("illegal character")
		}
		l.next()
	}
	return tok, l.src[pos:l.offset]
}

func (l *lexer) consumeWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.next()
	}
}

// There's a known incompatibility here for a template such as:
// {%{##}}}
// Django outputs {%}}. It has to see the opening and the closing symbols for it to be considered a tag.
// Godjan will return an error on this input. Disallowing this unlikely case makes
// lexing much easier and more efficient because we don't have to look ahead.
func (l *lexer) scanText() []byte {
	n := len(l.src) - 1
	off := l.offset
	pos := 0
consume:
	for {
		pos = bytes.IndexByte(l.src[off:], '{')
		if pos < 0 || off+pos >= n {
			break
		}
		off += pos
		switch l.src[off+1] {
		case '%', '{', '#':
			break consume
		}
		off++
	}
	var lit []byte
	if pos < 0 || pos == n {
		lit = l.src[l.offset:]
		l.offset = len(l.src)
		l.width = 0
		l.ch = -1
	} else {
		l.width = 1
		l.ch = '{'
		lit = l.src[l.offset:off]
		l.offset = off
		if l.src[l.offset] != byte(l.ch) {
			panic("Something's wrong")
		}
	}
	l.insideTag = true
	return lit
}

func (l *lexer) scanIdent() token {
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.next()
	}
	return tokIdent
}

func (l *lexer) scanNumber() token {
	tok := tokInt
	seenDecimal := false
	seenExponent := false

	if l.ch == '-' || l.ch == '+' {
		l.next()
	}
	for l.ch >= '0' && l.ch <= '9' ||
		!seenDecimal && !seenExponent && l.ch == '.' ||
		!seenExponent && (l.ch == 'e' || l.ch == 'E') {
		if l.ch == '.' {
			seenDecimal = true
			tok = tokFloat
		}
		if l.ch == 'e' || l.ch == 'E' {
			seenExponent = true
			tok = tokFloat
		}
		l.next()
	}
	return tok
}

func (l *lexer) scanString(start byte) token {
	// ' or " already consumed
	pos := bytes.IndexByte(l.src[l.offset:], start)
	if pos < 0 {
		panic("String not terminated")
	}
	l.offset += pos
	l.width = 1
	l.next()
	return tokString
}
