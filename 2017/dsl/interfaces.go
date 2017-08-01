package dsl

// Lex takes an input string and produces a series of tokens as output.
func Lex(input string) <-chan Token

// Parse consumes a series of tokens and produces an AST or an error.
func Parse(tokens <-chan Token) (AST, error)

// Eval walks the AST performing the perscribed actions.
func Eval(AST) error
