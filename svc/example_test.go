package svc_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/svc"
)

func Example() {
	env := svc.InitAll()
	defer env.Close()

	r := httpx.NewRouter()
	r.Handle("/hello", httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintln(w, "Hello world!")
		return nil
	}))

	h := svc.NewStandardHandler(svc.HandlerOpts{
		Router:         r,
		Reporter:       env.Reporter,
		HandlerTimeout: 15 * time.Second,
	})

	s := svc.NewServer(h, svc.WithPort("8080"))

	// To illustrate shutting down a background process when server shuts down.
	bg := NewBGProc()
	bg.Start()

	svc.RunServer(s, bg.Stop)
}

type BackgroundProcess struct {
	shutdown chan struct{}
	done     chan struct{}
}

func NewBGProc() *BackgroundProcess {
	return &BackgroundProcess{
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (p *BackgroundProcess) Start() {
	go p.start()
}

func (p *BackgroundProcess) start() {
	defer close(p.done)

	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-t.C:
			fmt.Println("tick")
		case <-p.shutdown:
			return
		}
	}
}

func (p *BackgroundProcess) Stop() {
	close(p.shutdown)
	<-p.done // Wait for p to finish.
}
