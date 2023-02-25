package proxy

import (
    "context"
    "github.com/denieryd/SimpleLoadBalancer/internal/backend"
    lb "github.com/denieryd/SimpleLoadBalancer/internal/loadbalancer"
    "log"
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
            log.Printf("[%s] %s\n", serverURL.Host, e.Error())
            retries := lb.GetRetryFromContext(r)
            if retries < 3 {
                select {
                case <-time.After(10 * time.Millisecond):
                    ctx := context.WithValue(r.Context(), lb.Retry, retries+1)
                    reverseProxy.ServeHTTP(w, r.WithContext(ctx))
                }
                return
            }

            lb.ServerPool.MarkBackendStatus(serverURL, false)

            attempts := lb.GetAttemptsFromContext(r)
            log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
            ctx := context.WithValue(r.Context(), lb.Attempts, attempts+1)
            lb.LoadBalance(w, r.WithContext(ctx))
        }

        lb.ServerPool.AddBackend(&backend.Backend{
            URL:          serverURL,
            Alive:        true,
            ReverseProxy: reverseProxy,
        })

        log.Printf("Configured server: %s\n", serverURL)
    }

    return nil
}
