package gonsible

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var Version = "vX.Y.Z" // defined at compile time

type SessionConfig struct {
	User, Host, PrivateKey, PrivateKeyWithPassphrase, Knownhosts string
}

func NewSession(conf *SessionConfig) error {
	key, err := ioutil.ReadFile(conf.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to read private key: %s", err)
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(conf.PrivateKeyWithPassphrase))
	if err != nil {
		return fmt.Errorf("unable to parse private key: %s", err)
	}
	hostKeyCallback, err := knownhosts.New(conf.Knownhosts)
	if err != nil {
		return fmt.Errorf("could not create hostkeycallback function: %s", err)
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", conf.Host), &ssh.ClientConfig{
		User: conf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("unable to connect: %s", err)
	}
	defer client.Close()

	ss, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("unable to create SSH session: %s", err)
	}
	defer ss.Close()

	var buf bytes.Buffer
	ss.Stdout = &buf
	if err = ss.Run("uptime"); err != nil {
		return err
	}
	fmt.Printf("uptime: %s", buf.String())
	return nil
}
