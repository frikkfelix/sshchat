package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/frikkfelix/sshchat/go/pkg/core"
	"github.com/frikkfelix/sshchat/go/pkg/server"
)

func main() {
	hub := core.NewHub()
	go hub.Run()

	srv, err := server.New(hub)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	log.Printf("Starting SSH server on %s", srv.SSH.Addr)

	go func() {
		if err := srv.SSH.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatal("Could not start server:", err)
		}
	}()

	server.WaitForShutdown()
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server shutdown error:", err)
	}

	hub.Shutdown()

}
