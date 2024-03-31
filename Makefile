
bin:
	test -d bin/ || mkdir bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/entry ./cmd/entry/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/instance ./cmd/instance/main.go

# local
local-toggle-debug:
	echo "" | nc -U /tmp/instance..8081.sock
	echo "" | nc -U /tmp/instance..8082.sock
	echo "" | nc -U /tmp/instance..8083.sock

local-run: bin
	HOST="" PORT="8084" ./bin/instance &
	HOST="" PORT="8085" ./bin/instance &
	HOST="" PORT="8086" ./bin/instance &
	HOST="" PORT="8083" INSTANCES=":8085,:8084,:8086" ./bin/entry &

local-restart: local-down local-run

local-down:
	pgrep instance | parallel 'kill -9 {}'
	pgrep entry | parallel 'kill -9 {}'

local-clean: local-down
	rm -rf ./bin

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

run: build
	docker run -d -i -t -p 127.0.0.1:8083:8083 \
		--network=$(network-request-count) \
		-e HOST= \
		-e PORT=8083 \
		-e INSTANCES=:8084,:8085,:8086 \
		--name $(container-entry) \
		$(image-entry)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST= \
		-e PORT=8084 \
		--name request-count-host1 \
		$(image-instance)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST= \
		-e PORT=8085 \
		--name request-count-host2 \
		$(image-instance)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST= \
		-e PORT=8086 \
		--name request-count-host3 \
		$(image-instance)
		# --rm \

down:
	docker ps -aq | awk '{print $$1}' | parallel 'docker stop {} && docker rm {}'

restart: down run

clean: down
	docker images | grep $(image-instance) | awk '{print $$3}' | xargs -I {} docker image rm -f "{}"	
	docker images | grep $(image-entry) | awk '{print $$3}' | xargs -I {} docker image rm -f "{}"	
	docker network rm $(network-request-count)
