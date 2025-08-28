package server

import (
	"crypto/md5"
	"encoding/hex"
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
	"github.com/frikkfelix/sshchat/go/pkg/tui"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

const (
	host = "0.0.0.0"
	port = "42069"
)

func NewServer() (*ssh.Server, error) {
	return wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			hash := md5.Sum(key.Marshal())
			fingerprint := hex.EncodeToString(hash[:])
			ctx.SetValue("fingerprint", fingerprint)
			ctx.SetValue("anonymous", false)
			return true
		}),
		wish.WithKeyboardInteractiveAuth(
			func(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) bool {
				ctx.SetValue("fingerprint", uuid.NewString())
				ctx.SetValue("anonymous", true)
				return true
			},
		),
	)
}

func WaitForShutdown() {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownCh
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	username := s.User()
	if username == "" {
		username = "anonymous"
	}

	m := tui.NewChatModel(username, s)
	return m, []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}
}
