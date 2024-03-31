package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

type counter struct{}

var (
	clusterCount atomic.Int64
	order        atomic.Int64
	instances    []string
	host         = os.Getenv("HOST")
	port         = os.Getenv("PORT")
	respBase     = "You are talking to instance %v.\nThis is request %v to this instance and request %v to the cluster.\n"
)

func (*counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	newClusterCount := clusterCount.Load() + 1
	go clusterCount.Add(1)

	idx := order.Load()
	order.Swap((idx + 1) % int64(len(instances)))
	instance := instances[idx]

	conn, err := net.Dial("tcp", instance)
	if err != nil {
		fmt.Fprintf(w, "request to %q failed: %q\ndiscarding %q, please send new request\n", instance, err, instance)
		instances = append(instances[:idx], instances[idx+1:]...) // MAYBE: TODO: lock/unlock resource
		order.Swap(idx % int64(len(instances)))
		// MAYBE: TODO: when instances is empty
		return
	}
	defer conn.Close()

	buf := make([]byte, 8)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Fprintf(w, "failed to read a response: %v\n", err)
		return
	}
	instanceCount := int64(binary.LittleEndian.Uint64(buf))

	resp := fmt.Sprintf(respBase, instance, instanceCount, newClusterCount)
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
