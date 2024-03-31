package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

type counter struct{}

var (
	order     atomic.Int64
	instances []string
	host      = os.Getenv("HOST")
	port      = os.Getenv("PORT")
)

func (*counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idx := order.Load()
	order.Swap((idx + 1) % int64(len(instances)))
	instance := instances[idx]
	conn, err := net.Dial("tcp", instance)
	if err != nil {
		fmt.Fprintf(w, "request failed: %v\n", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Fprintf(w, "failed to read a response: %v\n", err)
		return
	}
	buf = buf[:n]

	resp := string(buf)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(resp)))
	fmt.Fprintf(w, "%v", resp)
}

func main() {
	instances = strings.Split(os.Getenv("INSTANCES"), ",")
	address := fmt.Sprintf("%v:%v", host, port)
	http.ListenAndServe(address, &counter{})
}
