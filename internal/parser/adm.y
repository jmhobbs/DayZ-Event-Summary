%{
package parser

import "fmt"
%}

%token STRING FLOAT INTEGER IDENTIFIER
%token TIMESTAMP ID EMOTE
%token TOK_PLAYER TOK_ID TOK_HP TOK_POS TOK_IS INTO TOK_FOR TOK_DAMAGE
%token TOK_CONNECTING TOK_CONNECTED TOK_PERFORMED TOK_HITBY

%union {
  line LogLine

  stringValue string
  floatValue float64
  intValue int64

  player Player
  position Position
}

%token <stringValue> STRING ID TIMESTAMP EMOTE INTO IDENTIFIER
%token <floatValue> FLOAT
%token <intValue> INTEGER

%type <intValue> hp
%type <player> player
%type <position> position

%type <line> connectingLine connectedLine emoteLine hitByPlayerLine

%%

logLine
  : connectingLine {
    line = $1
  }
  | connectedLine {
    line = $1
  }
  | emoteLine {
    line = $1
  }
  | hitByPlayerLine {
    line = $1
  }
  ;

connectingLine
  : TIMESTAMP '|' player TOK_IS TOK_CONNECTING {
    $$ = LogLine{
      Timestamp: $1,
      Type: "CONNECTING",
      Subject: &$3,
    }
  }
  ;

connectedLine
  : TIMESTAMP '|' player TOK_IS TOK_CONNECTED {
    $$ = LogLine{
      Timestamp: $1,
      Type: "CONNECTED",
      Subject: &$3,
    }
  }
  ;

emoteLine
  : TIMESTAMP '|' player TOK_PERFORMED EMOTE {
    $$ = LogLine{
      Timestamp: $1,
      Type: "EMOTE",
      Subject: &$3,
    }
  }
  ;

hitByPlayerLine
  : TIMESTAMP '|' player TOK_HITBY player INTO '(' INTEGER ')' TOK_FOR INTEGER TOK_DAMAGE '(' IDENTIFIER ')' {
    fmt.Printf("!! hit by player: %s %v\n", $3.Name, $3.Position)
    fmt.Printf("!!     hit by: %s %v\n", $5.Name, $5.Position)
    fmt.Printf("!!     into: %s\n", $6)
    fmt.Printf("!!     damage: %d\n", $11)
    fmt.Printf("!! with: %s\n", $14)
    $$ = LogLine{
      Timestamp: $1,
      Type: "HIT",
      Subject: &$3,
      Object: &$5,
    }
  }
  ;

player
  : TOK_PLAYER STRING '(' ID ')' {
    fmt.Printf("!! player: %s\n", $2)
    fmt.Printf("!!     id: %s\n", $4)
    $$ = Player{
      Name: $2,
      ID: $4,
    }
  }
  | TOK_PLAYER STRING '(' ID position ')' {
    fmt.Printf("!! player: %s\n", $2)
    fmt.Printf("!!     id: %s\n", $4)
    fmt.Printf("!!     pos: <%f, %f, %f>\n", $5.X, $5.Y, $5.Z)
    $$ = Player{
      Name: $2,
      ID: $4,
      Position: &Position{
        X: $5.X,
        Y: $5.Y,
        Z: $5.Z,
      },
    }
  }
  | TOK_PLAYER STRING '(' ID position ')' hp {
    fmt.Printf("!! player: %s\n", $2)
    fmt.Printf("!!     id: %s\n", $4)
    fmt.Printf("!!     pos: <%f, %f, %f>\n", $5.X, $5.Y, $5.Z)
    fmt.Printf("!!     hp: %d\n", $7)
    $$ = Player{
      Name: $2,
      ID: $4,
      Position: &Position{
        X: $5.X,
        Y: $5.Y,
        Z: $5.Z,
      },
    }
  }
  ;

position
  : TOK_POS '<' FLOAT ',' FLOAT ',' FLOAT '>' {
    fmt.Printf("!!     position: <%f, %f, %f>\n", $3, $5, $7)
    $$ = Position{
      X: $3,
      Y: $5,
      Z: $7,
    }
  }
  ;

hp
  : '[' TOK_HP INTEGER ']' {
    fmt.Printf("!!     hp: %d\n", $3)
    $$ = $3
  }
  ;

%%
// 15:34:20 | Player "jmhobbs" (id=BFDIjY8X3a21fFxgXW339fE0IsxOK0kI7xlfig3HN1I= pos=<17491.9, 6849.4, 16.1>) is connected

