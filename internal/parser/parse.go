package parser

var log []LogLine

//go:generate go tool goyacc -o adm.go adm.y
func Parse(input []byte, debug bool) ([]LogLine, error) {
	if debug {
		yyDebug = 5
		yyErrorVerbose = true
	}
	lex := newLexer(input)
	if yyParse(lex) != 0 {
		return nil, lex.err
	}
	return log, nil
}
