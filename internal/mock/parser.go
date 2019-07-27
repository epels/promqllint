package mock

import (
	"github.com/epels/promqllint"
)

type Parser struct {
	ParseExprFunc func(string) (*promqllint.Expr, error)
}

func (p *Parser) ParseExpr(s string) (*promqllint.Expr, error) {
	return p.ParseExprFunc(s)
}
