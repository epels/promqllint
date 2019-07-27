package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/epels/promqllint/handler"
	"github.com/epels/promqllint/promql"
)

func main() {
	// Google App Engine sets the PORT environment variable.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	var parser promql.Parser
	tmpl, err := template.ParseFiles("./static/index.gohtml")
	if err != nil {
		log.Fatalf("html/template: ParseFiles: %s", err)
	}

	h, err := handler.New(&parser, tmpl)
	if err != nil {
		log.Fatalf("handler: New: %s", err)
	}

	s := http.Server{
		Addr:    addr,
		Handler: h,

		IdleTimeout:  60 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %q", addr)
		errCh <- s.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		log.Printf("Exiting with error: %s", err)
	case sig := <-sigCh:
		log.Printf("Exiting with signal: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("net/http: Server.Shutdown: %v", err)
	}
}
