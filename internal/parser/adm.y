%{
package parser

import "fmt"
%}

%token STRING FLOAT
%token TIMESTAMP ID
%token TOK_PIPE TOK_LEFT_ARROW TOK_RIGHT_ARROW TOK_EQUALS TOK_COMMA
%token TOK_PLAYER TOK_IS TOK_POS
%token TOK_CONNECTING TOK_CONNECTED

%union {
  lines []LogLine
  line LogLine

  stringValue string
  floatValue float64

  timestamp string
  playerID string
}


%token <stringValue> STRING
%token <floatValue> FLOAT

%token <timestamp> TIMESTAMP
%token <playerID> ID

%type <line> logLine
%type <lines> logLines

%%

log
  : logLines {
    log = $1
  }
  ;

logLines
  : logLines logLine {
    $$ = append($1, $2)
  }
  | logLine {
    $$ = []LogLine{$1}
  }
  ;

logLine
  : TIMESTAMP TOK_PIPE player TOK_IS TOK_CONNECTING {
    $$ = LogLine{
      Timestamp: $1,
    }
  }
  ;

player
  : TOK_PLAYER STRING ID {
    fmt.Printf("!! player: %s\n", $2)
    fmt.Printf("!!     id: %s\n", $3)
  }
  | TOK_PLAYER STRING ID position {
    fmt.Printf("!! player: %s\n", $2)
    fmt.Printf("!!     id: %s\n", $3)
  }
  ;

position
  : TOK_POS TOK_EQUALS TOK_LEFT_ARROW FLOAT TOK_COMMA FLOAT TOK_COMMA FLOAT TOK_COMMA {
    fmt.Printf("!!     position: <%f, %f, %f>\n", $4, $6, $8)
  }
  ;

%%
// 15:34:20 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I= pos=<17491.9, 6849.4, 16.1>) is connected


