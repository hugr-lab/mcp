package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/hugr-lab/mcp/pkg/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	c := config()
	log.Println("MCP Service configured to", c.URL)

	s := service.New(c.Config)

	err := s.Init(ctx)
	if err != nil {
		log.Println("Initialization error:", err)
		os.Exit(1)
	}

	log.Println("Initialization complete")

	srv := &http.Server{
		Addr:    c.Bind,
		Handler: s,
	}

	go func() {
		log.Println("Starting server on http://localhost", c.Bind)
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Println("Server closed")
			return
		}
		if err != nil {
			log.Println("Server error:", err)
		}
	}()

	<-ctx.Done()
}
