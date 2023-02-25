package backend

import (
    "net/http/httputil"
    "net/url"
    "sync"
)

type Backend struct {
    url          *url.URL
    alive        bool
    mux          sync.RWMutex
    reverseProxy *httputil.ReverseProxy
}

func CreateNewBackend(serverURL *url.URL, alive bool, reverseProxy *httputil.ReverseProxy) Backend {
    return Backend{
        url:          serverURL,
        alive:        alive,
        mux:          sync.RWMutex{},
        reverseProxy: reverseProxy,
    }
}

func (b *Backend) SetAlive(alive bool) {
    b.mux.Lock()
    b.alive = alive
    b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
    b.mux.RLock()
    alive := b.alive
    b.mux.RUnlock()
    return alive
}

func (b *Backend) GetServerURL() *url.URL {
    return b.url
}

func (b *Backend) GetReverseProxy() *httputil.ReverseProxy {
    return b.reverseProxy
}
