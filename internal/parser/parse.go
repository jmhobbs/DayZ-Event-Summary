package parser

import (
	"bufio"
	"io"
)

//go:generate go tool goyacc -o adm.go adm.y

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

func Parse(input io.Reader, debug bool) ([]LogLine, error) {
	logs := []LogLine{}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		logLine, err := ParseLine(scanner.Bytes(), debug)
		if err != nil {
		}
		logs = append(logs, *logLine)
	}

	return logs, scanner.Err()
}
