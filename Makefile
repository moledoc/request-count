
toggle-debug:
	echo "" | nc -U /tmp/req-count.proxy..8081.sock
	echo "" | nc -U /tmp/req-count.instance..8082.sock
	echo "" | nc -U /tmp/req-count.instance..8083.sock

run:
	HOST_NAME=host1 PORT=8082 go run instance.go &
	HOST_NAME=host2 PORT=8083 go run instance.go &
	HOST_NAME=host3 PORT=8084 go run instance.go &
	PORT="8081" INSTANCES=":8082,:8083,:8084" go run reqcount.go &

# IMPROVEME:
down:
	# pgrep "instance" | parallel 'kill -9 {}' > /dev/null
	pgrep "instance" | xargs -I {} kill -9 "{}" > /dev/null
	pgrep "reqcount" | xargs -I {} kill -9 "{}" > /dev/null

request:
	curl localhost:3001