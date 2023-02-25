package backend

import (
    log "github.com/sirupsen/logrus"
    "net/url"
    "sync/atomic"
)

type ServerPool struct {
    backends []*Backend
    current  uint64
}

const (
    BACKEND_STATUS_UP   = "up"
    BACKEND_STATUS_DOWN = "down"
)

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
        alive := isBackendAlive(b.URL)
        b.SetAlive(alive)

        status := BACKEND_STATUS_UP
        if !alive {
            status = BACKEND_STATUS_DOWN
        }

        log.Infof("%s [%s]\n", b.URL, status)
    }
}
