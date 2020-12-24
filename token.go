package deslang

type (
	tokentype int

	Token struct {
		Type    tokentype
		Lexeme  []byte
		Literal []byte
		Line    int
	}
)

const (
	_ tokentype = iota

	// Single-character tokens.
	_left_paren  // 1
	_right_paren // 2
	_left_brace  // 3
	_right_brace // 4
	_comma       // 5
	_minus       // 6
	_plus        // 7
	_semicolon   // 8
	_slash       // 9
	_star        // 10

	// One or two character tokens.
	_bang          // 11
	_bang_equal    // 12
	_equal         // 13
	_equal_equal   // 14
	_greater       // 15
	_greater_equal // 16
	_less          // 17
	_less_equal    // 18

	// Literals.
	_identifier // 19
	_string     // 20
	_number     // 21

	// Keywords.
	_and    // 22
	_else   // 23
	_false  // 24
	_fun    // 25
	_for    // 26
	_if     // 27
	_nil    // 28
	_or     // 29
	_print  // 30
	_return // 31
	_true   // 32
	_var    // 33
	_while  // 34
	_eof    // 35
)
