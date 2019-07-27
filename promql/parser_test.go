package promql

import (
	"testing"

	"github.com/epels/promqllint"
)

func TestParseExpr(t *testing.T) {
	t.Run("Parse error", func(t *testing.T) {
		var p Parser
		_, err := p.ParseExpr(`http_requests_total{method!="GET"}[1]`)
		pErr, ok := err.(*promqllint.ParseErr)
		if !ok {
			t.Fatalf("Expected error to be *promqllint.ParseErr")
		}
		if pErr.Line != 0 {
			t.Errorf("Got %d, expected 0", pErr.Line)
		}
		if pErr.Pos != 36 {
			t.Errorf("Got %d, expected 36", pErr.Pos)
		}
		if pErr.Err == nil {
			t.Errorf("Got nil, expected inner error")
		}
	})

	t.Run("Ok", func(t *testing.T) {
		tt := []struct {
			name    string
			inRaw   string
			outExpr promqllint.Expr
		}{
			{
				name:  "Instant vector",
				inRaw: `http_requests_total{method!="GET"}`,
				outExpr: promqllint.Expr{
					Raw:  `http_requests_total{method!="GET"}`,
					Type: "instantvector",
				},
			},
			{
				name:  "Instant vector (complex)",
				inRaw: `sum(rate(http_requests_total{method!="GET"}[1h]))`,
				outExpr: promqllint.Expr{
					Raw:  `sum(rate(http_requests_total{method!="GET"}[1h]))`,
					Type: "instantvector",
				},
			},
			{
				name:  "Range vector",
				inRaw: `http_requests_total{method!="GET"}[1m]`,
				outExpr: promqllint.Expr{
					Raw:  `http_requests_total{method!="GET"}[1m]`,
					Type: "rangevector",
				},
			},
			{
				name:  "Scalar",
				inRaw: `1.23`,
				outExpr: promqllint.Expr{
					Raw:  `1.23`,
					Type: "scalar",
				},
			},
			{
				name:  "String",
				inRaw: `"foo"`,
				outExpr: promqllint.Expr{
					Raw:  `"foo"`,
					Type: "string",
				},
			},
		}
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				var p Parser
				expr, err := p.ParseExpr(tc.inRaw)
				if err != nil {
					t.Fatalf("Unexpected error parsing: %s", err)
				}
				if expr.Raw != tc.outExpr.Raw {
					t.Errorf("Got %q, expected %q", expr.Raw, tc.outExpr.Raw)
				}
				if expr.Type != tc.outExpr.Type {
					t.Errorf("Got %q, expected %q", expr.Type, tc.outExpr.Type)
				}
			})
		}
	})
}
