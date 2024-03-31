# RequestCount

`RequestCount` is a simple cluster of web-servers that:
* count the number of requests handled by each server instance;
* count the number of requests handled by all the server instances.

## Getting started

Currently the easiest way to to start the servers is to run

```sh
make run
```

and to close them

```sh
make down
```

To interact with the servers, do

```sh
make request
```

or just do

```sh
curl localhost:8081
```

To get an understanding how the servers are currently set up, I invite you to investigate `Makefile`.

## TODO:

* Dockerize and verify current approach
	* iterate, when needed
* Improve readme
* write help

## Author

Meelis Utt