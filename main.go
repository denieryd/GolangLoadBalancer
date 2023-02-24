package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "sync"
    "sync/atomic"
    "time"
)

const (
    Attempts int = iota
    Retry
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

func GetAttemptsFromContext(r *http.Request) int {
    if attempts, ok := r.Context().Value(Attempts).(int); ok {
        return attempts
    }

    return 1
}

func GetRetryFromContext(r *http.Request) int {
    if retry, ok := r.Context().Value(Retry).(int); ok {
        return retry
    }
    return 0
}

func LoadBalance(w http.ResponseWriter, r *http.Request) {
    attempts := GetAttemptsFromContext(r)
    if attempts > 3 {
        log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
        http.Error(w, "Service not available", http.StatusServiceUnavailable)
        return
    }

    peer := serverPool.GetNextPeer()
    if peer != nil {
        peer.ReverseProxy.ServeHTTP(w, r)
        return
    }

    http.Error(w, "Service not available", http.StatusServiceUnavailable)

}

var serverPool ServerPool

func isBackendAlive(u *url.URL) bool {
    timeout := 2 * time.Second
    conn, err := net.DialTimeout("tcp", u.Host, timeout)
    if err != nil {
        log.Println("Site unreachable, error: ", err)
        return false
    }

    defer conn.Close()
    return true
}

func healthCheck() {
    t := time.NewTicker(time.Minute * 2)
    for {
        select {
        case <-t.C:
            log.Println("Start health checking")
            serverPool.HealthCheck()
            log.Println("Health check completed")
        }
    }
}

func main() {
    var serverList string
    var port int

    flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
    flag.IntVar(&port, "port", 3030, "Port to serve")
    flag.Parse()

    if len(serverList) == 0 {
        log.Fatal("Provide at least one backend to make load balancer works")
    }

    tokens := strings.Split(serverList, ",")
    for _, tok := range tokens {
        serverURL, err := url.Parse(tok)
        if err != nil {
            log.Fatal(err)
        }

        proxy := httputil.NewSingleHostReverseProxy(serverURL)
        proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
            log.Printf("[%s] %s\n", serverURL.Host, e.Error())
            retries := GetRetryFromContext(r)
            if retries < 3 {
                select {
                case <-time.After(10 * time.Millisecond):
                    ctx := context.WithValue(r.Context(), Retry, retries+1)
                    proxy.ServeHTTP(w, r.WithContext(ctx))
                }
                return
            }

            serverPool.MarkBackendStatus(serverURL, false)

            attempts := GetAttemptsFromContext(r)
            log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
            ctx := context.WithValue(r.Context(), Attempts, attempts+1)
            LoadBalance(w, r.WithContext(ctx))
        }

        serverPool.AddBackend(&Backend{
            URL:          serverURL,
            Alive:        true,
            ReverseProxy: proxy,
        })

        log.Printf("Configured server: %s\n", serverURL)

    }
    server := http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: http.HandlerFunc(LoadBalance),
    }

    go healthCheck()

    log.Printf("Load Balancer started at :%d\n", port)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
