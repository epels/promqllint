package mock

import (
	"io"
)

type Renderer struct {
	ExecuteFunc func(io.Writer, interface{}) error
}

// NopRenderer is a no-op renderer that does nothing and returns a nil error.
var NopRenderer = &Renderer{
	ExecuteFunc: func(_ io.Writer, _ interface{}) error { return nil },
}

func (r *Renderer) Execute(w io.Writer, data interface{}) error {
	return r.ExecuteFunc(w, data)
}
