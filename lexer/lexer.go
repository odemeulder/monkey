package lexer

import (
	"demeulder.us/monkey/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under consideration
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhiteSpace()
	switch l.ch {
	case '=':
		tok = l.twoCharToken()
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = l.twoCharToken()
	case '-':
		tok = l.twoCharToken()
	case '*':
		tok = l.twoCharToken()
	case '/':
		tok = l.twoCharToken()
	case '!':
		tok = l.twoCharToken()
	case '<':
		tok = l.twoCharToken()
	case '>':
		tok = l.twoCharToken()
	case '&':
		tok = l.twoCharToken()
	case '|':
		tok = l.twoCharToken()
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '"':
		tok.Literal = l.readString()
		tok.Type = token.STRING
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}

	}
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) twoCharToken() token.Token {
	var tok token.Token
	if l.peekChar() != 0 {
		literal := string(l.ch) + string(l.peekChar())
		switch literal {
		case "==":
			l.readChar()
			return token.Token{Type: token.EQ, Literal: literal}
		case "!=":
			l.readChar()
			return token.Token{Type: token.NOT_EQ, Literal: literal}
		case "<=":
			l.readChar()
			return token.Token{Type: token.LT_EQ, Literal: literal}
		case ">=":
			l.readChar()
			return token.Token{Type: token.GT_EQ, Literal: literal}
		case "&&":
			l.readChar()
			return token.Token{Type: token.AND, Literal: literal}
		case "||":
			l.readChar()
			return token.Token{Type: token.OR, Literal: literal}
		case "++":
			l.readChar()
			return token.Token{Type: token.PLUSPLUS, Literal: literal}
		case "--":
			l.readChar()
			return token.Token{Type: token.MINUSMINUS, Literal: literal}
		case "+=":
			l.readChar()
			return token.Token{Type: token.ASSIGNPLUS, Literal: literal}
		case "-=":
			l.readChar()
			return token.Token{Type: token.ASSIGNMINUS, Literal: literal}
		case "*=":
			l.readChar()
			return token.Token{Type: token.ASSIGNTIMES, Literal: literal}
		case "/=":
			l.readChar()
			return token.Token{Type: token.ASSIGNSLASH, Literal: literal}
		case "&=":
			l.readChar()
			return token.Token{Type: token.ASSIGNAND, Literal: literal}
		case "|=":
			l.readChar()
			return token.Token{Type: token.ASSIGNOR, Literal: literal}
		}
	}
	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case '!':
		tok = newToken(token.BANG, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	}
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	l.readChar() // first double quote
	position := l.position
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipWhiteSpace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		l.readChar()
	}
}
