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
              out.stringValue = string(lex.data[lex.ts:lex.te]);
              fbreak;
            };
            '"'[^"]*'"' => {
              tok = STRING;
              out.stringValue = string(lex.data[lex.ts+1:lex.te-1]);
              fbreak;
            };
            'id='[a-zA-Z0-9=]+ => {
              tok = ID;
              // todo remove id=
              out.stringValue = string(lex.data[lex.ts+3:lex.te]);
              fbreak;
            };
            'pos=' => {
              tok = TOK_POS;
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
            'Emote'[a-zA-Z]+ => {
              tok = EMOTE;
              out.stringValue = string(lex.data[lex.ts+5:lex.te]);
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
            # Actions
            'connecting' => {
              tok = TOK_CONNECTING;
              fbreak;
            };
            'connected' => {
              tok = TOK_CONNECTED;
              fbreak;
            };
            'performed' => {
              tok = TOK_PERFORMED;
              fbreak;
            };
            # Char literals
            [|<>,()] => {
              tok = int(lex.data[lex.ts])
              fbreak;
            };
            space;
        *|;

         write exec;
    }%%

    return tok;
}
