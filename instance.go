package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync/atomic"
)

type listenersChan chan net.Listener
type status byte

const (
	failure status = iota
	success
)

var (
	count    atomic.Int64
	debug    atomic.Bool
	hostname = os.Getenv("HOST_NAME")
	port     = os.Getenv("PORT")
	addr     = os.Getenv("ADDR")
	respBase = "You are talking to instance %v:%v.\nThis is request %v to this instance and request %v to the cluster.\n"
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

func send(c net.Conn, stat status, msg string) {
	var msgBytes []byte
	msgBytes = append(msgBytes, byte(stat))
	msgBytes = append(msgBytes, []byte(msg)...)
	c.Write(msgBytes)
	debugLog("sent message: %q\n", msg)
}

func toggleDebug() {
	address := fmt.Sprintf("/tmp/req-count.instance.%v.%v.sock", addr, port)
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

func main() {
	debugFlag := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()
	debug.Store(*debugFlag)
	go toggleDebug()

	address := fmt.Sprintf("%v:%v", addr, port)
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
			defer debugLog("handled %v-th connection\n", i+1)
			defer c.Close()

			instanceCount := count.Load() + 1
			go count.Add(1)

			clusterCount := make([]byte, 1024)
			n, err := c.Read(clusterCount)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR]: failed to read from connection: %v\n", err)
				msg := fmt.Sprintf("unexpected read error: %v\n", err)
				send(c, failure, msg)
				return
			}

			clusterCount = clusterCount[:n]
			debugLog("read %v byte(s) for cluster count: %q\n", n, string(clusterCount))
			msg := fmt.Sprintf(respBase, hostname, port, instanceCount, string(clusterCount))
			send(c, success, msg)
		}(conn)
	}
}
