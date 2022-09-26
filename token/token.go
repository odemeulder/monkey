package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL    = "ILLEGAL"
	EOF        = "EOF"
	IDENT      = "IDENT"
	INT        = "INT"
	ASSIGN     = "="
	PLUS       = "+"
	MINUS      = "-"
	BANG       = "!"
	ASTERISK   = "*"
	SLASH      = "/"
	COMMA      = ","
	SEMICOLON  = ";"
	LPAREN     = "("
	RPAREN     = ")"
	LBRACE     = "{"
	RBRACE     = "}"
	FUNCTION   = "FUNCTION"
	LET        = "LET"
	LT         = "<"
	GT         = ">"
	TRUE       = "TRUE"
	FALSE      = "FALSE"
	RETURN     = "RETURN"
	IF         = "IF"
	ELSE       = "ELSE"
	EQ         = "=="
	NOT_EQ     = "!="
	LT_EQ      = "<="
	GT_EQ      = ">="
	AND        = "&&"
	OR         = "||"
	PLUSPLUS   = "++"
	MINUSMINUS = "--"
	STRING     = "STRING"
	RBRACKET   = "]"
	LBRACKET   = "["
)

var keywords = map[string]TokenType{
	"let":    LET,
	"fn":     FUNCTION,
	"true":   TRUE,
	"false":  FALSE,
	"return": RETURN,
	"if":     IF,
	"else":   ELSE,
}

func LookupIdent(s string) TokenType {
	if tok, ok := keywords[s]; ok {
		return tok
	}
	return IDENT
}
