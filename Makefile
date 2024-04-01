run:
	docker-compose -f ./docker-compose.yml -p "count" up -d --build

build:
	docker-compose -f ./docker-compose.yml -p "count" build

up:
	docker-compose -f ./docker-compose.yml -p "count" up -d

down:
	docker-compose -f ./docker-compose.yml -p "count" down

clean:
	docker-compose -f ./docker-compose.yml -p "count" down --rmi all -v --remove-orphans
