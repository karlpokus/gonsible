package main

import (
	"flag"
	"fmt"

	"github.com/karlpokus/gonsible/internal/sshd"
)

var verbose = flag.Bool("v", false, "toggle logs")

func main() {
	flag.Parse()
	server := sshd.New(&sshd.Conf{
		AuthorizedKeys: "testdata/authorized_keys",
		HostKey:        "testdata/server_rsa",
		Verbose:        *verbose,
	})
	defer server.Quit()

	err := server.Run(nil)
	if err != nil {
		fmt.Printf("Failed to run server: %s", err)
		return
	}
}
