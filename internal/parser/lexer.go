package parser

import "fmt"

//go:generate ragel -Z -G2 -o lex.go lex.rl
type lexer struct {
	data        []byte
	p, pe, cs   int
	ts, te, act int
	err         error
}

func (lex *lexer) Error(e string) {
	lex.err = fmt.Errorf("error: %s", e)
}
