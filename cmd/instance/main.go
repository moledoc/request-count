package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type listenersChan chan net.Listener

var (
	instanceCount atomic.Int64
	debug         atomic.Bool
	debugFile     *os.File

	host     = os.Getenv("HOST")
	port     = os.Getenv("PORT")
	hostName = func(host string, port string) string {
		hostname := os.Getenv("HOSTNAME")
		if len(hostname) == 0 {
			hostname = host
		}
		return fmt.Sprintf("%v:%v", hostname, port)
	}(host, port)
	respBase = "You are talking to instance %v.\nThis is request %v to this instance and request %v to the cluster.\n"
	//
	listenersSize               = 1
	listeners     listenersChan = listenersChan(make(chan net.Listener, listenersSize))
)

func (lc listenersChan) add(l net.Listener) {
	if len(lc)+1 > listenersSize {
		fmt.Fprintf(debugFile, "[ERROR]: too many listeners opened\n")
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
		fmt.Fprintf(debugFile, "[INFO]: closing listener: %v\n", l.Addr().String())
		l.Close()
	}
	close(lc)
}

func debugLog(format string, a ...any) {
	if !debug.Load() {
		return
	}
	fmt.Fprintf(debugFile, format, a...)
}

func toggleDebug() {
	var err error
	debugFilename := fmt.Sprintf("/tmp/instance.%v.%v.debug.log", host, port)
	debugFile, err = os.OpenFile(debugFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(debugFile, "[WARNING]: failed to open %q for logging: %v\nSetting stderr as logs output\n", debugFilename, err)
		debugFile = os.Stderr
	}
	address := fmt.Sprintf("/tmp/instance.%v.%v.sock", host, port)
	if err := os.RemoveAll(address); err != nil {
		fmt.Fprintf(debugFile, "[ERROR]: failed to remove all from socket %q: %v\n", address, err)
		return
	}
	listenDebug, err := net.Listen("unix", address)
	if err != nil {
		fmt.Fprintf(debugFile, "[ERROR]: listening %v failed: %v\n", address, err)
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

	sigtermCh := make(chan os.Signal)
	signal.Notify(sigtermCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigtermCh
		listeners.close()
		os.Exit(0)
	}()

	address := fmt.Sprintf("%v:%v", host, port)
	debugLog("serving on address %q\n", address)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintf(debugFile, "[ERROR]: listening %v failed: %v\n", address, err)
		return
	}
	for i := 0; ; i++ {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Fprintf(debugFile, "[ERROR]: failed to get the next connection: %v\n", err)
			return
		}
		debugLog("accepted %v-th connection\n", i+1)
		go func(c net.Conn, i int) {
			debugLog("opened %v-th connection", i+1)
			defer debugLog("handled %v-th connection\n", i+1)
			defer c.Close()

			newInstanceCount := instanceCount.Load() + 1
			go instanceCount.Add(1)

			c.SetReadDeadline(time.Now().Add(5 * time.Second))
			buf := make([]byte, 8)
			n, err := c.Read(buf)
			clusterCount := int64(binary.LittleEndian.Uint64(buf))
			if err != nil || n != 8 {
				debugLog("failed to read a response: %v\n", err)
				return
			}

			resp := fmt.Sprintf(respBase, hostName, newInstanceCount, clusterCount)
			respBytes := []byte(resp)
			c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			n, err = c.Write(respBytes)
			debugLog("wrote %v/%v bytes (%v): %v\n", n, len(respBytes), respBytes, err)
			if err != nil {
				fmt.Fprintf(debugFile, "[ERROR]: failed to send new instance count: %v\n", err)
			}
		}(conn, i)
	}

}
