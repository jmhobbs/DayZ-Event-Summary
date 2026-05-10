%{
package parser

import "fmt"
%}

%token STRING FLOAT
%token TIMESTAMP ID
%token TOK_PIPE TOK_LEFT_ARROW TOK_RIGHT_ARROW TOK_EQUALS TOK_COMMA TOK_LEFT_PAREN TOK_RIGHT_PAREN
%token TOK_PLAYER TOK_ID TOK_IS TOK_POS
%token TOK_CONNECTING TOK_CONNECTED

%union {
  line LogLine

  stringValue string
  floatValue float64

  player Player
  position Position
}

%token <stringValue> STRING ID TIMESTAMP
%token <floatValue> FLOAT

%type <player> player
%type <position> position

%type <line> connectingLine connectedLine

%%

logLine
  : connectingLine {
    line = $1
  }
  | connectedLine {
    line = $1
  }
  ;

connectingLine
  : TIMESTAMP TOK_PIPE player TOK_IS TOK_CONNECTING {
    $$ = LogLine{
      Timestamp: $1,
      Type: "CONNECTING",
      Player: &$3,
    }
  }
  ;

connectedLine
  : TIMESTAMP TOK_PIPE player TOK_IS TOK_CONNECTED {
    $$ = LogLine{
      Timestamp: $1,
      Type: "CONNECTED",
      Player: &$3,
    }
  }
  ;

player
  : TOK_PLAYER STRING TOK_LEFT_PAREN ID TOK_RIGHT_PAREN {
    fmt.Printf("!! player: %s\n", $2)
    fmt.Printf("!!     id: %s\n", $4)
    $$ = Player{
      Name: $2,
      ID: $4,
    }
  }
  | TOK_PLAYER STRING TOK_LEFT_PAREN ID position TOK_RIGHT_PAREN {
    fmt.Printf("!! player: %s\n", $2)
    fmt.Printf("!!     id: %s\n", $4)
    fmt.Printf("!!     pos: <%f, %f, %f>\n", $5.X, $5.Y, $5.Z)
    $$ = Player{
      Name: $2,
      ID: $4,
      Position: &$5,
    }
  }
  ;

position
  : TOK_POS TOK_LEFT_ARROW FLOAT TOK_COMMA FLOAT TOK_COMMA FLOAT TOK_RIGHT_ARROW {
    fmt.Printf("!!     position: <%f, %f, %f>\n", $3, $5, $7)
    $$ = Position{
      X: $3,
      Y: $5,
      Z: $7,
    }
  }
  ;

%%
// 15:34:20 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I= pos=<17491.9, 6849.4, 16.1>) is connected


