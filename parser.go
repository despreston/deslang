package deslang

// Parses tokens into nodes. Errors are sent to the ErrorReporter. Caller should
// check for errors after parsing.
type parser struct {
	errh    errorHandler // any errors during scanning
	current int          // index of next token to be parsed
	tokens  []Token
}

func NewParser(tokens []Token, errh errorHandler) *parser {
	return &parser{
		errh:   errh,
		tokens: tokens,
	}
}

func (p *parser) Parse() []Stmt {
	var stmts []Stmt

	for !p.isAtEnd() {
		stmts = append(stmts, p.stmt())
	}

	return stmts
}

func (p *parser) syntaxError(t Token, msg string) {
	if t.Type == _eof {
		p.errh(t.Line, "at end", msg)
	} else {
		p.errh(t.Line, "at '"+string(t.Lexeme)+"'", msg)
	}
	p.synchronize()
}

// Discard tokens until a statement boundary is found. This is used for error
// production. If an error is found during parsing, the parser will try to parse
// the remaining code after this synchronization point.
func (p *parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == _semicolon {
			return
		}

		switch p.peek().Type {
		case _fun, _var, _for, _if, _while, _print, _return:
			return
		}

		p.advance()
	}
}

func (p *parser) peek() Token {
	return p.tokens[p.current]
}

// Checks if the current token's Type matches any of the given types
func (p *parser) match(types ...tokentype) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parser) check(t tokentype) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *parser) isAtEnd() bool {
	return p.peek().Type == _eof
}

func (p *parser) consume(tt tokentype, msg string) Token {
	if p.check(tt) {
		return p.advance()
	}

	p.syntaxError(p.peek(), msg)
	return Token{}
}

func (p *parser) stmt() Stmt {
	if p.match(_print) {
		return p.printStmt()
	}
	return p.exprStmt()
}

func (p *parser) printStmt() Stmt {
	val := p.expression()
	p.consume(_semicolon, "Expect ';' after value.")
	return PrintStmt{Expr: val}
}

func (p *parser) exprStmt() Stmt {
	expr := p.expression()
	p.consume(_semicolon, "Expect ';' after value.")
	return ExprStmt{Expr: expr}
}

func (p *parser) primary() Expr {
	if p.match(_false) {
		return BasicLit{Value: "false", Kind: boolLit}
	}

	if p.match(_true) {
		return BasicLit{Value: "true", Kind: boolLit}
	}

	if p.match(_number) {
		return BasicLit{Value: string(p.previous().Literal), Kind: floatLit}
	}

	if p.match(_string) {
		return BasicLit{Value: string(p.previous().Literal), Kind: stringLit}
	}

	if p.match(_left_paren) {
		expr := p.expression()
		p.consume(_right_paren, "Expect ')' after expression.")
		return Grouping{X: expr}
	}

	p.errh(p.peek().Line, "", "Expected expression")

	return BasicLit{Value: "", Kind: nilLit}
}

func (p *parser) unary() Expr {
	if p.match(_bang, _minus) {
		op := p.previous()
		right := p.unary()

		return Unary{
			Op:    op,
			Right: right,
		}
	}

	return p.primary()
}

func (p *parser) factor() Expr {
	expr := p.unary()

	for p.match(_slash, _star) {
		op := p.previous()
		right := p.unary()

		expr = Binary{
			Left:  expr,
			Right: right,
			Op:    op,
		}
	}

	return expr
}

func (p *parser) term() Expr {
	expr := p.factor()

	for p.match(_minus, _plus) {
		op := p.previous()
		right := p.factor()

		expr = Binary{
			Left:  expr,
			Right: right,
			Op:    op,
		}
	}

	return expr
}

func (p *parser) comparison() Expr {
	expr := p.term()

	for p.match(_greater, _greater_equal, _less, _less_equal) {
		op := p.previous()
		right := p.term()

		expr = Binary{
			Left:  expr,
			Right: right,
			Op:    op,
		}
	}

	return expr
}

func (p *parser) expression() Expr {
	return p.equality()
}

func (p *parser) equality() Expr {
	expr := p.comparison()

	for p.match(_bang_equal, _equal_equal) {
		op := p.previous()
		rightExpr := p.comparison()

		expr = Binary{
			Left:  expr,
			Right: rightExpr,
			Op:    op,
		}
	}

	return expr
}
