
toggle-debug:
	echo "" | nc -U /tmp/instance..8081.sock
	echo "" | nc -U /tmp/instance..8082.sock

run:
	HOST="" PORT="8081" RECV_HOST="" RECV_PORT="8181" SEND_HOST="" SEND_PORT="8182" go run instance.go
	HOST="" PORT="8082" RECV_HOST="" RECV_PORT="8182" SEND_HOST="" SEND_PORT="8181" go run instance.go

down:
	pgrep instance | parallel 'kill -9 {}'