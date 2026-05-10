package parser

//go:generate go tool goyacc -o adm.go adm.y

type Parser struct {
	Debug bool
}

var line LogLine

func ParseLine(input []byte, debug bool) (*LogLine, error) {
	if debug {
		yyDebug = 5
		yyErrorVerbose = true
	}
	lex := newLexer(input)
	if yyParse(lex) != 0 {
		return nil, lex.err
	}
	return &line, nil
}
