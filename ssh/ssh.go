package ssh

import (
	"context"
	"fmt"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"io"
	"log"
	"os/exec"
	"time"
)

type Config struct {
	Port        uint
	Address     string
	IdleTimeout time.Duration
	HostKeyFile string
}

type server struct {
	port        uint
	address     string
	idleTimeout time.Duration
	hostKeyFile string
}

func NewServer(c *Config) *server {
	return &server{
		port:        c.Port,
		address:     c.Address,
		idleTimeout: c.IdleTimeout,
		hostKeyFile: c.HostKeyFile,
	}
}

func (s *server) Start() {
	path := fmt.Sprintf("%s:%d", s.address, s.port)
	ssh.Handle(handler)
	log.Printf("Starting deslang server @ %s", path)
	log.Fatal(ssh.ListenAndServe(path, nil, ssh.HostKeyFile(s.hostKeyFile)))
}

func handler(sesh ssh.Session) {
	cmdCtx, cancelCmd := context.WithCancel(sesh.Context())
	defer cancelCmd()

	cmd := exec.CommandContext(cmdCtx, "bin/deslang")

	f, err := pty.Start(cmd)
	if err != nil {
		io.WriteString(sesh, err.Error())
	}

	defer f.Close()

	go func() {
		io.Copy(f, sesh)
	}()

	io.Copy(sesh, f)
}
