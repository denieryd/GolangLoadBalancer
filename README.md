# Purpose
Just to train some skills in writing applications using Go.

# SimpleLoadBalancer
This load balancer uses RoundRobin algorithm to send requests into set of backends and support
retries too.

It also performs active cleaning and passive recovery for unhealthy backends.

# How to use
```bash
Usage:
  -backends string
        Load balanced backends, use commas to separate
  -port int
        Port to serve (default 3030)
```

Example:

To add followings as load balanced backends
- http://localhost:3031
- http://localhost:3032
- http://localhost:3033
- http://localhost:3034
```bash
./lb --backends=http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034
```