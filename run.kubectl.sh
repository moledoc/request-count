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

up() {
	kubectl apply -f ./devopsing/instance_v2.yaml
	INSTANCE_IP=$(kubectl get service/instance -o jsonpath='{.spec.clusterIP}')
	sed "s/INSTANCE_IP/${INSTANCE_IP}/g" ./devopsing/entry_v2.yaml | kubectl apply -f -
	NODE_PORT=$(kubectl get services/entry -o go-template='{{(index .spec.ports 0).nodePort}}')
	MINIKUBE_IP=$(minikube ip)
	echo "send request with 'curl ${MINIKUBE_IP}:${NODE_PORT}'"
	sleep 25
	curl ${MINIKUBE_IP}:${NODE_PORT}
}

down() {
	kubectl get pods | grep "entry\|instance" | awk '{print $1}' | xargs -I {} kubectl delete pods/"{}"
	kubectl get deployments | grep "entry\|instance" | awk '{print $1}' | xargs -I {} kubectl delete deployments/"{}"
	kubectl get services | grep "entry\|instance" | awk '{print $1}' | xargs -I {} kubectl delete services/"{}"
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