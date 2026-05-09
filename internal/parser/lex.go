
//line lex.rl:1
package parser

import "strconv"


//line lex.go:9
const lex_start int = 33
const lex_first_final int = 33
const lex_error int = 0

const lex_en_main int = 33


//line lex.rl:11


func newLexer(data []byte) *lexer {
    lex := &lexer{ 
        data: data,
        pe: len(data),
    }
    
//line lex.go:26
	{
	 lex.cs = lex_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line lex.rl:19
    return lex
}

func (lex *lexer) Lex(out *yySymType) int {
    eof := lex.pe
    tok := 0
    
//line lex.go:42
	{
	if ( lex.p) == ( lex.pe) {
		goto _test_eof
	}
	switch  lex.cs {
	case 33:
		goto st_case_33
	case 0:
		goto st_case_0
	case 1:
		goto st_case_1
	case 2:
		goto st_case_2
	case 3:
		goto st_case_3
	case 4:
		goto st_case_4
	case 5:
		goto st_case_5
	case 34:
		goto st_case_34
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 35:
		goto st_case_35
	case 8:
		goto st_case_8
	case 9:
		goto st_case_9
	case 10:
		goto st_case_10
	case 11:
		goto st_case_11
	case 12:
		goto st_case_12
	case 13:
		goto st_case_13
	case 14:
		goto st_case_14
	case 15:
		goto st_case_15
	case 16:
		goto st_case_16
	case 17:
		goto st_case_17
	case 18:
		goto st_case_18
	case 19:
		goto st_case_19
	case 20:
		goto st_case_20
	case 21:
		goto st_case_21
	case 22:
		goto st_case_22
	case 23:
		goto st_case_23
	case 24:
		goto st_case_24
	case 25:
		goto st_case_25
	case 26:
		goto st_case_26
	case 27:
		goto st_case_27
	case 28:
		goto st_case_28
	case 29:
		goto st_case_29
	case 30:
		goto st_case_30
	case 31:
		goto st_case_31
	case 32:
		goto st_case_32
	}
	goto st_out
tr1:
//line lex.rl:32
 lex.te = ( lex.p)+1
{
              tok = STRING;
              out.stringValue = string(lex.data[lex.ts:lex.te]);
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr6:
//line NONE:1
	switch  lex.act {
	case 0:
	{{goto st0 }}
	case 3:
	{( lex.p) = ( lex.te) - 1

              tok = ID;
              out.playerID = string(lex.data[lex.ts:lex.te]);
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	}
	
	goto st33
tr17:
//line lex.rl:27
 lex.te = ( lex.p)+1
{
              tok = TIMESTAMP;
              out.timestamp = string(lex.data[lex.ts:lex.te]);
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr22:
//line lex.rl:51
 lex.te = ( lex.p)+1
{
              tok = TOK_PLAYER;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr31:
//line lex.rl:63
 lex.te = ( lex.p)+1
{
              tok = TOK_CONNECTED;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr33:
//line lex.rl:59
 lex.te = ( lex.p)+1
{
              tok = TOK_CONNECTING;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr34:
//line lex.rl:55
 lex.te = ( lex.p)+1
{
              tok = TOK_IS;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr36:
//line lex.rl:67
 lex.te = ( lex.p)+1
{
              tok = TOK_POS;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr37:
//line lex.rl:91
 lex.te = ( lex.p)+1

	goto st33
tr39:
//line lex.rl:83
 lex.te = ( lex.p)+1
{
              tok = TOK_COMMA;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr41:
//line lex.rl:75
 lex.te = ( lex.p)+1
{
              tok = TOK_LEFT_ARROW;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr42:
//line lex.rl:87
 lex.te = ( lex.p)+1
{
              tok = TOK_EQUALS;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr43:
//line lex.rl:79
 lex.te = ( lex.p)+1
{
              tok = TOK_RIGHT_ARROW;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr48:
//line lex.rl:71
 lex.te = ( lex.p)+1
{
              tok = TOK_PIPE;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr49:
//line lex.rl:37
 lex.te = ( lex.p)
( lex.p)--
{
              tok = ID;
              out.playerID = string(lex.data[lex.ts:lex.te]);
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
tr50:
//line lex.rl:42
 lex.te = ( lex.p)
( lex.p)--
{
              n, err := strconv.ParseFloat(string(lex.data[lex.ts:lex.te]), 64);
              if err != nil {
                panic(err)
              }
              out.floatValue = n;
              tok = FLOAT;
              {( lex.p)++;  lex.cs = 33; goto _out }
            }
	goto st33
	st33:
//line NONE:1
 lex.ts = 0

//line NONE:1
 lex.act = 0

		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof33
		}
	st_case_33:
//line NONE:1
 lex.ts = ( lex.p)

//line lex.go:278
		switch  lex.data[( lex.p)] {
		case 32:
			goto tr37
		case 34:
			goto st1
		case 40:
			goto st2
		case 44:
			goto tr39
		case 60:
			goto tr41
		case 61:
			goto tr42
		case 62:
			goto tr43
		case 80:
			goto st15
		case 99:
			goto st20
		case 105:
			goto st30
		case 112:
			goto st31
		case 124:
			goto tr48
		}
		switch {
		case  lex.data[( lex.p)] > 13:
			if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
				goto st6
			}
		case  lex.data[( lex.p)] >= 9:
			goto tr37
		}
		goto st0
st_case_0:
	st0:
		 lex.cs = 0
		goto _out
	st1:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof1
		}
	st_case_1:
		if  lex.data[( lex.p)] == 34 {
			goto tr1
		}
		goto st1
	st2:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof2
		}
	st_case_2:
		if  lex.data[( lex.p)] == 105 {
			goto st3
		}
		goto st0
	st3:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof3
		}
	st_case_3:
		if  lex.data[( lex.p)] == 100 {
			goto st4
		}
		goto st0
	st4:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof4
		}
	st_case_4:
		if  lex.data[( lex.p)] == 61 {
			goto st5
		}
		goto st0
	st5:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof5
		}
	st_case_5:
		switch  lex.data[( lex.p)] {
		case 34:
			goto tr6
		case 41:
			goto tr7
		}
		goto st5
tr7:
//line NONE:1
 lex.te = ( lex.p)+1

//line lex.rl:37
 lex.act = 3;
	goto st34
	st34:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof34
		}
	st_case_34:
//line lex.go:378
		switch  lex.data[( lex.p)] {
		case 34:
			goto tr49
		case 41:
			goto tr7
		}
		goto st5
	st6:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof6
		}
	st_case_6:
		if  lex.data[( lex.p)] == 46 {
			goto st7
		}
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st8
		}
		goto st0
	st7:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof7
		}
	st_case_7:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st35
		}
		goto st0
	st35:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof35
		}
	st_case_35:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st35
		}
		goto tr50
	st8:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof8
		}
	st_case_8:
		switch  lex.data[( lex.p)] {
		case 46:
			goto st7
		case 58:
			goto st10
		}
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st9
		}
		goto st0
	st9:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof9
		}
	st_case_9:
		if  lex.data[( lex.p)] == 46 {
			goto st7
		}
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st9
		}
		goto st0
	st10:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof10
		}
	st_case_10:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st11
		}
		goto st0
	st11:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof11
		}
	st_case_11:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st12
		}
		goto st0
	st12:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof12
		}
	st_case_12:
		if  lex.data[( lex.p)] == 58 {
			goto st13
		}
		goto st0
	st13:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof13
		}
	st_case_13:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st14
		}
		goto st0
	st14:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof14
		}
	st_case_14:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto tr17
		}
		goto st0
	st15:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof15
		}
	st_case_15:
		if  lex.data[( lex.p)] == 108 {
			goto st16
		}
		goto st0
	st16:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof16
		}
	st_case_16:
		if  lex.data[( lex.p)] == 97 {
			goto st17
		}
		goto st0
	st17:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof17
		}
	st_case_17:
		if  lex.data[( lex.p)] == 121 {
			goto st18
		}
		goto st0
	st18:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof18
		}
	st_case_18:
		if  lex.data[( lex.p)] == 101 {
			goto st19
		}
		goto st0
	st19:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof19
		}
	st_case_19:
		if  lex.data[( lex.p)] == 114 {
			goto tr22
		}
		goto st0
	st20:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof20
		}
	st_case_20:
		if  lex.data[( lex.p)] == 111 {
			goto st21
		}
		goto st0
	st21:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof21
		}
	st_case_21:
		if  lex.data[( lex.p)] == 110 {
			goto st22
		}
		goto st0
	st22:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof22
		}
	st_case_22:
		if  lex.data[( lex.p)] == 110 {
			goto st23
		}
		goto st0
	st23:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof23
		}
	st_case_23:
		if  lex.data[( lex.p)] == 101 {
			goto st24
		}
		goto st0
	st24:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof24
		}
	st_case_24:
		if  lex.data[( lex.p)] == 99 {
			goto st25
		}
		goto st0
	st25:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof25
		}
	st_case_25:
		if  lex.data[( lex.p)] == 116 {
			goto st26
		}
		goto st0
	st26:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof26
		}
	st_case_26:
		switch  lex.data[( lex.p)] {
		case 101:
			goto st27
		case 105:
			goto st28
		}
		goto st0
	st27:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof27
		}
	st_case_27:
		if  lex.data[( lex.p)] == 100 {
			goto tr31
		}
		goto st0
	st28:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof28
		}
	st_case_28:
		if  lex.data[( lex.p)] == 110 {
			goto st29
		}
		goto st0
	st29:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof29
		}
	st_case_29:
		if  lex.data[( lex.p)] == 103 {
			goto tr33
		}
		goto st0
	st30:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof30
		}
	st_case_30:
		if  lex.data[( lex.p)] == 115 {
			goto tr34
		}
		goto st0
	st31:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof31
		}
	st_case_31:
		if  lex.data[( lex.p)] == 111 {
			goto st32
		}
		goto st0
	st32:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof32
		}
	st_case_32:
		if  lex.data[( lex.p)] == 115 {
			goto tr36
		}
		goto st0
	st_out:
	_test_eof33:  lex.cs = 33; goto _test_eof
	_test_eof1:  lex.cs = 1; goto _test_eof
	_test_eof2:  lex.cs = 2; goto _test_eof
	_test_eof3:  lex.cs = 3; goto _test_eof
	_test_eof4:  lex.cs = 4; goto _test_eof
	_test_eof5:  lex.cs = 5; goto _test_eof
	_test_eof34:  lex.cs = 34; goto _test_eof
	_test_eof6:  lex.cs = 6; goto _test_eof
	_test_eof7:  lex.cs = 7; goto _test_eof
	_test_eof35:  lex.cs = 35; goto _test_eof
	_test_eof8:  lex.cs = 8; goto _test_eof
	_test_eof9:  lex.cs = 9; goto _test_eof
	_test_eof10:  lex.cs = 10; goto _test_eof
	_test_eof11:  lex.cs = 11; goto _test_eof
	_test_eof12:  lex.cs = 12; goto _test_eof
	_test_eof13:  lex.cs = 13; goto _test_eof
	_test_eof14:  lex.cs = 14; goto _test_eof
	_test_eof15:  lex.cs = 15; goto _test_eof
	_test_eof16:  lex.cs = 16; goto _test_eof
	_test_eof17:  lex.cs = 17; goto _test_eof
	_test_eof18:  lex.cs = 18; goto _test_eof
	_test_eof19:  lex.cs = 19; goto _test_eof
	_test_eof20:  lex.cs = 20; goto _test_eof
	_test_eof21:  lex.cs = 21; goto _test_eof
	_test_eof22:  lex.cs = 22; goto _test_eof
	_test_eof23:  lex.cs = 23; goto _test_eof
	_test_eof24:  lex.cs = 24; goto _test_eof
	_test_eof25:  lex.cs = 25; goto _test_eof
	_test_eof26:  lex.cs = 26; goto _test_eof
	_test_eof27:  lex.cs = 27; goto _test_eof
	_test_eof28:  lex.cs = 28; goto _test_eof
	_test_eof29:  lex.cs = 29; goto _test_eof
	_test_eof30:  lex.cs = 30; goto _test_eof
	_test_eof31:  lex.cs = 31; goto _test_eof
	_test_eof32:  lex.cs = 32; goto _test_eof

	_test_eof: {}
	if ( lex.p) == eof {
		switch  lex.cs {
		case 5:
			goto tr6
		case 34:
			goto tr49
		case 35:
			goto tr50
		}
	}

	_out: {}
	}

//line lex.rl:95


    return tok;
}

