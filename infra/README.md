# Infra

## Minikube

You will need to start minikube with at least 4GB of memory

```shell
minikube start --memory 4092
```
### Setting up namespaces

```
kubectl apply -f 000-namespaces.yaml
```

### Setting up the db

[Reference](https://postgres-operator.readthedocs.io/en/latest/quickstart/)

1. Run db.yaml

	```
	kubectl apply -f db.yaml
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

3. Once the cluster is created, you can try connecting to it:

	```
	export HOST_PORT=$(minikube -n db service acid-minimal-cluster --url | sed 's,.*/,,')
	export PGHOST=$(echo $HOST_PORT | cut -d: -f 1)
	export PGPORT=$(echo $HOST_PORT | cut -d: -f 2)
	export PGPASSWORD=$(kubectl -n db get secret postgres.acid-minimal-cluster.credentials.postgresql.acid.zalan.do -o 'jsonpath={.data.password}' | base64 -d)
	export PGSSLMODE=require
	psql -U postgres
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