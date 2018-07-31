# Service Catalog Sample - Cloud SQL (MySQL)

This sample demonstrates how to build a simple Kubernetes web application
using [Kubernetes Service Catalog](
https://kubernetes.io/docs/concepts/service-catalog/)
and a [Google Cloud Platform Service Broker](
https://cloud.google.com/kubernetes-engine/docs/concepts/add-on/service-broker),
an implementation of the [Open Service Broker](
https://www.openservicebrokerapi.org/) standard.

The sample highlights a number of Kubernetes and Open Service Broker concepts:

*   Using Service Catalog and the Service Broker to *provision*
    a *service instance*.
*   *Binding* the provisioned *service instance* to a Kubernetes application.
*   Use of the *binding* by the application to access the *service instance*.

The sample application allows users to store and retrieve musician information.

At the center of the sample is a Cloud SQL (MySQL) instance, which will be
provisioned in your project by the Service Broker. A Kubernetes application
will access the MySQL instance using a *binding*.

## Objectives

To deploy and run the musician application, you must:

1.  Create a new namespace for all Kubernetes resources used by the sample.
2.  Provision Cloud SQL (MySQL) instance using Kubernetes Service Catalog.
3.  Deploy the musician application:
      1.  Provision a service account for the application.
      2.  Create a binding to the MySQL instance with the service account.
      3.  Deploy the Kubernetes application.
4.  Use the application's web API to create and read musician data.
5.  Deprovision and delete all resources used by the sample.


## Before you begin

Review the [information](../README.md) applicable to all Service Catalog
samples, including [prerequisites](../README.md#prerequisites):

*   A Kubernetes cluster, minimum version 1.8.x.
*   Kubernetes Service Catalog and the Service Broker [installed](
    https://cloud.google.com/kubernetes-engine/docs/how-to/add-on/service-broker/install-service-catalog).
*   The [Service Catalog CLI](
    https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#installing-the-service-catalog-cli)
    (`svcat`) installed.

## Step 1: Create a new Kubernetes namespace

```shell
kubectl create namespace cloud-mysql
```

## Step 2: Provision a Cloud SQL instance

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
instance.

Check on the status of the instance provisioning:

```shell
svcat describe instance --namespace cloud-mysql cloudsql-instance
```

The instance is provisioned when its status is `Ready`.

## Step 3: Deploy the musician application

The sample application assumes an identity of a [service account](
https://cloud.google.com/iam/docs/understanding-service-accounts) to access the
Cloud SQL instance. The service account will be granted the `client` role
(specified in [cloudsql-binding.yaml](manifests/cloudsql-binding.yaml)).
For more information on Cloud SQL's IAM roles and permissions, please refer to
Cloud SQL's [access control documentation](
https://cloud.google.com/sql/docs/mysql/project-access-control#permissions_and_roles).

Even though the application only uses Cloud SQL, a typical application may use
a number of different Google Cloud services. For example, the application can
be extended to store photos from concerts in Cloud Storage bucket. In that case,
the application will use a single service account to access all Google Cloud
resources. The musician sample application follows this pattern - it uses
a single service account created specifically for the application.

To deploy the musician application, you will:

*   Create a service account by provisioning a special 'service account'
    service.
*   Create a binding to the service account instance. This will:
      *    Create a service account private key.
      *    Store the private key in the Kubernetes secret as `privateKeyData`.
*   Create a binding to Cloud SQL instance, referencing the service account.
    This will:
      *    Grant the service account the roles needed to use the Cloud SQL
           instance.
      *    Store the MySQL connection information (`connectionName`)
           in a Kubernetes secret.

### Step 3.1: Provision a service account

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

Once the status is `Ready`, create a binding to make the service account
private key available in the `cloud-mysql` namespace as a secret:

```shell
svcat bind --namespace cloud-mysql service-account
```

Check on the status of the binding operation:

```shell
svcat get binding --namespace cloud-mysql service-account
```

Once the binding status is `Ready`, view the secret containing the service
account credentials:

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

## Step 3.2: Create a Binding to the Cloud SQL Instance

Create a binding to the `cloudsql-instance` using the parameters in
[cloudsql-binding.yaml](manifests/cloudsql-binding.yaml):

```shell
kubectl create -f ./manifests/cloudsql-binding.yaml
```

Check on the binding status:

```shell
svcat get binding --namespace cloud-mysql cloudsql-binding
```

Once the `cloudsql-binding` status is `Ready`, view the secret containing the
information for the application (Cloud SQL Proxy) to connect to the
Cloud SQL instance (the default name of the secret is the same as the name
of the binding resource: `cloudsql-binding`):

```shell
kubectl get secret --namespace cloud-mysql -o yaml cloudsql-credentials
```

Notice the `connectionName` value, which is referenced from the
[user-deployment.yaml](manifests/user-deployment.yaml) and tells the application
which MySQL instance to access.

## Step 3.3: Deploy the Application

Create the Kubernetes deployment using configuration in
[user-deployment.yaml](manifests/user-deployment.yaml). The configuration
uses two secrets to obtain the service account credentials (`service-account`
secret) and MySQL connection information (`cloudsql-binding` secret).

```shell
kubectl create -f ./manifests/user-deployment.yaml
```

Wait for the deployment to complete:

```shell
# Deployment
kubectl get deployment --namespace cloud-mysql musicians

# Load balancer service
kubectl get service --namespace cloud-mysql cloudsql-user-service
```

Next, find the external IP address of the Kubernetes load balancer:

```shell
kubectl get service --namespace cloud-mysql
IP= ... # External IP address of the Kubernetes load balancer.
```

or:

```shell
IP=$(kubectl get service --namespace cloud-mysql cloudsql-user-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Step 4: Access the Application

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

## Step 5: Cleanup

To avoid incurring charges to your Google Cloud Platform account for the
resources used in this sample, delete and deprovision all resources.

An expedient way is to delete the Kubernetes namespace; however make sure that
the namespace doesn't contain any resources you want to keep:

```shell
kubectl delete namespace cloud-mysql
```

Alternatively, delete all resources individually by running the following
commands:

**Note:** You may have to wait several minutes between steps to allow for the
previous operations to complete.

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
