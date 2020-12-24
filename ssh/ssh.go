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

type Server struct {
	Port        uint
	Address     string
	IdleTimeout time.Duration
	HostKeyFile string
}

func (s *Server) Start() {
	path := fmt.Sprintf("%s:%d", s.Address, s.Port)
	ssh.Handle(handler)
	log.Printf("Starting deslang server @ %s", path)
	log.Fatal(ssh.ListenAndServe(path, nil, ssh.HostKeyFile(s.HostKeyFile)))
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
