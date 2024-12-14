package main

import (
	"fmt"
	"net/http"
)

func main() {
	servers := []Server{
		NewSimpleServer("https://www.facebook.com"),
		NewSimpleServer("https://www.github.com"),
	}

	lb := NewLoadBalancer("8000", servers)
	handleRedirect := func(res http.ResponseWriter, req *http.Request) {
		lb.ServeProxy(res, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Listening on port %s\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
