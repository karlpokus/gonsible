package sshd

import "sync"

type State struct {
	sync.Mutex
	listening bool
}

func (s *State) Listening() bool {
	s.Lock()
	defer s.Unlock()
	return s.listening
}

func (s *State) ListeningState(b bool) {
	s.Lock()
	defer s.Unlock()
	s.listening = b
}
