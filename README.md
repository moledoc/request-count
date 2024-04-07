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
* netcat (optional: needed to toggle extra logs for debugging)
* minikube
* kubectl
* helm

## Getting started

This proof-of-concept project has set up 5 ways to run the application:
* locally;
* using `docker`;
* using `docker-compose`;
* using `kubectl`;
* using `helm 3 charts`.

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

### Docker-compose

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

### Kubectl

Due to some annoyances with `make` and `Makefile`, separate run script for deploying application with `kubectl` was created.
Also to note, in this project `minikube` was used to setup local Kubernetes cluster.
When alternative to `minikube` is desired, then the alternative needs to be able to:
* pull docker images from local docker registry;
* get local kubernetes cluster's ip.

Before starting the application with `kubectl`, run
```sh
./run.kubectl.sh env
```
This will start `minikube`, build and push docker images to local docker registry.

To start application, run
```sh
./run.kubectl.sh up
```
and to cleanly stop, run
```sh
./run.kubectl.sh clean
```

Possible script targets are:

* `env` - start `minikube`, build and push docker images to local docker registry
* `up` - deploy and expose application; application is accessible at `localhost:8083`
* `down` - stops and deletes application related kubernetes pods/deployments/services
* `restart` - runs targets `down` and `up`
* `clean` - stops the application, local kubernetes cluster and local docker registry

### Helm 3 charts

Due to some annoyances with `make` and `Makefile`, separate run script for deploying application with `helm 3 charts` was created.
Also to note, in this project `minikube` was used to setup local Kubernetes cluster.
When alternative to `minikube` is desired, then the alternative needs to be able to:
* pull docker images from local docker registry;
* get local kubernetes cluster's ip.

Before starting the application with `helm 3 charts`, run
```sh
./run.helm.sh env
```
This will start `minikube`, build and push docker images to local docker registry.

To start application, run
```sh
./run.helm.sh up
```
and to cleanly stop, run
```sh
./run.helm.sh clean
```

Possible script targets are:

* `env` - start `minikube`, build and push docker images to local docker registry
* `dry` - dry-runs `helm install`, showing the manifests that would be used in the deployment
* `up` - deploy and expose application; application is accessible at `localhost:8083`
* `down` - stops and deletes application related kubernetes pods/deployments/services
* `restart` - runs targets `down` and `up`
* `clean` - stops the application, local kubernetes cluster and local docker registry


### Debugging

To toggle extra debug logs, you need to send a TCP request to socket `/tmp/instance.<host>.<port>.sock`.
When application is started using the steps from sections 

* [Docker](#Docker) or [Docker-compose](#Docker-compose), then connect to one of the `instances` container by running
```sh
docker exec -it <container id/name> sh
```
* [Kubectl](#Kubectl) or [Helm 3 charts](#Helm-3-charts), then connect to one of the `instances` pods by running
```sh
kubectl exec -it <instance pod name> -- sh
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