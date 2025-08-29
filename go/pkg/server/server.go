package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/frikkfelix/sshchat/go/pkg/core"
	"github.com/frikkfelix/sshchat/go/pkg/tui"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
	xssh "golang.org/x/crypto/ssh"
)

const (
	host = "0.0.0.0"
	port = "42069"
)

type Server struct {
	hub *core.Hub
	SSH *ssh.Server
}

func New(hub *core.Hub) (*Server, error) {
	teaHandler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		session := core.NewSession(s)
		hub.RegisterSession(session)

		go func() {
			<-s.Context().Done()
			session.Close()
		}()

		model := tui.NewModel(session, hub)

		return model, []tea.ProgramOption{
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		}
	}
	wishServer, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			fingerprint := xssh.FingerprintSHA256(key.(gossh.PublicKey))
			ctx.SetValue("fingerprint", fingerprint)
			return true
		}),
		wish.WithKeyboardInteractiveAuth(
			func(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) bool {
				ctx.SetValue("fingerprint", uuid.NewString())
				return true
			},
		),
	)

	return &Server{
		hub: hub,
		SSH: wishServer,
	}, err
}

func WaitForShutdown() {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownCh
}

func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		_ = s.SSH.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
