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
* create a local docker registry;
```sh
docker run -d -p 5000:5000 --restart=always --name registry registry:2
```
* build docker images with tag suitable for local registry
```sh
docker build -t localhost:5000/count_entry -f Dockerfile.entry .
docker build -t localhost:5000/count_instance -f Dockerfile.entry .
```
* push the image to the registry
```sh
docker push localhost:5000/count_entry
```
to see the registry, run
```sh
curl localhost:5000/v2/_catalog
```

--- TODO: below this line I still have some things to figure out - getting `(BadRequest): container "entry" in pod "entry" is waiting to start: trying and failing to pull image`

Next we need to create a manifest file.
For example:
```sh
printf "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: entry-deployment\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: entry\n  template:\n    metadata:\n      labels:\n        app: entry\n    spec:\n      containers:\n      - name: entry-container\n        image: localhost:5000/count_entry\n        ports:\n        - containerPort: 8083" > manifest.yaml
```

Lastly we deploy it with `kubectl`
```sh
kubectl create -f ./manifest.yaml
```

### Possible fixes/helpful commands

Solving the TODO mentioned above

```sh
eval $(minikube -p minikube docker-env)
```


## Author

Meelis Utt