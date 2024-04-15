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

const (
	version = 1
)

var (
	clusterCount atomic.Int64
	order        atomic.Int64
	instances    []string
	host         = os.Getenv("HOST")
	port         = os.Getenv("PORT")
)

func (*counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	newClusterCount := clusterCount.Load() + 1
	go clusterCount.Add(1)

	idx := order.Load()
	order.Swap((idx + 1) % int64(len(instances)))
	instance := instances[idx]

	conn, err := net.Dial("tcp", instance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "request to %q failed: %q\ndiscarding %q, please send new request\n", instance, err, instance)
		instances = append(instances[:idx], instances[idx+1:]...)
		order.Swap(idx % int64(len(instances)))
		return
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 11)
	// store version (in big-endianess) manually for now
	buf[0] = version % 10
	buf[1] = version / 10
	buf[2] = 1 // success byte

	binary.BigEndian.PutUint64(buf[3:], uint64(newClusterCount))
	n, err := conn.Write(buf)
	if err != nil || n != 11 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal write error: %v\n", err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp := make([]byte, 3+256)
	n, err = conn.Read(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal read error: %v\n", err)
		return
	}
	readVersion := int(buf[1]*10 + buf[0])
	if readVersion != version {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal read error: internal communication version doesn't match: got %q, expected %q\n", readVersion, version)
		return
	}
	if resp[2] == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Length", fmt.Sprintf("%v", len(resp)-3))
		fmt.Fprintf(w, string(resp[3:]))
		return
	}
	resp = resp[3:n]

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(resp)))
	fmt.Fprintf(w, "%v", string(resp))

}

func main() {
	instances = strings.Split(os.Getenv("INSTANCES"), ",")
	address := fmt.Sprintf("%v:%v", host, port)
	http.ListenAndServe(address, &counter{})
}
