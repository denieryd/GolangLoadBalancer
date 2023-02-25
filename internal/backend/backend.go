package backend

import (
    log "github.com/sirupsen/logrus"
    "net"
    "net/http/httputil"
    "net/url"
    "sync"
    "time"
)

type Backend struct {
    URL          *url.URL
    Alive        bool
    mux          sync.RWMutex
    ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) {
    b.mux.Lock()
    b.Alive = alive
    b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
    b.mux.RLock()
    alive := b.Alive
    b.mux.RUnlock()
    return alive
}

func isBackendAlive(u *url.URL) bool {
    timeout := 2 * time.Second
    conn, err := net.DialTimeout("tcp", u.Host, timeout)
    if err != nil {
        log.Warnf("Site unreachable, error: %v", err)
        return false
    }

    defer conn.Close()
    return true
}
