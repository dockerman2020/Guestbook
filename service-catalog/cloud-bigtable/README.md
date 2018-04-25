# Service Catalog Sample - Cloud Bigtable

This sample demonstrates how to build a simple Kubernetes web application
using [Kubernetes Service Catalog](
https://kubernetes.io/docs/concepts/service-catalog/)
and a [Google Cloud Platform Service Broker](
https://cloud.google.com/kubernetes-engine/docs/concepts/add-on/service-broker),
an implementation of the [Open Service Broker](
https://www.openservicebrokerapi.org/) standard.

The sample highlights a number of Kubernetes and Open Service Broker concepts:

*   Using Service Catalog and GCP Service Broker to *provision*
    a *service instance*.
*   *Binding* the provisioned *service instance* to a Kubernetes application.
*   Use of the *binding* by the application to access the *service instance*.

The sample is a simple web application which records page visits using
[Cloud Bigtable](https://cloud.google.com/bigtable/). The application's
web UI displays the page visit statistics.

At the core of the sample is a Bigtable instance, which is provisioned in your
project by the Service Broker. The Bigtable instance is accessed by two
Kubernetes applications: an [admin](admin/) job and a [counter](counter/)
deployment. The applications access the Bigtable instance using *bindings*.

The admin job uses a binding which will allow it to create a Bigtable table in
the Bigtable instance.

The counter deployment uses a binding which will only allow user-level
access to the Bigtable instance - to read from and write to the table in the
Cloud Bigtable table.

## Objectives

To deploy and run the sample application, you must:

1.  Create a new namespace for all Kubernetes resources used by the sample.
2.  Provision a Bigtable instance using Kubernetes Service Catalog.
3.  Administer the Bigtable instance using a Kubernetes job:
    1.  Create a binding for the admin job.
    2.  Deploy the admin job in your Kubernetes cluster.
4.  Deploy the web application:
    1.  Provision a service account for the web application.
    2.  Create a binding to the Bigtable instance with the web application
        service account.
    3.  Create the Kubernetes counter deployment.
5.  Interact with the application.
6.  Deprovision and delete all resources used by the sample.

## Before you begin

Review the [information](../README.md) applicable to all Service Catalog
samples, including [prerequisites](../README.md#prerequisites):

*   A Kubernetes cluster, minimum version 1.8.x.
*   Kubernetes Service Catalog and the Service Broker [installed](
    https://cloud.google.com/kubernetes-engine/docs/how-to/add-on/service-broker/install-service-catalog).
*   The Service Catalog CLI (`svcat`) [installed](
    https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#installing-the-service-catalog-cli).

## Step 1: Create a new Kubernetes namespace

```shell
kubectl create namespace bigtable
```

## Step 2: Provision Cloud Bigtable instance

To provision an instance of Cloud Bigtable, run:

```shell
kubectl create -f manifests/bigtable-instance.yaml
```

This command will use the Kubernetes Service Catalog to provision an empty
instance of Cloud Bigtable using parameters in
[bigtable-instance.yaml](./manifests/bigtable-instance.yaml).

Check on the status of the instance provisioning:

```shell
svcat get instance --namespace bigtable bigtable-instance
```

The instance is provisioned when its status is `Ready`.

## Step 3: Administer the Cloud Bigtable Instance

The admin job sets up the Bigtable instance by creating a table. To do so, it
requires administrator privileges granted on the Bigtable instance.

To express the intent of granting the administrator privileges to the admin job,
you will create a *binding* using the parameters in
[admin-bigtable-binding.yaml](manifests/admin-bigtable-binding.yaml). Creating
the binding will:

*   Create a service account for the admin job to authenticate with Cloud
    Bigtable.
*   Grant the service account the `roles/bigtable.admin` role.
*   Store the service account private key (`privateKeyData`) and the Bigtable
    instance connection information (`projectId`, `instanceId`) in a Kubernetes
    secret.

The admin job consumes the information stored in the secret via
environment variables `GOOGLE_APPLICATION_CREDENTIALS`, `GOOGLE_CLOUD_PROJECT`,
and `BIGTABLE_INSTANCE`. Review the admin job configuration in
[admin-job.yaml](manifests/admin-job.yaml).

### Step 3.1: Create Binding for the Admin Job

Create the admin binding to the Cloud Bigtable instance using the parameters in
[admin-bigtable-binding.yaml](manifests/admin-bigtable-binding.yaml):

```shell
kubectl create -f ./manifests/admin-bigtable-binding.yaml
```

The command will use the Kubernetes Service Catalog to create a binding to the
Bigtable instance provisioned earlier.

Check on the status of the binding operation:

```shell
svcat get binding --namespace bigtable admin-bigtable-binding
```

Once the binding status is `Ready`, view the Kubernetes secret containing the
result of the binding (the default name of the secret is the same as the name
of the binding resource - `admin-bigtable-binding`)

```shell
kubectl get secret --namespace bigtable admin-bigtable-binding -o yaml
```

Notice the values `privateKeyData`, `projectId`, and `instanceId` which contain
the result of the binding, ready for the admin job to use.

### Step 3.2: Create the Admin Job

The admin job is executed once to initialize the Cloud Bigtable instance.
It creates a table. Create the admin job using parameters
in [admin-job.yaml](manifests/admin-job.yaml).

```shell
kubectl create -f ./manifests/admin-job.yaml
```

Check on completion of the job:

```shell
kubectl get job --namespace bigtable bigtable-admin-job
```

The Bigtable instance is now ready to be used by the `counter` application.

## Step 4: Deploy the application

The `counter` deployment reads from and writes to the table. It
performs no administrative operations. Therefore, it only requires user level
access and will assume the identity of a service account with user level
privileges.

Even though the `counter` deployment only uses Cloud Bigtable, a typical
application may use a number of different Google Cloud services. For example,
the `counter` can be extended to store large objects in Cloud Storage bucket.
In this case, the application will use a single service account rather than
creating a new one for each binding.

The `counter` application follows this pattern. You will:

*   Create a service account instance by provisioning a special
    'service account' service.
*   Create a binding to the service account instance. This will:
    *   Create a service account private key.
    *   Store the private key in the Kubernetes secret as `privateKeyData`.
*   Create a binding to Cloud Bigtable instance, referencing the service
    account. This will:
    *   Grant the service account the roles needed to use the Cloud Bigtable
        instance.
    *   Store the Bigtable instance connection information (`projectId`,
        `instanceId`) in a Kubernetes secret.

The `counter` deployment consumes **both** secrets via environment variables
`GOOGLE_APPLICATION_CREDENTIALS`, `GOOGLE_CLOUD_PROJECT`, and
`BIGTABLE_INSTANCE`.
Review the counter deployment configuration in
[counter-deployment.yaml](manifests/counter-deployment.yaml).

### Step 4.1: Provision a User Service Account

Create a user service account as follows:

```shell
svcat provision user-account \
    --class cloud-iam-service-account \
    --plan beta \
    --namespace bigtable \
    --param accountId=bigtable-user
```

Check on the status of the service account provisioning:

```shell
svcat get instance --namespace bigtable user-account
```

Once the status is `Ready`, create a binding to make the service account
private key available in a secret.

```shell
svcat bind user-account --namespace bigtable
```

Check the binding status:

```shell
svcat get binding --namespace bigtable user-account
```

When the binding status is `Ready`, view the secret containing the service
account credentials:

```shell
kubectl get secret --namespace bigtable user-account -o yaml
```

Notice the `privateKeyData` value which contains the service account private
key.

### Step 4.2: Grant user service account access to Bigtable

Create the user binding to the Cloud Bigtable instance using the parameters in
[user-bigtable-binding.yaml](manifests/user-bigtable-binding.yaml):

```shell
kubectl create -f ./manifests/user-bigtable-binding.yaml
```

Check the binding status:

```shell
svcat get binding --namespace bigtable user-bigtable-binding
```

Once the `user-bigtable-binding` status is `Ready`, view the secret containing
the result of the binding (the default name of the secret is the same as the
name of the binding resource - `user-bigtable-binding`):

```shell
kubectl get secret --namespace bigtable user-bigtable-binding -o yaml
```

Notice the `projectId` and `instanceId` values. They are referenced from
[counter-deployment.yaml](manifests/counter-deployment.yaml) and tell
the counter deployment which Bigtable instance to access.

### Step 4.3: Create the Counter Deployment

Create the Kubernetes deployment using configuration in
[counter-deployment.yaml](manifests/counter-deployment.yaml). The
configuration uses two secrets:

| Contents                      | Secret name             | Field names               |
| ----------------------------- | ----------------------- | ------------------------- |
| Service account credentials   | `user-account`          | `privateKeyData`          |
| Bigtable instance information | `user-bigtable-binding` | `projectId`, `instanceId` |

```shell
kubectl create -f ./manifests/counter-deployment.yaml
```

Wait for the deployment to complete and then find the the Kubernetes service
external IP address:

```shell
kubectl get service --namespace bigtable bigtable-counter
IP= ... # External IP address of the Kubernetest load balancer.
```

or:

```shell
IP=$(kubectl get service bigtable-counter --namespace bigtable -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Step 5: Use the counter application

Use the IP address of the Kubernetes load balancer service along with a `curl`
command to access the application.

`GET /` will return the counter of one.

```shell
# Query the current counter:
curl http://${IP}/

<!doctype html><title>Bigtable example</title>
<h1>Bigtable example</h1>
<p><code>/</code> has been visited 1 times.
```

The second `GET /` will return the counter of two.

```shell
# Query the current counter:
curl http://${IP}/

<!doctype html><title>Bigtable example</title>
<h1>Bigtable example</h1>
<p><code>/</code> has been visited 2 times.
```

## Step 6: Cleanup

To avoid incurring charges to your Google Cloud Platform account for the
resources used in this sample, delete and deprovision all resources.

An expedient way is to delete the Kubernetes namespace; however make sure that
the namespace doesn't contain any resources you want to keep:

```shell
kubectl delete namespace bigtable
```

Alternatively, delete all resources individually by running the following
commands:

**Note:** You may have to wait several minutes between steps to allow for the
previous operations to complete.

Delete the application deployment and the load balancer service:

```shell
kubectl delete -f ./manifests/counter-deployment.yaml
```

Delete the admin Kubernetes job:

```shell
kubectl delete -f ./manifests/admin-job.yaml
```

Delete the user binding to the Cloud Bigtable instance:

```shell
kubectl delete -f ./manifests/user-bigtable-binding.yaml
```

Delete the admin binding to the Cloud Bigtable instance. This also deletes
the service account created for the admin binding.

```shell
kubectl delete -f ./manifests/admin-bigtable-binding.yaml
```

Unbind the user service account:

```shell
svcat unbind user-account --namespace bigtable
```

Deprovision the user service account:

```shell
svcat deprovision user-account --namespace bigtable
```

Deprovision the Cloud Bigtable instance:

```shell
kubectl delete -f ./manifests/bigtable-instance.yaml
```

If the `bigtable` namespace contains no resource you wish to keep,
delete it:

```shell
kubectl delete namespace bigtable
```

Remove all the roles associated with the service accounts `bigtable-admin` and
`bigtable-user` following [this
guide](https://cloud.google.com/iam/docs/granting-changing-revoking-access#revoking_access_to_team_members).

## Troubleshooting

Please find the troubleshooting information
[here](../README.md#troubleshooting).
