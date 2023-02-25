package loadbalancer

import (
    "github.com/denieryd/SimpleLoadBalancer/internal/backend"
    "log"
    "net/http"
    "time"
)

const (
    Attempts int = iota
    Retry
)

var ServerPool backend.ServerPool

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

    peer := ServerPool.GetNextPeer()
    if peer != nil {
        peer.ReverseProxy.ServeHTTP(w, r)
        return
    }

    http.Error(w, "Service not available", http.StatusServiceUnavailable)

}

func HealthCheck() {
    timeToStart := time.Second * 30
    log.Printf("first health check in %v\n", timeToStart)

    t := time.NewTicker(timeToStart)
    for {
        select {
        case <-t.C:
            log.Println("Start health checking")
            ServerPool.HealthCheck()
            log.Println("Health check completed")
        }
    }
}
