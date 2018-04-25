# Service Catalog Sample - Cloud SQL (MySQL)

This sample demonstrates how to build a simple Kubernetes web application
using [Kubernetes Service Catalog](
https://kubernetes.io/docs/concepts/service-catalog/)
and a [Google Cloud Platform Service Broker](
https://cloud.google.com/kubernetes-engine/docs/concepts/add-on/service-broker),
an implementation of the [Open Service Broker](
https://www.openservicebrokerapi.org/) standard.
The sample web application allows users to store and retrieve very simple
musician data.

The sample highlights a number of Kubernetes and Open Service Broker concepts:

*   Using Service Catalog and the Service Broker to *provision*
    a *service instance*.
*   *Binding* the provisioned *service instance* to a Kubernetes application.
*   Use of the *binding* by the application to access the *service instance*.


## Objectives

To deploy and run the simple musician application, you must:

1.  Create a new namespace for all Kubernetes resources used by the sample.
2.  Provision Cloud SQL (MySQL) instance using Kubernetes Service Catalog.
3.  Provision a service account for the application to access cloud resources.
4.  Create bindings to the Cloud SQL instance and service account.
5.  Deploy a Kubernetes application (configured to use the bindings).
6.  Use the application's web API to create and read musician data.


## Before you begin

Review the [information](../README.md) applicable to all Service Catalog samples.

Successful use of this sample requires:

*   A Kubernetes cluster, minimum version 1.8.x.
*   Kubernetes Service Catalog and the Service Broker [installed](
    https://cloud.google.com/kubernetes-engine/docs/how-to/add-on/service-broker/install-service-catalog).
*   The [Service Catalog CLI](
    https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#installing-the-service-catalog-cli)
    (`svcat`) installed.
*   [Cloud SQL Administration API](https://console.cloud.google.com/apis/library/sqladmin.googleapis.com)
    enabled.


## Create a new namespace in Kubernetes cluster

Create a new namespace for all Kubernetes resources used by the sample:

```shell
kubectl create namespace cloud-mysql
```

## Provision a Cloud SQL instance

To provision an instance of Cloud SQL (MySQL), run:

```shell
svcat provision cloudsql-instance \
    --namespace cloud-mysql \
    --class cloud-sql-mysql \
    --plan beta \
    --params-json '{
        "instanceId": "musicians-'${RANDOM}'",
        "databaseVersion": "MYSQL_5_7",
        "settings": {
            "tier": "db-n1-standard-1"
        }
    }'
```

This command will use the Kubernetes Service Catalog to provision a Cloud SQL
instance. To check on the status of instance provisioning, run:

```shell
svcat describe instance --namespace cloud-mysql cloudsql-instance
```

The provisioning of a Cloud SQL instance may take several minutes. Once the
provisioning is complete (its status is `Ready`), proceed to the next step.


## Provision a Service Account

The sample application assumes an identity of a [service account](
https://cloud.google.com/iam/docs/understanding-service-accounts) to access the
Cloud SQL instance. In this section you will provision a service account for
the application to use and make the service account credentials (private key)
available to the application by creating a binding.

To provision a service account using the Kubernetes Service Catalog, run:

```shell
svcat provision service-account \
    --namespace cloud-mysql \
    --class cloud-iam-service-account \
    --plan beta \
    --param accountId=cloudsql-user-service-account \
    --param displayName="A service account for Cloud SQL sample"
```

Check on the status of the service account provisioning:

```shell
svcat get instance --namespace cloud-mysql service-account
```

Once the service account is provisioned (its status is `Ready`), create
a binding for the application. The act of creating the binding will make
the service account credentials available in the `cloud-mysql` namespace
as a secret:

```shell
svcat bind --namespace cloud-mysql service-account
```

Check on the status of the binding operation:

```shell
svcat bind --namespace cloud-mysql service-account
```

Once the binding is completed (its status is `Ready`), you can view the secret
containing the service account credentials:

```shell
kubectl get secret --namespace cloud-mysql service-account -oyaml
```

Notice the `serviceAccount` key which contains base-64 encoded service account
email address and the `privateKeyData` key containing base-64 encoded private
key of the service account. The [user-deployment.yaml](
manifests/user-deployment.yaml) manifest references the `privateKeyData`
to make the service account private key available to the Cloud SQL proxy for
authentication with Cloud SQL and enables the web application (specifically
the Cloud SQL Proxy) to assume the identity of the service account.

## Create a Binding to the Cloud SQL Instance

The application, executing with the assumed identity of the service account,
needs to be permitted to access the Cloud SQL instance. To express the intent
of giving the application permission to access the Cloud SQL instance, create
a binding. The act of creating the binding will grant the service account
used by the application appropriate permissions to access the Cloud SQL
instance and will make the information necessary to access the Cloud SQL
instance available to the application in the form of Kubernetes secret.

The list of roles which should be granted to the service account is specified when
creating the binding ([cloudsql-binding.yaml](manifests/cloudsql-binding.yaml)).
Because the web application uses Cloud SQL Proxy, the only applicable role is
`client`.

For more information on Cloud SQL's IAM roles and permissions, please refer to
Cloud SQL's [access control documentation](
https://cloud.google.com/sql/docs/mysql/project-access-control#permissions_and_roles).

Create a binding to the `cloudsql-instance`:

```shell
kubectl create -f ./manifests/cloudsql-binding.yaml
```

Check on the status of the binding operation:

```shell
svcat get binding --namespace cloud-mysql cloudsql-binding
```

Once the binding is completed (its status is `Ready`), you can view the secret
containing information for the application (Cloud SQL Proxy) to connect to the
Cloud SQL instance:

```shell
kubectl get secret --namespace cloud-mysql -o yaml \
  "$(kubectl get servicebinding --namespace cloud-mysql cloudsql-binding -o=jsonpath='{.spec.secretName}')"
```

Notice the key `connectionName` which contains base-64 encoded value for the
proxy to use when connecting to the Cloud SQL instance. You can find the
reference to the `connectionName` in the [user-deployment.yaml](
manifests/user-deployment.yaml) manifest file.

## Deploy the Web Application

Having satisfied both dependencies of the web applications:
*   a service account
*   a binding to the Cloud SQL instance

You are now ready to deploy the web application:

```shell
kubectl create -f ./manifests/user-deployment.yaml
```

The web application consists of a Kubernetes deployment and a load balancer service.
The deployment contains a single pod with two containers: the Cloud SQL Proxy
and the web application frontend.
The load balancer forwards traffic to the web application frontend.

You can monitor the status of the deployment and the load balancer:

```shell
# Deployment
kubectl get deployment --namespace cloud-mysql musicians

# Load balancer service
kubectl get service --namespace cloud-mysql cloudsql-user-service
```

After the deployment finishes, find the the Kubernetes service external IP
address:

```shell
kubectl get service --namespace cloud-mysql
IP= ... # External IP address of the Kubernetes load balancer.
```

or:

```shell
IP=$(kubectl get service --namespace cloud-mysql cloudsql-user-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Access the Application

Use the IP address of the Kubernetes load balancer service along with a `curl`
command to access the application:

```shell
# Create (or reset) musicians table:
curl -X POST http://${IP}/reset

# Query the musician information (will return empty result):
curl http://${IP}/musicians

# Create new musicians:
curl http://${IP}/musicians -d '{"name": "John", "instrument": "Guitar"}'
curl http://${IP}/musicians -d '{"name": "Paul", "instrument": "Bass Guitar"}'
curl http://${IP}/musicians -d '{"name": "Ringo", "instrument": "Drums"}'
curl http://${IP}/musicians -d '{"name": "George", "instrument": "Lead Guitar"}'

# Query the musicians again:
curl http://${IP}/musicians
```

## Cleanup

To avoid incurring charges to your Google Cloud Platform account for the
resources used in this sample, run the following commands.

**Note:** You may have to wait several minutes between steps to allow for the
previous operations to complete. An expedient alternative to de-provisioning
all resources individually is to delete the Kubernetes namespace:
`kubectl delete namespace cloud-mysql`.

Delete the application deployment and the load balancer service:

```shell
kubectl delete -f ./manifests/user-deployment.yaml
```

Unbind the Cloud SQL instance from the application:

```shell
svcat unbind --namespace cloud-mysql cloudsql-instance --name cloudsql-binding
```

Unbind the service account:

```shell
svcat unbind --namespace cloud-mysql service-account
```

Deprovision the service account:

```shell
svcat deprovision --namespace cloud-mysql service-account
```

Deprovision the Cloud SQL instance:

```shell
svcat deprovision --namespace cloud-mysql cloudsql-instance
```

Delete the `cloud-mysql` Kubernetes namespace you used for the sample:

```shell
kubectl delete namespace cloud-mysql
```

Remove all the roles associated with the service account
`cloudsql-user-service-account` following [this
guide](https://cloud.google.com/iam/docs/granting-changing-revoking-access#revoking_access_to_team_members).

## Troubleshooting

Please find the troubleshooting information
[here](../README.md#troubleshooting).
