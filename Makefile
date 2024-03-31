# TODO: cleanup and naming

# local
bin:
	test -d bin/ || mkdir bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/entry ./cmd/entry/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/instance ./cmd/instance/main.go

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
image-instance = count_instance
image-entry = count_entry
network-request-count = request-count
container-entry = count_entry
container-instance = count_instance

docker-network-init:
	docker network create -d bridge $(network-request-count)

docker-build:
	docker build -t $(image-instance) -f Dockerfile.instance .
	docker build -t $(image-entry) -f Dockerfile.entry .

docker-run: docker-build
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
		--name $(container-instance)1 \
		$(image-instance)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST= \
		-e PORT=8085 \
		--name $(container-instance)2 \
		$(image-instance)
		# --rm \
	docker run -d -i -t \
		--network container:$(container-entry) \
		-e HOST= \
		-e PORT=8086 \
		--name $(container-instance)3 \
		$(image-instance)
		# --rm \

docker-down:
	docker ps -aq | awk '{print $$1}' | parallel 'docker stop {} && docker rm {}'

docker-restart: down run

docker-network-clean:
	docker network rm $(network-request-count)

docker-clean: docker-down
	docker images | grep $(image-instance) | awk '{print $$3}' | xargs -I {} docker image rm -f "{}"	
	docker images | grep $(image-entry) | awk '{print $$3}' | xargs -I {} docker image rm -f "{}"	
	docker images | grep none | awk '{print $$3}' | xargs -I {} docker image rm -f "{}"	

# docker-compose
build:
	docker-compose -f ./docker-compose.yml -p "count" build

up:
	docker-compose -f ./docker-compose.yml -p "count" up -d

down:
	docker-compose -f ./docker-compose.yml -p "count" down

run:
	docker-compose -f ./docker-compose.yml -p "count" up -d --build

clean:
	docker-compose -f ./docker-compose.yml -p "count" down --rmi all -v --remove-orphans
