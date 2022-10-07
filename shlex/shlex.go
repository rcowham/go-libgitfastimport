// // Package shlex provides a simple lexical analysis like Unix shell.
// Based on this one.
// Copyright (c) anmitsu <anmitsu.s@gmail.com>

// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:

// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package shlex

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
)

var (
	ErrNoClosing = errors.New("No closing quotation")
	ErrNoEscaped = errors.New("No escaped character")
)

// Tokenizer is the interface that classifies a token according to
// words, whitespaces, quotations, escapes and escaped quotations.
type Tokenizer interface {
	IsWord(rune) bool
	IsWhitespace(rune) bool
	IsQuote(rune) bool
	IsEscape(rune) bool
	IsEscapedQuote(rune) bool
}

// DefaultTokenizer implements a simple tokenizer like Unix shell.
type DefaultTokenizer struct{}

func (t *DefaultTokenizer) IsWord(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}
func (t *DefaultTokenizer) IsQuote(r rune) bool {
	switch r {
	// case '\'', '"': // For Git we only consider double quotes
	case '"':
		return true
	default:
		return false
	}
}
func (t *DefaultTokenizer) IsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}
func (t *DefaultTokenizer) IsEscape(r rune) bool {
	return r == '\\'
}
func (t *DefaultTokenizer) IsEscapedQuote(r rune) bool {
	return r == '"'
}

// Lexer represents a lexical analyzer.
type Lexer struct {
	reader          *bufio.Reader
	tokenizer       Tokenizer
	posix           bool
	whitespacesplit bool
}

// NewLexer creates a new Lexer reading from io.Reader.  This Lexer
// has a DefaultTokenizer according to posix and whitespacesplit
// rules.
func NewLexer(r io.Reader, posix, whitespacesplit bool) *Lexer {
	return &Lexer{
		reader:          bufio.NewReader(r),
		tokenizer:       &DefaultTokenizer{},
		posix:           posix,
		whitespacesplit: whitespacesplit,
	}
}

// NewGitLexer creates a new Lexer reading from io.Reader.  This Lexer
// has a DefaultTokenizer according to posix and whitespacesplit
// rules.
func NewGitLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader:          bufio.NewReader(r),
		tokenizer:       &DefaultTokenizer{},
		posix:           true,
		whitespacesplit: true,
	}
}

// NewLexerString creates a new Lexer reading from a string.  This
// Lexer has a DefaultTokenizer according to posix and whitespacesplit
// rules.
func NewLexerString(s string, posix, whitespacesplit bool) *Lexer {
	return NewLexer(strings.NewReader(s), posix, whitespacesplit)
}

// NewGitLexerString creates a new Lexer reading from a string.  This
// Lexer has a DefaultTokenizer according to posix and whitespacesplit
// rules.
func NewGitLexerString(s string) *Lexer {
	return NewLexer(strings.NewReader(s), true, true)
}

// Split splits a string according to posix or non-posix rules.
func Split(s string, posix bool) ([]string, error) {
	return NewLexerString(s, posix, true).Split()
}

// SetTokenizer sets a Tokenizer.
func (l *Lexer) SetTokenizer(t Tokenizer) {
	l.tokenizer = t
}

func (l *Lexer) Split() ([]string, error) {
	result := make([]string, 0)
	for {
		token, err := l.readToken()
		if token != "" {
			result = append(result, token)
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (l *Lexer) readToken() (string, error) {
	t := l.tokenizer
	token := ""
	quoted := false
	state := ' '
	escapedstate := ' '
scanning:
	for {
		next, _, err := l.reader.ReadRune()
		if err != nil {
			if t.IsQuote(state) {
				return token, ErrNoClosing
			} else if t.IsEscape(state) {
				return token, ErrNoEscaped
			}
			return token, err
		}

		switch {
		case t.IsWhitespace(state):
			switch {
			case t.IsWhitespace(next):
				break scanning
			case l.posix && t.IsEscape(next):
				escapedstate = 'a'
				state = next
			case t.IsWord(next):
				token += string(next)
				state = 'a'
			case t.IsQuote(next):
				if !l.posix {
					token += string(next)
				}
				state = next
			default:
				token = string(next)
				if l.whitespacesplit {
					state = 'a'
				} else if token != "" || (l.posix && quoted) {
					break scanning
				}
			}
		case t.IsQuote(state):
			quoted = true
			switch {
			case next == state:
				if !l.posix {
					token += string(next)
					break scanning
				} else {
					state = 'a'
				}
			case l.posix && t.IsEscape(next) && t.IsEscapedQuote(state):
				escapedstate = state
				state = next
			default:
				token += string(next)
			}
		case t.IsEscape(state):
			if t.IsQuote(escapedstate) && next != state && next != escapedstate {
				token += string(state)
			}
			token += string(next)
			state = escapedstate
		case t.IsWord(state):
			switch {
			case t.IsWhitespace(next):
				if token != "" || (l.posix && quoted) {
					break scanning
				}
			case l.posix && t.IsQuote(next):
				state = next
			case l.posix && t.IsEscape(next):
				escapedstate = 'a'
				state = next
			case t.IsWord(next) || t.IsQuote(next):
				token += string(next)
			default:
				if l.whitespacesplit {
					token += string(next)
				} else if token != "" {
					l.reader.UnreadRune()
					break scanning
				}
			}
		}
	}
	return token, nil
}
