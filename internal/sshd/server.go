package sshd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	*Conf
	net.Listener
	State
}

type Conf struct {
	AuthorizedKeys, HostKey string
	Verbose                 bool
}

func New(conf *Conf) *Server {
	return &Server{conf, nil, State{}}
}

func (s *Server) Quit() {
	s.ListeningState(false)
}

func (s *Server) Run(ready chan<- bool) error { // TODO: split into smaller funcs
	if !s.Verbose {
		log.SetOutput(ioutil.Discard)
	}

	// read authorized_keys
	authorizedKeysBytes, err := ioutil.ReadFile(s.Conf.AuthorizedKeys)
	if err != nil {
		return fmt.Errorf("Failed to load authorized_keys, err: %s", err)
	}
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			return err
		}
		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}
	log.Println("authorized_keys read")

	// set server conf
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() == "testuser" && string(pass) == "tiger" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		// Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}
	log.Println("server conf set")

	// add hostkey
	privateBytes, err := ioutil.ReadFile(s.Conf.HostKey)
	if err != nil {
		return fmt.Errorf("Failed to load private key: %s", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return fmt.Errorf("Failed to parse private key: %s", err)
	}
	config.AddHostKey(private)
	log.Println("hostkey added")

	// listen and accept
	host := fmt.Sprintf("%s:%s", "127.0.0.1", "2222")
	s.Listener, err = net.Listen("tcp", host)
	if err != nil {
		return fmt.Errorf("failed to listen for connection: %s", err)
	}
	log.Printf("listening on %s", host)
	if ready != nil {
		ready <- true
	}
	s.ListeningState(true)

	for s.Listening() {
		nConn, err := s.Listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept incoming connection: %s", err)
		}
		// handshake
		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			return fmt.Errorf("failed to handshake: %s", err)
		}
		//defer conn.Close() already closed via channel below
		log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
	return nil
}

func handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
		go func(in <-chan *ssh.Request) {
			for req := range in {
				if req.Type == "exec" {
					command := string(req.Payload[4 : req.Payload[3]+4])
					cmd := exec.Command(command)
					cmd.Stdout = channel
					cmd.Stderr = channel
					//cmd.Stdin = channel

					log.Printf("running %s", command)
					err := cmd.Run()
					if err != nil {
						log.Printf("could not start command (%s)", err)
						continue
					}
					channel.Close()
					log.Printf("session closed")
				}
				req.Reply(req.Type == "exec", nil)
			}
		}(requests)
	}
}
