package dsl

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"unicode"
)

// BEGIN stateFn OMIT
type stateFn func(l *lexer) stateFn

// END stateFn OMIT

// BEGIN token OMIT
type TokenType int

type Token struct {
	Pos   Position // Line and Char information
	Type  TokenType
	Value string
}

// END token OMIT

// BEGIN tokenTypes OMIT
const (
	TokenError TokenType = iota
	TokenEOF

	TokenSet
	TokenGet
	TokenVar
	// ...
	TokenWord
	TokenPathSeparator
)

// END tokenTypes OMIT

// BEGIN lex.run OMIT
func (l *lexer) lex() {
	for state := lexToken; state != nil; {
		state = state(l)
	}
}

// END lex.run OMIT

func lexing() {

	// BEGIN choices OMIT
	l.emit(TokenTime)
	// Ignore space between time digits and AM|PM.
	l.ignoreSpace() // HL
	return lexAMPM
	// END choices OMIT
}

func (l *lexer) current() string {
	return l.input[l.start:l.pos]
}
func (l *lexer) emit(t TokenType) {
	value := l.current()
	l.tokens <- Token{
		Pos:   l.position(),
		Type:  t,
		Value: value,
	}
	l.updatePosCounters()
}

func lexNumberOrTimeOrDuration(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == ':':
			return lexTimeDigits
			// Handle cases for number or durations ...
		}
	}
}

func lexTimeDigits(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case unicode.IsDigit(r):
			//absorb
		default:
			l.backup()
			l.emit(TokenTime)
			// Ignore space between time and AM|PM.
			l.ignoreSpace()
			return lexAMPM
		}
	}
}

// BEGIN AST OMIT
type Node interface {
	Pos() Position
}
type SetStatementNode struct {
	Position
	DeviceMatch *PathMatchNode
	Value       *ValueNode
}

// END AST OMIT

func (p *parser) setStatement() *SetStatementNode {
	t := p.expect(TokenSet)
	pm := p.pathMatch()
	v := p.value()
	return &SetStatementNode{
		Position:    t.Pos,
		DeviceMatch: pm,
		Value:       v,
	}
}
func (p *parser) blockStatement() Node {
	switch p.peek().Type {
	case TokenSet:
		return p.setStatement()
	case TokenGet:
		return p.getStatement()
	case TokenVar:
		return p.varStatement()
	case TokenAt:
		return p.atStatement()
	case TokenWhen:
		return p.whenStatement()
	default:
		p.unexpected(p.next(), TokenSet, TokenVar, TokenAt, TokenWhen)
		return nil
	}
}

func (e *Evaluator) eval(node dsl.Node) (Result, error) {
	switch n := node.(type) {
	case *dsl.ProgramNode:
		return e.evalNodeList(n.Statements)
	case *dsl.SetStatementNode:
		return e.evalSet(n)
	case *dsl.GetStatementNode:
		return e.evalGet(n)
	case *dsl.WhenStatementNode:
		return e.evalWhen(n)
	case *dsl.BlockNode:
		return e.evalNodeList(n.Statements)
	default:
		return nil, fmt.Errorf("unknown command %T", node)
	}
}

func main() {
	server, err := smartmqtt.New()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	e := repl.NewEvaluator(server)
	for scanner.Scan() {
		ast, err := dsl.Parse(scanner.Text())
		r, err := e.Eval(ast)
		if err != nil {
			fmt.Println("E", err)
			continue
		}
		if r != nil {
			fmt.Println(r.String())
		}
	}
}
