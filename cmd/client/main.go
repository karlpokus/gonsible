package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/karlpokus/gonsible"
)

var version = flag.Bool("version", false, "print version and exit")

func main() {
	flag.Parse()
	if *version {
		fmt.Println(gonsible.Version)
		return
	}
	client, err := gonsible.NewClient(&gonsible.ClientConfig{
		PrivateKeyWithPassphrase: os.Getenv("SSH_PWD"),
		User:                     "bixa",
		Host:                     "127.0.0.1",
		Port:                     "2222",
		PrivateKey:               "testdata/client_rsa",
		Knownhosts:               "testdata/known_hosts",
	})
	if err != nil {
		fmt.Printf("Failed to run client %s", err)
		return
	}
	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create session: %s", err)
		return
	}
	defer session.Close()
	session.Output(os.Stdout, os.Stderr)
	err = session.Run("hostname")
	if err != nil && err != io.EOF {
		fmt.Printf("run failed: %s", err)
		return
	}
	fmt.Println("done")
}
