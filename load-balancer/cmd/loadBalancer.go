package main

import "net/http"

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func (lb *LoadBalancer) GetNextAvailableServer() Server {
	if len(lb.servers) == 0 {
		return nil
	}
	nextServer := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !nextServer.IsAlive() {
		nextServer = lb.servers[lb.roundRobinCount%len(lb.servers)]
		lb.roundRobinCount++
	}
	lb.roundRobinCount++
	return nextServer
}

func (lb *LoadBalancer) ServeProxy(res http.ResponseWriter, req *http.Request) {
	nextServer := lb.GetNextAvailableServer()
	if nextServer == nil {
		res.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	nextServer.Serve(res, req)
}
