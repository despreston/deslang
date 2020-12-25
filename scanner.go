package deslang

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

// For lexing an io.Reader. Groups characters into tokens.
type Scanner struct {
	errh    errorHandler
	source  *bufio.Reader // source code to scan
	tokens  []Token       // tokens seen
	currLex []byte        // partial lexeme
	line    int           // current line
	ch      byte          // most recently read character
}

var keywords = map[string]tokentype{
	"and":    _and,
	"else":   _else,
	"false":  _false,
	"for":    _for,
	"fun":    _fun,
	"if":     _if,
	"or":     _or,
	"print":  _print,
	"return": _return,
	"true":   _true,
	"var":    _var,
	"while":  _while,
}

func NewScanner(errh errorHandler) *Scanner {
	return &Scanner{
		errh: errh,
		line: 1,
	}
}

func (s *Scanner) reset() {
	s.tokens = []Token{}
	s.line = 1
}

// If reading the next byte fails, Scan will return an error. All syntax errors
// are reported via the errorHandler.
func (s *Scanner) Scan(src io.Reader) ([]Token, error) {
	s.reset()
	s.source = bufio.NewReader(src)

	for {
		s.currLex = []byte{}

		if err := s.next(); err != nil {
			if err == io.EOF {
				s.addToken(_eof, nil)
			}
			return s.tokens, err
		}

		s.parseCh()
	}
}

func (s *Scanner) addToken(ttype tokentype, lit []byte) {
	t := Token{
		Type:    ttype,
		Lexeme:  s.currLex,
		Literal: lit,
		Line:    s.line,
	}

	s.tokens = append(s.tokens, t)
}

// Read the next character and store the byte in s.ch. Append the character to
// s.currLex.
func (s *Scanner) next() error {
	b, err := s.source.ReadByte()
	if err != nil {
		return err
	}

	s.ch = b
	s.currLex = append(s.currLex, s.ch)

	return nil
}

func (s *Scanner) peek() byte {
	b, _ := s.source.Peek(1)
	return b[0]
}

// Return true if the previous character in the current lexeme matches the
// expected byte. Advances s.source via s.next() if it's a match.
func (s *Scanner) match(expected byte) bool {
	if s.peek() == expected {
		s.next()
		return true
	}
	return false
}

// Consumes a string including the closing quotation mark.
func (s *Scanner) string() {
	// Move to first character inside quote.
	s.next()

	for s.ch != '"' {
		if err := s.next(); err == io.EOF {
			s.errh(s.line, "", "Unterminated string")
			return
		}
	}

	s.addToken(_string, bytes.Trim(s.currLex, "\""))
}

// Consumes a number
func (s *Scanner) number() {
	for unicode.IsNumber(rune(s.peek())) {
		if err := s.next(); err != nil {
			return
		}
	}

	// In order to handle decimals, if there's a period, consume it, then keep
	// consuming any remaining digits.
	if s.peek() == '.' {
		s.next()
		for unicode.IsNumber(rune(s.peek())) {
			if err := s.next(); err != nil {
				break
			}
		}
	}

	s.addToken(_number, s.currLex)
}

func (s *Scanner) identifier() {
	r := rune(s.peek())

	for unicode.IsLetter(r) || unicode.IsNumber(r) {
		s.next()
		r = rune(s.peek())
	}

	// Check if it's a reserved keyword.
	if tokenType, has := keywords[string(s.currLex)]; has {
		s.addToken(tokenType, nil)
	} else {
		s.addToken(_identifier, nil)
	}
}

// Parse s.ch
func (s *Scanner) parseCh() {
	switch s.ch {
	case '(':
		s.addToken(_left_paren, nil)
	case ')':
		s.addToken(_right_paren, nil)
	case '{':
		s.addToken(_left_brace, nil)
	case '}':
		s.addToken(_right_brace, nil)
	case ',':
		s.addToken(_comma, nil)
	case '-':
		s.addToken(_minus, nil)
	case '+':
		s.addToken(_plus, nil)
	case ';':
		s.addToken(_semicolon, nil)
	case '*':
		s.addToken(_star, nil)
	case '!':
		if s.match('=') {
			s.addToken(_bang_equal, nil)
		} else {
			s.addToken(_bang, nil)
		}
	case '=':
		if s.match('=') {
			s.addToken(_equal_equal, nil)
		} else {
			s.addToken(_equal, nil)
		}
	case '>':
		if s.match('=') {
			s.addToken(_greater_equal, nil)
		} else {
			s.addToken(_greater, nil)
		}
	case '<':
		if s.match('=') {
			s.addToken(_less_equal, nil)
		} else {
			s.addToken(_less, nil)
		}
	case '/':
		// If it's a comment, clear this entire line.
		if s.match('/') {
			for s.peek() != '\n' {
				if err := s.next(); err != nil {
					// Most likely an EOF err but doesn't hurt to break on any error.
					break
				}
			}
		} else {
			s.addToken(_slash, nil)
		}
	case ' ', '\r', '\t':
		return
	case '\n':
		s.line++
	case '"':
		s.string()
	case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		s.number()
	default:
		if unicode.IsLetter(rune(s.ch)) {
			s.identifier()
		} else {
			s.errh(s.line, "", "Unexpected character")
		}
	}
}
