package promqllint

import "fmt"

// Expr is a parsed promql expression.
type Expr struct {
	// Raw is the node that was parsed to this expression.
	Raw string
	// Type is the type of this expression, e.g. "vector".
	Type string
}

// ParseErr wraps a parsing error with line and position context.
type ParseErr struct {
	Line, Pos int
	Err       error
}

func (e *ParseErr) Error() string {
	if e.Line == 0 {
		return fmt.Sprintf("parse error at char %d: %s", e.Pos, e.Err)
	}
	return fmt.Sprintf("parse error at line %d, char %d: %s", e.Line, e.Pos, e.Err)
}
