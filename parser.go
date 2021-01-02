package deslang

import (
	"reflect"
)

// Parses tokens into nodes. Errors are sent to the ErrorReporter. Caller should
// check for errors after parsing.
type Parser struct {
	errh    errorHandler // any errors during scanning
	current int          // index of next token to be parsed
	tokens  []Token
}

func NewParser(errh errorHandler) *Parser {
	return &Parser{errh: errh}
}

func (p *Parser) reset() {
	p.tokens = []Token{}
	p.current = 0
}

func (p *Parser) Parse(tokens []Token) []Stmt {
	p.reset()
	p.tokens = tokens
	var stmts []Stmt

	for !p.isAtEnd() {
		stmts = append(stmts, p.decl())
	}

	return stmts
}

func (p *Parser) syntaxError(t Token, msg string) {
	if t.Type == _eof {
		p.errh(t.Line, "at end", msg)
	} else {
		p.errh(t.Line, "at '"+string(t.Lexeme)+"'", msg)
	}
	p.synchronize()
}

// Discard tokens until a statement boundary is found. This is used for error
// production. If an error is found during parsing, the Parser will try to parse
// the remaining code after this synchronization point.
func (p *Parser) synchronize() {
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

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

// Checks if the current token's Type matches any of the given types
func (p *Parser) match(types ...tokentype) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

// Check if the current token Type matches t. Return false if the end of the
// source is reached.
func (p *Parser) check(t tokentype) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == _eof
}

func (p *Parser) consume(tt tokentype, msg string) Token {
	if p.check(tt) {
		return p.advance()
	}

	p.syntaxError(p.peek(), msg)
	return Token{}
}

func (p *Parser) decl() Stmt {
	if p.match(_var) {
		return p.varDecl()
	}
	return p.stmt()
}

func (p *Parser) varDecl() Stmt {
	var expr Expr
	name := p.consume(_identifier, "Expect variable name.")

	if p.match(_equal) {
		expr = p.expression()
	}

	p.consume(_semicolon, "Expect ';' after variable declaration.")
	return VarStmt{Name: name, Expr: expr}
}

func (p *Parser) stmt() Stmt {
	if p.match(_if) {
		return p.ifStmt()
	}

	if p.match(_print) {
		return p.printStmt()
	}

	if p.match(_left_brace) {
		return BlockStmt{Stmts: p.block()}
	}

	return p.exprStmt()
}

func (p *Parser) exprStmt() Stmt {
	expr := p.expression()
	p.consume(_semicolon, "Expect ';' after value.")
	return ExprStmt{Expr: expr}
}

func (p *Parser) ifStmt() Stmt {
	p.consume(_left_paren, "Expect '(' after 'if'.")
	expr := p.expression()
	p.consume(_right_paren, "Expect ')' after if condition.")

	thenBranch := p.stmt()
	var elseBranch Stmt

	if p.match(_else) {
		elseBranch = p.stmt()
	} else {
		elseBranch = NilStmt{}
	}

	return IfStmt{
		Cond: expr,
		Then: thenBranch,
		Else: elseBranch,
	}
}

func (p *Parser) printStmt() Stmt {
	val := p.expression()
	p.consume(_semicolon, "Expect ';' after value.")
	return PrintStmt{Expr: val}
}

func (p *Parser) block() []Stmt {
	var stmts []Stmt

	for !p.check(_right_brace) {
		stmts = append(stmts, p.decl())
	}

	p.consume(_right_brace, "Expect '}' after block.")
	return stmts
}

func (p *Parser) primary() Expr {
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

	if p.match(_identifier) {
		return Variable{Name: p.previous()}
	}

	if p.match(_left_paren) {
		expr := p.expression()
		p.consume(_right_paren, "Expect ')' after expression.")
		return Grouping{X: expr}
	}

	p.errh(p.peek().Line, "", "Expected expression")

	return BasicLit{Value: "", Kind: nilLit}
}

func (p *Parser) unary() Expr {
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

func (p *Parser) factor() Expr {
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

func (p *Parser) term() Expr {
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

func (p *Parser) comparison() Expr {
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

func (p *Parser) expression() Expr {
	return p.assignment()
}

func (p *Parser) equality() Expr {
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

func (p *Parser) and() Expr {
	expr := p.equality()

	for p.match(_and) {
		op := p.previous()
		right := p.equality()

		expr = Logical{
			Left:  expr,
			Right: right,
			Op:    op,
		}
	}

	return expr
}

func (p *Parser) or() Expr {
	expr := p.and()

	for p.match(_or) {
		op := p.previous()
		right := p.and()

		expr = Logical{
			Left:  expr,
			Right: right,
			Op:    op,
		}
	}

	return expr
}

func (p *Parser) assignment() Expr {
	expr := p.or()

	if p.match(_equal) {
		equals := p.previous()
		val := p.assignment()

		// check if left expr is Variable
		f := reflect.ValueOf(expr).FieldByName("Name")
		if f.IsValid() {
			return Assign{
				Name:  f.Interface().(Token),
				Value: val,
			}
		}

		p.errh(equals.Line, "", "Invalid assignment target.")
	}

	return expr
}
