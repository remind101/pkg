package svc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// NewServerOpt allows users to customize the http.Server used by RunServer.
type NewServerOpt func(*http.Server)

// ServerDefaults specifies default server options to use for RunServer.
var ServerDefaults = func(srv *http.Server) {
	srv.Addr = ":8080"
	srv.WriteTimeout = 5 * time.Second
	srv.ReadHeaderTimeout = 5 * time.Second
	srv.IdleTimeout = 120 * time.Second
}

// WithPort sets the port for the server to run on.
func WithPort(port string) NewServerOpt {
	return func(srv *http.Server) {
		srv.Addr = ":" + port
	}
}

// NewServer offers some convenience and good defaults for creating an http.Server
func NewServer(h http.Handler, opts ...NewServerOpt) *http.Server {
	srv := &http.Server{Handler: h}

	// Prepend defaults to server options.
	opts = append([]NewServerOpt{ServerDefaults}, opts...)
	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

// RunServer handles the biolerplate of starting an http server and handling
// signals gracefully.
func RunServer(srv *http.Server) {
	idleConnsClosed := make(chan struct{})

	go func() {
		// Handle SIGINT and SIGTERM.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigCh
		fmt.Println("Received signal, stopping.", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			fmt.Printf("HTTP server Shutdown: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	fmt.Printf("HTTP server listening on address: \"%s\"\n", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		fmt.Printf("HTTP server ListenAndServe: %v\n", err)
		os.Exit(1)
	}

	<-idleConnsClosed
}
