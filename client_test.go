package gonsible

import (
	"io"
	"testing"

	"github.com/karlpokus/bufw"
	"github.com/karlpokus/gonsible/internal/sshd"
)

func TestClient(t *testing.T) {
	server := sshd.New(&sshd.Conf{
		AuthorizedKeys: "testdata/authorized_keys",
		HostKey:        "testdata/server_rsa",
	})
	defer server.Quit()
	ready := make(chan bool)
	go func() {
		err := server.Run(ready)
		if err != nil {
			t.Errorf("Failed to run server %s", err)
			return
		}
	}()
	<-ready

	client, err := NewClient(&ClientConfig{
		User:       "bixa",
		Host:       "127.0.0.1",
		Port:       "2222",
		PrivateKey: "testdata/client_rsa",
		Knownhosts: "testdata/known_hosts",
	})
	if err != nil {
		t.Errorf("Failed to run client %s", err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		t.Errorf("Failed to create session: %s", err)
		return
	}
	defer session.Close()

	// setpipes and capture io
	w := bufw.New(true)
	session.Output(w, w)

	go func() {
		err = session.Run("hostname")
		if err != nil && err != io.EOF {
			t.Errorf("run failed: %s", err)
			return
		}
	}()

	err = w.Wait()
	if err != nil {
		t.Errorf("Expected nil, got %s", err)
	}
	output := w.String()
	expected := "mb.lan"
	if output != expected {
		t.Errorf("Expected %s, got %s", expected, output)
		return
	}
}
