package proxy

import (
    "context"
    "github.com/denieryd/SimpleLoadBalancer/internal/backend"
    lb "github.com/denieryd/SimpleLoadBalancer/internal/loadbalancer"
    log "github.com/sirupsen/logrus"
    "net/http"
    "net/http/httputil"
    "net/url"
    "time"
)

func SetupProxyServers(serverTokens []string) error {
    for _, token := range serverTokens {
        serverURL, err := url.Parse(token)
        if err != nil {
            return err
        }

        reverseProxy := httputil.NewSingleHostReverseProxy(serverURL)
        reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
            log.Infof("[%s] %s", serverURL.Host, e.Error())
            retries := lb.GetRetryFromContext(r)
            if retries < 3 {
                select {
                case <-time.After(10 * time.Millisecond):
                    ctx := context.WithValue(r.Context(), lb.RETRY, retries+1)
                    reverseProxy.ServeHTTP(w, r.WithContext(ctx))
                }
                return
            }

            lb.ServerPool.MarkBackendStatus(serverURL, false)

            attempts := lb.GetAttemptsFromContext(r)
            log.Infof("%s(%s) Attempting retry %d", r.RemoteAddr, r.URL.Path, attempts)
            ctx := context.WithValue(r.Context(), lb.ATTEMPTS, attempts+1)
            lb.LoadBalance(w, r.WithContext(ctx))
        }

        b := backend.CreateNewBackend(serverURL, true, reverseProxy)
        var ib backend.IBackend = &b
        lb.ServerPool.AddBackend(&ib)

        log.Infof("Configured server: %s", serverURL)
    }

    return nil
}
