# DevOpsing

## Plan

1. install local Kubernetes distribution (k3s, microk8s, minikube)
2. run Helm's `hello-world` project
3. create k8s manifest yamls for `request-count` application
4. create Helm 3 charts for `request-count` application

## Step 1

Chose minikube as local k8s distribution, since it seemed the most straightforward for me to set up.
Following [minicube start](https://minikube.sigs.k8s.io/docs/start).

```sh
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_amd64.deb
doas dpkg -i minikube_latest_amd64.deb
minikube start
kubectl get po -A
```

Running through the example `minikube` provides is really insightful.

```sh
# (optional: to see the dashboard) minikube dashboard
kubectl create deployment hello-minikube --image=kicbase/echo-server:1.0
kubectl expose deployment hello-minikube --type=NodePort --port=8080
kubectl get services hello-minikube

kubectl port-forward service/hello-minikube 7080:8080
# or
minikube service hello-minikube
```

To pause/clean up the example:
```sh
minikube pause
minikube delete --all
```

## Step 2

Installing Helm

```sh
git clone https://github.com/helm/helm.git
cd helm
make
export PATH="$PATH:$(pwd)/bin"
```

Going through an [example](https://helm.sh/docs/chart_template_guide/getting_started/#a-starter-chart).

Create and deploy the chart and run application
```sh
helm create mychart
rm -rf mychart/templates/*
printf "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: mychart-configmap\ndata:\n  myvalue: \"Hello World\"" > ./mychart/templates/configmap.yaml
helm install full-coral ./mychart
helm get manifest full-coral
kubectl create deployment full-coral --image=kicbase/echo-server:1.0
kubectl expose deployment full-coral --type=ClusterIP --port=8080
kubectl get services full-coral
kubectl port-forward service/full-coral 8081:8080
```

Send a request
```sh
curl localhost:8081
# Request served by full-coral-58597fd848-xvd8g
# 
# HTTP/1.1 GET /
# 
# Host: localhost:8081
# Accept: */*
# User-Agent: curl/7.88.1
```

And to clean up
```sh
helm uninstall full-coral
kubectl delete service/full-coral
```

## Step 3

Aim is to deploy local docker image to kubernetes.
For that we need to 
* start minikube in a way that we can access docker images from local registry;
```sh
# https://minikube.sigs.k8s.io/docs/handbook/registry/
minikube start --insecure-registry "10.0.0.0/24"
minikube addons enable registry
docker run --rm -it -d --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"
```
* build docker images with tag suitable for local registry
```sh
docker build -t localhost:5000/count_entry -f Dockerfile.entry .
docker build -t localhost:5000/count_instance -f Dockerfile.instance .
```
* push the image to the registry
```sh
docker push localhost:5000/count_entry
docker push localhost:5000/count_instance
```
to see the registry
```sh
curl localhost:5000/v2/_catalog
```

Now we need to create and expose the application with `kubectl`
```sh
kubectl apply -f ./devopsing/instance.yaml
kubectl expose deployment instance --protocol=TCP

INSTANCE_IP=$(kubectl get service/instance -o jsonpath='{.spec.clusterIP}')
sed "s/INSTANCE_IP/${INSTANCE_IP}/" ./devopsing/entry.yaml | kubectl apply -f -
kubectl expose deployment entry --type=NodePort --port=8083 --target-port=8083 --protocol=TCP
```

and then we can send a request against deployed application
```sh
export NODE_PORT="$(kubectl get services/entry -o go-template='{{(index .spec.ports 0).nodePort}}')"
echo "NODE_PORT=$NODE_PORT"
curl $(minikube ip):${NODE_PORT}
```

Also, some helpful commands
```sh
kubectl exec -it <pod name> -- <cmd ran in the pod (can be `sh`)>
```

With this we are able to deploy our application.
But we should be able to use better manifest files to reduce number of commands ran.
So here is an example with less commands:

```sh
kubectl apply -f ./devopsing/instance_v2.yaml
INSTANCE_IP=$(kubectl get service/instance-v2 -o jsonpath='{.spec.clusterIP}')
sed "s/INSTANCE_IP/${INSTANCE_IP}/" ./devopsing/entry_v2.yaml | kubectl apply -f -
curl $(minikube ip):30003
```

## Author

Meelis Utt