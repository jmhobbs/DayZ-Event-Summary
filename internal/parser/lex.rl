package parser

import "strconv"

%%{
  machine lex;
  write data;
  access lex.;
  variable p lex.p;
  variable pe lex.pe;
}%%

func newLexer(data []byte) *lexer {
    lex := &lexer{ 
        data: data,
        pe: len(data),
    }
    %% write init;
    return lex
}

func (lex *lexer) Lex(out *yySymType) int {
    eof := lex.pe
    tok := 0
    %%{ 
        main := |*
            [0-9][0-9]':'[0-9][0-9]':'[0-9][0-9] => {
              tok = TIMESTAMP;
              out.timestamp = string(lex.data[lex.ts:lex.te]);
              fbreak;
            };
            '"'[^"]*'"' => {
              tok = STRING;
              out.stringValue = string(lex.data[lex.ts:lex.te]);
              fbreak;
            };
            '(id='[^"]*')' => {
              tok = ID;
              out.playerID = string(lex.data[lex.ts:lex.te]);
              fbreak;
            };
            [0-9]+'.'[0-9]+ => {
              n, err := strconv.ParseFloat(string(lex.data[lex.ts:lex.te]), 64);
              if err != nil {
                panic(err)
              }
              out.floatValue = n;
              tok = FLOAT;
              fbreak;
            };
            'Player' => {
              tok = TOK_PLAYER;
              fbreak;
            };
            'is' => {
              tok = TOK_IS;
              fbreak;
            };
            'connecting' => {
              tok = TOK_CONNECTING;
              fbreak;
            };
            'connected' => {
              tok = TOK_CONNECTED;
              fbreak;
            };
            'pos' => {
              tok = TOK_POS;
              fbreak;
            };
            '|' => {
              tok = TOK_PIPE;
              fbreak;
            };
            '<' => {
              tok = TOK_LEFT_ARROW;
              fbreak;
            };
            '>' => {
              tok = TOK_RIGHT_ARROW;
              fbreak;
            };
            ',' => {
              tok = TOK_COMMA;
              fbreak;
            };
            '=' => {
              tok = TOK_EQUALS;
              fbreak;
            };
            space;
        *|;

         write exec;
    }%%

    return tok;
}

