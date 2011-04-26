package template

import (
	"bytes"
	"unicode"
	"utf8"
)

type Token int

const (
	TokIllegal Token = iota
	TokEof

	TokIdent  // for
	TokInt    // 12345
	TokFloat  // 123.45e2
	TokString // 'abc' or "abc"

	TokText // text not inside a tag

	TokTagStart // {%
	TokTagEnd   // %}
	TokVarStart // {{
	TokVarEnd   // }}

	TokAdd // +
	TokSub // -
	TokMul // *
	TokDiv // /
	TokRem // %

	TokAnd // and
	TokOr  // or
	TokNot // not

	TokEqual     // ==
	TokLess      // <
	TokLessEq    // <=
	TokGreater   // >
	TokGreaterEq // >=
	TokNotEq     // !=

	TokDot   // .
	TokBar   // |
	TokColon // :
)

var tokStrings = map[Token]string{
	TokIllegal:  "illegal",
	TokEof:      "eof",
	TokIdent:    "ident",
	TokInt:      "int",
	TokFloat:    "float",
	TokString:   "string",
	TokText:     "text",
	TokTagStart: "{%",
	TokTagEnd:   "%}",
	TokVarStart: "{{",
	TokVarEnd:   "}}",
	TokDot:      ".",
	TokBar:      "|",
	TokColon:    ":",
}

func (t Token) String() string {
	return tokStrings[t]
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

func (l *lexer) scan() (Token, []byte) {
scanAgain:
	if !l.insideTag && l.ch != '{' {
		lit := l.scanText()
		return TokText, lit
	}
	l.insideTag = true
	l.consumeWhitespace()

	pos := l.offset
	tok := TokIllegal

	switch ch := l.ch; {
	case unicode.IsLetter(ch), l.ch == '_':
		tok = l.scanIdent()
	case unicode.IsDigit(ch):
		tok = l.scanNumber()
	case ch == '-', ch == '+':
		tok = l.scanNumber()
	case ch == '|':
		tok = TokBar
		l.next()
	case ch == '.':
		tok = TokDot
		l.next()
	case ch == ':':
		tok = TokColon
		l.next()
	default:
		l.next()
		switch ch {
		case -1:
			tok = TokEof
		case '{':
			switch l.ch {
			case '%':
				tok = TokTagStart
			case '{':
				tok = TokVarStart
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
				goto scanAgain
				return l.scan()
			}
		case '%':
			if l.ch != '}' {
				goto illegal
			}
			l.insideTag = false
			tok = TokTagEnd
		case '}':
			if l.ch != '}' {
				goto illegal
			}
			l.insideTag = false
			tok = TokVarEnd
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

func (l *lexer) scanIdent() Token {
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.next()
	}
	return TokIdent
}

func (l *lexer) scanNumber() Token {
	tok := TokInt
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
			tok = TokFloat
		}
		if l.ch == 'e' || l.ch == 'E' {
			seenExponent = true
			tok = TokFloat
		}
		l.next()
	}
	return tok
}

func (l *lexer) scanString(start byte) Token {
	// ' or " already consumed
	pos := bytes.IndexByte(l.src[l.offset:], start)
	if pos < 0 {
		panic("String not terminated")
	}
	l.offset += pos
	l.width = 1
	l.next()
	return TokString
}
