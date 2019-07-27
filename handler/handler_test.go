package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/epels/promqllint"
	"github.com/epels/promqllint/internal/mock"
)

func TestNew(t *testing.T) {
	var p mock.Parser
	var r mock.Renderer

	h, err := New(&p, &r)
	if err != nil {
		t.Fatalf("handler: New: %s", err)
	}

	if h.pqlParser == nil {
		t.Errorf("New did not set pqlParser")
	}
	if h.lintRenderer == nil {
		t.Errorf("New did not set lintRenderer")
	}
}

func TestServeHTTP(t *testing.T) {
	t.Run("404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/foo", nil)

		var h handler
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Got %d, expected 404", rec.Code)
		}
	})

	t.Run("405", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/", nil)

		var h handler
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Got %d, expected 405", rec.Code)
		}
	})

	t.Run("Show lint", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		var called, ok bool
		h := handler{
			lintRenderer: &mock.Renderer{
				ExecuteFunc: func(w io.Writer, data interface{}) error {
					called = true
					_, ok = data.(*lintResponse)
					return nil
				},
			},
		}
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Got %d, expected 200", rec.Code)
		}

		if !called {
			t.Errorf("Expected call to renderer")
		}
		if ok {
			t.Errorf("Got *lintResponse, expected nil")
		}
	})

	t.Run("Create lint", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			h := handler{lintRenderer: mock.NopRenderer}
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Got %d, expected 200", rec.Code)
			}
		})

		t.Run("Parse error", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("body=sum(1.23)"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			var res *lintResponse
			h := handler{
				pqlParser: &mock.Parser{
					ParseExprFunc: func(s string) (*promqllint.Expr, error) {
						return nil, &promqllint.ParseErr{
							Line: 0,
							Pos:  10,
							Err:  errors.New("some-error"),
						}
					},
				},
				lintRenderer: &mock.Renderer{
					ExecuteFunc: func(w io.Writer, data interface{}) error {
						res = data.(*lintResponse)
						return nil
					},
				},
			}
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Got %d, expected 200", rec.Code)
			}

			if res.Valid {
				t.Errorf("Got true, expected false")
			}
			if res.Raw != "sum(1.23)" {
				t.Errorf("Got %q, expected sum(1.23)", res.Raw)
			}
			if res.Type != "" {
				t.Errorf("Got %q, expected empty string", res.Type)
			}
			if res.ErrorPos != 10 {
				t.Errorf("Got %d, expected 10", res.ErrorPos)
			}
			if res.ErrorLine != 0 {
				t.Errorf("Got %d, expected 0", res.ErrorLine)
			}
			if res.ErrorText != "parse error at char 10: some-error" {
				t.Errorf("Got %q, expected parse error at char 10: some-error", res.ErrorText)
			}
		})

		t.Run("Internal error", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("body=sum(1.23)"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			h := handler{
				pqlParser: &mock.Parser{
					ParseExprFunc: func(s string) (*promqllint.Expr, error) {
						return nil, errors.New("handler must 500: this is not a ParseErr")
					},
				},
				lintRenderer: mock.NopRenderer,
			}
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("Got %d, expected 500", rec.Code)
			}
		})

		t.Run("OK", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("body=1.23"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			var res *lintResponse
			h := handler{
				pqlParser: &mock.Parser{
					ParseExprFunc: func(s string) (*promqllint.Expr, error) {
						return &promqllint.Expr{
							Raw:  "1.23",
							Type: "scalar",
						}, nil
					},
				},
				lintRenderer: &mock.Renderer{
					ExecuteFunc: func(w io.Writer, data interface{}) error {
						res = data.(*lintResponse)
						return nil
					},
				},
			}
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Got %d, expected 200", rec.Code)
			}

			if !res.Valid {
				t.Errorf("Got false, expected true")
			}
			if res.Raw != "1.23" {
				t.Errorf("Got %q, expected 1.23", res.Raw)
			}
			if res.Type != "scalar" {
				t.Errorf("Got %q, expected scalar", res.Type)
			}
			if res.ErrorPos != 0 {
				t.Errorf("Got %d, expected 0", res.ErrorPos)
			}
			if res.ErrorLine != 0 {
				t.Errorf("Got %d, expected 0", res.ErrorLine)
			}
			if res.ErrorText != "" {
				t.Errorf("Got %q, expected empty string", res.ErrorText)
			}
		})
	})
}
