package main

import (
    "flag"
    "fmt"
    lb "github.com/denieryd/SimpleLoadBalancer/internal/loadbalancer"
    "github.com/denieryd/SimpleLoadBalancer/internal/proxy"
    log "github.com/sirupsen/logrus"
    "net/http"
    "strings"
)

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
    if err := proxy.SetupProxyServers(tokens); err != nil {
        log.Fatal(err)
    }
    log.Info("Proxies are set up")

    server := http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: http.HandlerFunc(lb.LoadBalance),
    }

    go lb.HealthCheck()

    log.Infof("Backends are: %v. Load balancer starts at :%v port", serverList, port)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
