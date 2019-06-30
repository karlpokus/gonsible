package gonsible

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var Version = "vX.Y.Z" // defined at compile time

type Client struct {
	*ssh.Client
}

type Session struct {
	*ssh.Session
}

type ClientConfig struct {
	User, Host, Port, PrivateKey, PrivateKeyWithPassphrase, Knownhosts string
}

func (c *Client) NewSession() (*Session, error) {
	session, err := c.Client.NewSession()
	return &Session{session}, err
}

func (s *Session) Close() error {
	return s.Session.Close()
}

func (s *Session) Run(cmd string) error {
	return s.Session.Run(cmd)
}

func (s *Session) Output(outw, errw io.Writer) error {
	outpipe, err := s.Session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stdout for session: %s", err)
	}
	go io.Copy(outw, outpipe)

	errpipe, err := s.Session.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stderr for session: %s", err)
	}
	go io.Copy(errw, errpipe)
	return nil
}

func NewClient(conf *ClientConfig) (*Client, error) {
	key, err := ioutil.ReadFile(conf.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %s", err)
	}
	var signer ssh.Signer
	if conf.PrivateKeyWithPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(conf.PrivateKeyWithPassphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(key)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %s", err)
	}
	hostKeyCallback, err := knownhosts.New(conf.Knownhosts)
	if err != nil {
		return nil, fmt.Errorf("could not create hostkeycallback function: %s", err)
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", conf.Host, conf.Port), &ssh.ClientConfig{
		User: conf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		//HostKeyCallback: ssh.FixedHostKey(pub),
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("Dial err: %s", err)
	}
	return &Client{client}, nil
}
