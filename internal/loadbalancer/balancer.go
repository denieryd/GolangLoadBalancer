package loadbalancer

import (
    "github.com/denieryd/SimpleLoadBalancer/internal/backend"
    log "github.com/sirupsen/logrus"
    "net/http"
    "time"
)

const (
    ATTEMPTS int = iota
    RETRY
)

var ServerPool backend.ServerPool

func GetAttemptsFromContext(r *http.Request) int {
    if attempts, ok := r.Context().Value(ATTEMPTS).(int); ok {
        return attempts
    }
    return 1
}

func GetRetryFromContext(r *http.Request) int {
    if retry, ok := r.Context().Value(RETRY).(int); ok {
        return retry
    }
    return 0
}

func LoadBalance(w http.ResponseWriter, r *http.Request) {
    attempts := GetAttemptsFromContext(r)
    if attempts > 3 {
        log.Infof("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
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
    log.Infof("first backends health check will start in %v\n", timeToStart)

    t := time.NewTicker(timeToStart)
    for {
        select {
        case <-t.C:
            log.Info("Start health checking")
            ServerPool.HealthCheck()
            log.Info("Health check completed")
        }
    }
}
