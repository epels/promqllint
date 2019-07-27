package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/epels/promqllint"
)

type handler struct {
	pqlParser    parser
	lintRenderer renderer
}

type parser interface {
	ParseExpr(s string) (*promqllint.Expr, error)
}

type renderer interface {
	Execute(io.Writer, interface{}) error
}

func New(p parser, r renderer) (*handler, error) {
	return &handler{
		pqlParser:    p,
		lintRenderer: r,
	}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "http" {
		r.URL.Host = r.Host
		r.URL.Scheme = "https"
		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
		return
	} else if proto == "https" {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")
	}

	if r.URL.Path == "/favicon.ico" {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "./static/assets/favicon.ico")
			return
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path == "/assets/codemirror.css" {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "./static/assets/codemirror.css")
			return
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path == "/assets/codemirror.js" {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "./static/assets/codemirror.js")
			return
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path == "/" {
		if r.Method == http.MethodGet {
			h.showLint(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.createLint(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, r)
}

func (h *handler) showLint(w http.ResponseWriter, r *http.Request) {
	if err := h.lintRenderer.Execute(w, nil); err != nil {
		log.Printf("%T: Execute: %s", h.lintRenderer, err)
	}
}

type lintResponse struct {
	Valid bool
	Raw   string

	Type string // only populated if Valid is true.

	ErrorText string // only populated if Valid is false.
	ErrorPos  int    // only populated if Valid is false.
	ErrorLine int    // only populated if Valid is false.
}

func (h *handler) createLint(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	body := r.PostForm.Get("body")
	if body == "" {
		res := lintResponse{
			Valid:     false,
			Raw:       body,
			ErrorText: "parse error at char 1: no expression found in input",
			ErrorLine: 0,
			ErrorPos:  0,
		}
		if err := h.lintRenderer.Execute(w, &res); err != nil {
			log.Printf("%T: Execute: %s", h.lintRenderer, err)
		}
		return
	}

	expr, err := h.pqlParser.ParseExpr(body)
	if err != nil {
		if pErr, ok := err.(*promqllint.ParseErr); ok {
			res := lintResponse{
				Valid:     false,
				Raw:       body,
				ErrorText: pErr.Error(),
				ErrorLine: pErr.Line,
				ErrorPos:  pErr.Pos,
			}
			if err := h.lintRenderer.Execute(w, &res); err != nil {
				log.Printf("%T: Execute: %s", h.lintRenderer, err)
			}
			return
		}

		log.Printf("%T: ParseExpr: %s", h.pqlParser, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res := lintResponse{
		Valid: true,
		Raw:   body,
		Type:  expr.Type,
	}
	if err := h.lintRenderer.Execute(w, &res); err != nil {
		log.Printf("%T: Execute: %s", h.lintRenderer, err)
	}
}
