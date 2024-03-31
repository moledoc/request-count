
# local
local-toggle-debug:
	echo "" | nc -U /tmp/instance..8081.sock
	echo "" | nc -U /tmp/instance..8082.sock
	echo "" | nc -U /tmp/instance..8083.sock

local-run:
	HOST="" PORT="8081" RECV_HOST="" RECV_PORT="8181" SEND_HOST="" SEND_PORT="8182" go run instance.go &
	HOST="" PORT="8082" RECV_HOST="" RECV_PORT="8182" SEND_HOST="" SEND_PORT="8183" go run instance.go &
	HOST="" PORT="8083" RECV_HOST="" RECV_PORT="8183" SEND_HOST="" SEND_PORT="8181" go run instance.go &
	HOST="" PORT="3000" INSTANCES=":8082,:8081,:8083" go run entry.go &

local-restart: local-down local-run

local-down:
	pgrep instance | parallel 'kill -9 {}'
	pgrep entry | parallel 'kill -9 {}'

# docker
build:
	docker build -t round-robin .
	docker network create -d bridge request-count

# http://host1
run: build
	docker run -d -i -t -p 8081:8081 \
		-e HOST="http://round-robin-host1" \
		-e PORT="8081" \
		-e RECV_HOST="http://round-robin-host1" \
		-e RECV_PORT="8181" \
		-e SEND_HOST="http://round-robin-host2" \
		-e SEND_PORT="8182" \
		--name round-robin-host1 \
		round-robin
		# --rm \
	docker run -d -i -t -p 8082:8082 \
		-e HOST="http://round-robin-host2" \
		-e PORT="8082" \
		-e RECV_HOST="http://round-robin-host2" \
		-e RECV_PORT="8182" \
		-e SEND_HOST="http://round-robin-host3" \
		-e SEND_PORT="8183" \
		--name round-robin-host2 \
		round-robin
		# --rm \
	docker run -d -i -t -p 8083:8083 \
		-e HOST="http://round-robin-host3" \
		-e PORT="8083" \
		-e RECV_HOST="http://round-robin-host3" \
		-e RECV_PORT="8183" \
		-e SEND_HOST="http://round-robin-host1" \
		-e SEND_PORT="8181" \
		--name round-robin-host3 \
		round-robin
		# --rm \

down:
	docker ps -aq | awk '{print $$1}' | parallel 'docker stop {} && docker rm {}'


clean: down
	docker images | grep round-robin | awk '{print $$3}' | xargs -I {} docker image rm "{}"	
	docker network rm request-count
