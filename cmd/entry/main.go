package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type counter struct{}

var (
	clusterCount atomic.Int64
	order        atomic.Int64
	instances    []string
	host         = os.Getenv("HOST")
	port         = os.Getenv("PORT")
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
		instances = append(instances[:idx], instances[idx+1:]...)
		order.Swap(idx % int64(len(instances)))
		return
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(newClusterCount))
	_, err = conn.Write(buf)
	if err != nil {
		fmt.Fprintf(w, "internal write error: %v\n", err)
		return
	}

	resp := make([]byte, 256)
	n, err := conn.Read(resp)
	if err != nil {
		fmt.Fprintf(w, "internal read error: %v\n", err)
		return
	}
	resp = resp[:n]

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(resp)))
	fmt.Fprintf(w, "%v", string(resp))

}

func main() {
	instances = strings.Split(os.Getenv("INSTANCES"), ",")
	address := fmt.Sprintf("%v:%v", host, port)
	http.ListenAndServe(address, &counter{})
}
