# RequestCount

`RequestCount` is a proof-of-concept application that
* counts the number of requests handled by a server instance;
* counts the total number of requests handled by all servers that are running.

The application serves these numbers through plain text HTTP response.

## Dependencies

* unix-like environment
	* NOTE: tested only on linux.
* go, 1.20+ recommended
* docker
* docker-compose
* make (optional: you could always copy the commands from `Makefile`)
* parallel (optional: if you have this installed, you might want to modify `Makefile`-s)

## Getting started

This proof-of-concept project has set up 3 ways to run the application:
* locally;
* using `docker`;
* using `docker-compose`.

### Locally

The `make` targets for locally running the application are located in `Makefile.local`.
This means that `-f Makefile.local` flag needs to be used when running `make`.
For example

```sh
make -f Makefile.local build
```

To start the application, run
```sh
make -f Makefile.local run
```
and to cleanly stop, run
```sh
make -f Makefile.local clean
```

Possible `make` targets are:

* `bin` - create `./bin` directory and compile the application binaries
* `run`- compiles and starts the application on `localhost:8083`
* `down` - stops the application
* `clean` - stops the application and deletes `./bin` directory 
* `toggle-debug` - toggles the debug flag, which enables/disables extra logs; logs can be found at `/tmp/instance.<host>.<port>.debug.log`

### Docker

The `make` targets for running the application with `docker` are located in `Makefile.docker`.
This means that `-f Makefile.docker` flag needs to be used when running `make`.
For example

```sh
make -f Makefile.docker build
```

Before starting the application for the first time, run
```sh
make -f Makefile.docker network
```
To start the application, run
```sh
make -f Makefile.docker run
```
and to cleanly stop, run
```sh
make -f Makefile.docker clean
make -f Makefile.docker clean-network
```

Possible `make` targets are:

* `network` - creates a docker network that enables application internal communication
* `build` - build `docker` images
* `run`- builds docker images and starts the application; application is accessible at `localhost:8083`
* `down` - stops and removes application related `docker` containers
* `restart` - restarts application; will consider new changes to application
* `clean-network` - removes network that was created with `network` target
* `clean` - stops the application and deletes created `docker` images

#### Debugging

To enable extra debug logs, you need to connect to one of the `instances` container by running
```sh
docker exec -it <container id/name> sh
```
and do the following steps:
* update the packages
```sh
apt update -y && apt upgrade -y
```
* install `netcat`
```sh
apt install -y netcat-openbsd
```
* toggle debug logs
```sh
echo "" | nc -U /tmp/instance.<host>.<port>.sock
```
* read logs at `/tmp/instance.<host>.<port>.debug.log`

## Docker-compose

The `make` targets for running the application with `docker-compose` are located in `Makefile.docker-compose`.
This means that `-f Makefile.docker-compose` flag needs to be used when running `make`.
For example

```sh
make -f Makefile.docker-compose build
```

To start the application, run
```sh
make -f Makefile.docker-compose run
```
and to cleanly stop, run
```sh
make -f Makefile.docker-compose clean
```

Possible `make` targets are:

* `build` - build `docker` images
* `up` - start built application; application is accessible at `localhost:8083`
* `run`- builds docker images and starts the application; application is accessible at `localhost:8083`
* `down` - stops and removes application related `docker` containers
* `clean` - stops the application and deletes created `docker` images

#### Debugging

To enable extra debug logs, you need to connect to one of the `instances` container by running
```sh
docker exec -it <container id/name> sh
```
and do the following steps:
* update the packages
```sh
apt update -y && apt upgrade -y
```
* install `netcat`
```sh
apt install -y netcat-openbsd
```
* toggle debug logs
```sh
echo "" | nc -U /tmp/instance.<host>.<port>.sock
```
* read logs at `/tmp/instance.<host>.<port>.debug.log`

## Things to imporve

* enable adding/removing/updating `instances` list in `./cmd/entry/main.go`
* make accessing `instances` list concurrently safe in `./cmd/entry/main.go`
* better handling of empty `instances` list in `./cmd/entry/main.go`

## Author

Meelis Utt