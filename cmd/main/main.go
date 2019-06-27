package main

import (
  "fmt"
  "flag"
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
  err := gonsible.NewSession(&gonsible.SessionConfig{
    User: "bixa",
    Host: "165.227.157.210",
    PrivateKey: "/Users/pokus/.ssh/do_rsa",
    PrivateKeyWithPassphrase: os.Getenv("SSH_PWD"),
    Knownhosts: "/Users/pokus/.ssh/known_hosts",
  })
  if err != nil {
    fmt.Printf("%s", err)
    return
  }
  fmt.Println("done")
}
