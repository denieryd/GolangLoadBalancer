package backend

import (
    log "github.com/sirupsen/logrus"
    "net"
    "net/http/httputil"
    "net/url"
    "time"
)

type IBackend interface {
    SetAlive(bool)
    IsAlive() bool
    GetServerURL() *url.URL
    GetReverseProxy() *httputil.ReverseProxy
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
