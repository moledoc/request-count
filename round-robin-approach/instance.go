package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type counter struct{}
type listenersChan chan net.Listener

var (
	instanceCount atomic.Int64
	clusterCount  atomic.Int64
	debug         atomic.Bool
	host          = os.Getenv("HOST")
	port          = os.Getenv("PORT")
	syncRecvHost  = os.Getenv("RECV_HOST")
	syncRecvPort  = os.Getenv("RECV_PORT")
	syncSendHost  = os.Getenv("SEND_HOST")
	syncSendPort  = os.Getenv("SEND_PORT")
	respBase      = "You are talking to instance %v:%v.\nThis is request %v to this instance and request %v to the cluster.\n"
	//
	listenersSize               = 1
	listeners     listenersChan = listenersChan(make(chan net.Listener, listenersSize))
)

func (lc listenersChan) add(l net.Listener) {
	if len(lc)+1 > listenersSize {
		fmt.Fprintf(os.Stderr, "[ERROR]: too many listeners opened\n")
		l.Close()
		lc.close()
		os.Exit(1)
	}
	lc <- l
}

func (lc listenersChan) close() {
	if len(lc) == 0 {
		return
	}
	for i := 0; i < listenersSize; i++ {
		l := <-lc
		fmt.Fprintf(os.Stderr, "[INFO]: closing listener: %v\n", l.Addr().String())
		l.Close()
	}
	close(lc)
}
func debugLog(format string, a ...any) {
	if !debug.Load() {
		return
	}
	fmt.Fprintf(os.Stderr, format, a...)
}

func toggleDebug() {
	address := fmt.Sprintf("/tmp/instance.%v.%v.sock", strings.ReplaceAll(host, "http://", ""), port)
	if err := os.RemoveAll(address); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: failed to remove all from socket %q: %v\n", address, err)
		return
	}
	listenDebug, err := net.Listen("unix", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: listening %v failed: %v\n", address, err)
		return
	}
	listeners.add(listenDebug)
	for {
		conn, err := listenDebug.Accept()
		if err != nil {
			debugLog("[ERROR]: failed to get the next connection: %v\n", err)
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			debug.Swap(!debug.Load())
		}(conn)
	}
}

func syncRecv() {
	address := fmt.Sprintf("%v:%v", syncRecvHost, syncRecvPort)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: failed to listen address %q: %v\n", address, err)
		return
	}
	debugLog("listening address: %q\n", address)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR]: failed to accept connection: %v\n", err)
			return
		}
		go func(c net.Conn) {
			debugLog("connection to address %q established\n", address)
			defer debugLog("connection to address %q finished\n", address)
			defer c.Close()
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			buf := make([]byte, 8)
			_, err := c.Read(buf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR]: failed to read on receive: %v\n", err)
				return
			}
			recvClusterCount := int64(binary.LittleEndian.Uint64(buf))
			debugLog("from connection to address %q read: %v -> %v\n", address, buf, recvClusterCount)
			if recvClusterCount <= clusterCount.Load() {
				debugLog("address %q cluster count didn't change\n", address)
				return
			}
			syncSend(recvClusterCount)
			clusterCount.Store(recvClusterCount)
			debugLog("address %q stored new cluster count: %v \n", address, recvClusterCount)
		}(conn)
	}
}

func syncSend(newClusterCount int64) {
	address := fmt.Sprintf("%v:%v", syncSendHost, syncSendPort)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: sending new cluster count failed: %v\n", err)
		return
	}
	defer conn.Close()
	debugLog("dialed address: %q\n", address)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(newClusterCount))
	n, err := conn.Write(buf)
	debugLog("wrote %v bytes (%v) to address %q and the write had error: %v\n", n, buf, address, err)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING]: failed to send new cluster count: %v\n", err)
	}
}

func (*counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	newInstanceCount := instanceCount.Load() + 1
	newClusterCount := clusterCount.Load() + 1

	go instanceCount.Add(1)
	go clusterCount.Add(1)

	resp := fmt.Sprintf(respBase, host, port, newInstanceCount, newClusterCount)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(resp)))
	fmt.Fprintf(w, "%v", resp)
	go syncSend(newClusterCount)
}

func serve(conn net.Conn) {
	newInstanceCount := instanceCount.Load() + 1
	newClusterCount := clusterCount.Load() + 1

	go instanceCount.Add(1)
	go clusterCount.Add(1)

	resp := fmt.Sprintf(respBase, host, port, newInstanceCount, newClusterCount)
	_, err := conn.Write([]byte(resp))
	if err != nil {
		debugLog("failed to serve a response: %q\n", resp)
	}
	syncSend(newClusterCount)
}

func main() {
	debugFlag := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()
	debug.Store(*debugFlag)

	if port == "" || syncRecvPort == "" || syncSendPort == "" {
		fmt.Fprintf(os.Stderr, "[ERROR]: missing port configuration\n")
		os.Exit(1)
	}

	go toggleDebug()

	sigtermCh := make(chan os.Signal)
	signal.Notify(sigtermCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigtermCh
		listeners.close()
		os.Exit(0)
	}()

	go syncRecv()
	address := fmt.Sprintf("%v:%v", host, port)
	debugLog("serving on address %q\n", address)
	// http.ListenAndServe(address, &counter{})

	listen, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: listening %v failed: %v\n", address, err)
		return
	}
	for i := 0; ; i++ {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR]: failed to get the next connection: %v\n", err)
			return
		}
		debugLog("accepted %v-th connection\n", i+1)
		go func(c net.Conn) {
			debugLog("opened %v-th connection", i+1)
			defer debugLog("handled %v-th connection\n", i+1)
			defer c.Close()
			serve(c)
		}(conn)
	}

}
