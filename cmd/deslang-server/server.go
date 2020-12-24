package main

import (
	"flag"
	"fmt"
	"github.com/despreston/deslang/ssh"
	"os"
	"time"
)

func main() {
	var hostKeyFile string
	var port uint
	var address string

	flag.StringVar(&hostKeyFile, "key", "", "Host private key file")
	flag.StringVar(&address, "addr", "", "Address to listen on")
	flag.UintVar(&port, "port", 3000, "Port to listen on")
	flag.Parse()

	if len(hostKeyFile) < 1 {
		fmt.Println("Provide the path to the host public key with the -key flag.")
		os.Exit(64)
	}

	config := ssh.Config{
		Port:        port,
		Address:     address,
		IdleTimeout: time.Duration(60 * time.Second),
		HostKeyFile: hostKeyFile,
	}

	ssh.NewServer(&config).Start()
}
