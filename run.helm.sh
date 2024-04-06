#!/bin/sh

set -x

env() {
	minikube start --insecure-registry "10.0.0.0/24"
	minikube addons enable registry
	MINIKUBE_IP=$(minikube ip)
	echo ${MINIKUBE_IP}
	docker run --rm -it -d --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:${MINIKUBE_IP}:5000"
	docker build -t localhost:5000/count_entry -f Dockerfile.entry .
	docker build -t localhost:5000/count_instance -f Dockerfile.instance .
	sleep 45
	docker push localhost:5000/count_entry
	docker push localhost:5000/count_instance
	curl localhost:5000/v2/_catalog
}

dry() {
	helm install --debug --dry-run request-count ./request-count
}

up() {
	helm install request-count ./request-count
	sleep 15
	NODE_PORT="$(kubectl get services/entry -o go-template='{{(index .spec.ports 0).nodePort}}')"
	echo "NODE_PORT=$NODE_PORT"
	curl $(minikube ip):${NODE_PORT}
}

down() {
	helm uninstall request-count
}

clean() {
	down
	minikube delete --all
	docker ps -aq | xargs -I {} docker stop "{}"
}

for action in $@; do
	case ${action} in
		env)
			env
			;;
		up)
			up
			;;
		down)
			down
			;;
		restart)
			down
			up
			;;
		clean)
			clean
			;;
	esac
done