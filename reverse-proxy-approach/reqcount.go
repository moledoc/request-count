package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
)

type counter struct{}
type listenersChan chan net.Listener

var (
	count atomic.Int64
	debug atomic.Bool
	order atomic.Int64
	addrs []string
	addr  = os.Getenv("ADDR")
	port  = os.Getenv("PORT")
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
	address := fmt.Sprintf("/tmp/req-count.proxy.%v.%v.sock", addr, port)
	if err := os.RemoveAll(address); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: failed to remove all from socket %q: %v\n", address, err)
		return
	}
	listenDebug, err := net.Listen("unix", address)
	if err != nil {
		debugLog("[ERROR]: listening %v failed: %v\n", address, err)
		return
	}
	listeners.add(listenDebug)
	for {
		conn, err := listenDebug.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR]: failed to get the next connection: %v\n", err)
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			debug.Swap(!debug.Load())
		}(conn)
	}
}

func (*counter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clusterCount := count.Load() + 1
	go count.Add(1)

	idx := order.Load()
	order.Swap((idx + 1) % int64(len(addrs)))
	addr := addrs[idx]
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		debugLog("dial error: %v", err)
		fmt.Fprintf(w, "request failed: %v\n", err)
		return
	}
	defer conn.Close()

	msg := fmt.Sprintf("%v", clusterCount)
	n, err := conn.Write([]byte(msg))
	if err != nil {
		debugLog("write error: %v", err)
		fmt.Fprintf(w, "request failed: %v\n", err)
		return
	}
	debugLog("wrote %v bytes: %q\n", n, msg)

	respBody := make([]byte, 1024)
	n, err = conn.Read(respBody)
	if err != nil || n < 2 {
		debugLog("read error: %v", err)
		fmt.Fprintf(w, "request failed: %v\n", err)
		return
	}
	switch respBody[0] {
	case '0':
		w.WriteHeader(http.StatusInternalServerError)
	case '1':
		w.WriteHeader(http.StatusOK)
	}
	respBody = respBody[1:n]
	debugLog("read %v bytes: %q\n", n, string(respBody))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(respBody)))
	fmt.Fprintf(w, "%v", string(respBody))
}

func main() {
	debugFlag := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()
	debug.Store(*debugFlag)
	go toggleDebug()

	sigtermCh := make(chan os.Signal)
	signal.Notify(sigtermCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigtermCh
		listeners.close()
		os.Exit(0)
	}()

	instanceAddrs := os.Getenv("INSTANCES")
	if len(instanceAddrs) == 0 {
		fmt.Fprintf(os.Stderr, "invalid instances addresses\n")
		return
	}
	addrs = strings.Split(instanceAddrs, ",")

	http.ListenAndServe(fmt.Sprintf("%v:%v", addr, port), &counter{})
}
