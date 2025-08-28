package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/frikkfelix/sshchat/go/pkg/server"
)

func main() {
	s, err := server.NewServer()

	if err != nil {
		log.Fatalf("failed to create SSH server: %v", err)
	}

	log.Printf("Starting SSH server on %s", s.Addr)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatal("Could not start server:", err)
		}
	}()

	server.WaitForShutdown()
	log.Println("Stopping SSH server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Fatal("Could not stop server:", err)
	}
}
