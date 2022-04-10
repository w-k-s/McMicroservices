# Infra

## Environment

You can deploy this in a local Kubernetes cluster using [Minikube](#minikube).
You can deploy this project on EKS using the [eksctl](#eksctl) tool

## eksctl

To install eksctl, follow the instructions [here](sudo mv /tmp/eksctl /usr/local/bin)

To create the kubernetes cluster, run:

```
eksctl create cluster --config-file deployment/eksctl.yaml
```

You can update the `eksctl.yaml` file and update the cluster using:

```
eksctl upgrade cluster --config-file deployment/eksctl.yaml
```



## Minikube

You will need to start minikube with at least 4GB of memory

```shell
minikube start --memory 4092
```

## Deployment Steps

### Setting up namespaces

```
kubectl apply -f 000-namespaces.yaml
```

### Setting up the db

[Reference](https://postgres-operator.readthedocs.io/en/latest/quickstart/)

1. Run db.yaml

	```
	kubectl apply -f 020-db.yaml
	```

	This should print:

	```
	namespace/db created
	configmap/postgres-operator created
	serviceaccount/postgres-operator created
	clusterrole.rbac.authorization.k8s.io/postgres-operator created
	clusterrolebinding.rbac.authorization.k8s.io/postgres-operator created
	clusterrole.rbac.authorization.k8s.io/postgres-pod created
	deployment.apps/postgres-operator created
	service/postgres-operator created
	postgresql.acid.zalan.do/acid-minimal-cluster created
	```

2. Check that the status of the postgres cluster by running the following command:

	```
	 kubectl -n db get postgresql -w
	```

	This should print:

	```
	NAME                   TEAM   VERSION   PODS   VOLUME   CPU-REQUEST   MEMORY-REQUEST   AGE   STATUS
	acid-minimal-cluster   acid   14        1      1Gi                                     45s   Creating
	```

	To see the pods, you can run:

	```
	kubectl -n db get pod -l name=postgres-operator
	```

3. To get the password for user `postgres`:

	```
	kubectl -n db get secret postgres.acid-minimal-cluster.credentials.postgresql.acid.zalan.do -o 'jsonpath={.data.password}' | base64 -d
	```

	To get the password for user `root`:
	```
	kubectl -n db get secret root.acid-minimal-cluster.credentials.postgresql.acid.zalan.do -o 'jsonpath={.data.password}' | base64 -d
	```

	

4. If the cluster fails, you can get more info using:

	```
	kubectl -n db describe postgresql
	```

5. Delete the cluster:

	```
	kubectl -n db delete postgresql acid-minimal-cluster
	```
### Setting up Kafka

1. Install Strimzi

	```
	kubectl apply -f kafka/000-namespaces.yaml
	kubectl -n app apply -f kafka/020-RoleBinding-strimzi-cluster-operator.yaml
	kubectl -n app apply -f kafka/031-RoleBinding-strimzi-cluster-operator-entity-operator-delegation.yaml 
	kubectl -n kafka apply -f kafka/
	```

2. Create a Cluster

	```
	kubectl -n app apply -f 020-kafka.yaml
	```

3. Wait for the cluster to be ready

	```
	kubectl -n app wait kafka/kafka-cluster --for=condition=Ready --timeout=300s
	```
	Note: `kafka/kafka-cluster` is the `Kind` and `metadata.name` defined in `020-kafka.yml` 

### Deploying the services

**Config Service**

```
kubectl -n app apply -f 030-common-config.yaml
kubectl -n app apply -f 040-config-service.yaml
```

You can verify the config service is running by running the following commands in terminal:

```
kubectl -b app port-forward svc/config-service 8888:80
curl http://localhost:8888/actuator/health
```

This will print:
```json
{"status":"UP","groups":["liveness","readiness"]}
```

You can verify that the server is providing configurations using:
```
curl http://localhost:8888/order-service/default 
```