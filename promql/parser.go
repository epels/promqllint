package promql

import (
	"fmt"

	"github.com/prometheus/prometheus/promql"

	"github.com/epels/promqllint"
)

// Parser parses raw promql expressions.
type Parser struct{}

// ParseExpr parses a raw promql expression. When the expression is invalid,
// the returned error will be a promqllint.ParseErr.
func (p *Parser) ParseExpr(s string) (*promqllint.Expr, error) {
	expr, err := promql.ParseExpr(s)
	if err != nil {
		pErr, ok := err.(*promql.ParseErr)
		if ok {
			return nil, &promqllint.ParseErr{
				Line: pErr.Line,
				Pos:  pErr.Pos,
				Err:  pErr.Err,
			}
		}
		return nil, fmt.Errorf("promql: ParseExpr: %s", err)
	}
	return &promqllint.Expr{
		Raw:  s,
		Type: p.typeString(expr.Type()),
	}, nil
}

// typeString gets a string representation of vt. It override some values
// because promql parser returns types that, in some cases, are known by
// different names by the typical user. E.g. a matrix is documented as a range
// vector.
func (p *Parser) typeString(vt promql.ValueType) string {
	switch s := string(vt); s {
	case "matrix":
		return "rangevector"
	case "vector":
		return "instantvector"
	default:
		return s
	}
}
