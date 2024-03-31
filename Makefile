
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

image-instance = request-count-instance
image-entry = request-count-entry
network-request-count = request-count
container-entry = request-count

init:
	docker network create -d bridge $(network-request-count)

build:
	docker build -t $(image-instance) -f Dockerfile.instance .
	docker build -t $(image-entry) -f Dockerfile.entry .

# http://host1
run: build
	docker run -d -i -t -p 127.0.0.1:8083:8081 \
		--network=$(network-request-count) \
		-e HOST="" \
		-e PORT="8081" \
		-e INSTANCES=":8082,:8084,:8083" \
		--name $(container-entry) \
		$(image-entry)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST="" \
		-e PORT="8082" \
		-e RECV_HOST="" \
		-e RECV_PORT="8182" \
		-e SEND_HOST="" \
		-e SEND_PORT="8183" \
		--name request-count-host1 \
		$(image-instance)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST="" \
		-e PORT="8083" \
		-e RECV_HOST="" \
		-e RECV_PORT="8183" \
		-e SEND_HOST="" \
		-e SEND_PORT="8184" \
		--name request-count-host2 \
		$(image-instance)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST="" \
		-e PORT="8084" \
		-e RECV_HOST="" \
		-e RECV_PORT="8184" \
		-e SEND_HOST="" \
		-e SEND_PORT="8182" \
		--name request-count-host3 \
		$(image-instance)
		# --rm \

down:
	docker ps -aq | awk '{print $$1}' | parallel 'docker stop {} && docker rm {}'

restart: down run

clean: down
	docker images | grep $(image-instance) | awk '{print $$3}' | xargs -I {} docker image rm "{}"	
	docker images | grep $(image-entry) | awk '{print $$3}' | xargs -I {} docker image rm "{}"	
	docker network rm $(network-request-count)
