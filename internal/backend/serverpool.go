package backend

import (
	"log"
	"net/url"
	"sync/atomic"
)

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) NewPeerIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) MarkBackendStatus(backendURL *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendURL.String() {
			b.SetAlive(alive)
			return
		}
	}
}

func (s *ServerPool) GetNextPeer() *Backend {
	peerInd := s.NewPeerIndex()
	for i := peerInd; i < peerInd+len(s.backends); i++ {
		idx := i % len(s.backends)
		if s.backends[idx].IsAlive() {
			if i != peerInd {
				atomic.StoreUint64(&s.current, uint64(idx))
			}

			return s.backends[idx]
		}
	}

	return nil
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)

		if !alive {
			status = "down"
		}

		log.Printf("%s [%s]\n", b.URL, status)
	}
}
