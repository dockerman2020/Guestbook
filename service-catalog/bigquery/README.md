# Service Catalog Sample - BigQuery

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

The sample application allows users to find Git commits whose message matches
a given prefix in the [public GitHub dataset](
https://cloud.google.com/bigquery/public-data/github) using
[BigQuery](https://cloud.google.com/bigquery/).

At the core of the sample is a BigQuery dataset, which is provisioned in your
project by the Service Broker. The BigQuery dataset is accessed by
three Kubernetes applications: an admin job, a web deployment for querying GitHub
data, and a cleanup job. The applications access the dataset using *bindings*.

The admin job uses a binding which will allow it to copy data from the
[public GitHub dataset](https://cloud.google.com/bigquery/public-data/github)
into a new table in the BigQuery dataset.

The GitHub data web deployment uses a binding which will only allow user-level
access to the BigQuery dataset - to create jobs to query the dataset and view
the contents of the dataset.

Finally, the cleanup job reuses the admin binding, allowing the job to delete
the table from the dataset, preparing the BigQuery dataset for deprovisioning.

## Objectives

To deploy and run the sample application, you must:

1.  Create a new namespace for all Kubernetes resources used by the sample.
2.  Provision a BigQuery dataset using Kubernetes Service Catalog.
3.  Administer the BigQuery dataset using a Kubernetes job:
      1.  Create a binding to the BigQuery dataset for the admin
          and cleanup jobs.
      2.  Deploy the admin job in your Kubernetes cluster; the admin job creates
          and populates the BigQuery dataset with data.
4.  Deploy the GitHub data application.
      1.  Provision a service account for the application.
      2.  Create a binding to the BigQuery dataset with the GitHub data application
          service account.
      3.  Create the Kubernetes GitHub data application deployment.
5.  Interact with the GitHub data application via a web API.
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
kubectl create namespace bigquery
```

## Step 2: Provision BigQuery dataset

To provision an instance of a BigQuery dataset, run:

```shell
kubectl create -f ./manifests/bigquery-instance.yaml
```

This command will use the Kubernetes Service Catalog to provision an empty
BigQuery dataset using parameters in
[bigquery-instance.yaml](./manifests/bigquery-instance.yaml).

Check on the status of the BigQuery instance provisioning:

```shell
svcat get instance --namespace bigquery bigquery-instance
```

The instance is provisioned when its status is `Ready`.

## Step 3: Administer the BigQuery dataset

The admin job sets up the BigQuery dataset by creating a table and copying data
from the [public GitHub dataset](https://cloud.google.com/bigquery/public-data/github)
into it.

The cleanup job cleans up at the end of the sample walkthrough
by deleting the table from the dataset, allowing for deprovisioning of the
BigQuery dataset.

Both of these jobs require administrator privileges granted at the project level,
so they will use the same service account.

To express the intent of granting the administrator privileges to an admin service account,
you will create a *binding* using the parameters in
[admin-bigquery-binding.yaml](./manifests/admin-bigquery-binding.yaml). Creating the binding will:

*   Create a service account for the admin and cleanup jobs to authenticate with BigQuery.
*   Grant the service account the `roles/bigquery.admin` role.
*   Store the service account private key (`privateKeyData`) and the BigQuery
    dataset connection information (`projectId`, `datasetId`) in a Kubernetes
    secret.

The admin and cleanup jobs consume the information stored in the secret via
environment variables `GOOGLE_APPLICATION_CREDENTIALS`, `BIGQUERY_PROJECT`, and
`BIGQUERY_DATASET`. Review the job configurations in
[admin-job.yaml](./manifests/admin-job.yaml) and
[cleanup-job.yaml](./manifests/cleanup-job.yaml).

### Step 3.1: Create Binding for the Admin and Cleanup Jobs

Create the admin binding to the BigQuery dataset using the parameters in
[admin-bigquery-binding.yaml](./manifests/admin-bigquery-binding.yaml):

```shell
kubectl create -f ./manifests/admin-bigquery-binding.yaml
```

The command will use the Kubernetes Service Catalog to create a binding to the
BigQuery instance provisioned earlier.

Check on the status of the binding operation:

```shell
svcat get binding -n bigquery admin-bigquery-binding
```

Once the binding status is `Ready`, view the Kubernetes secret containing the
result of the binding (the default name of the secret is the same as the name
of the binding resource - `admin-bigquery-binding`).

```shell
kubectl get secret -n bigquery admin-bigquery-binding -oyaml
```

Notice the values `privateKeyData`, `projectId`, and `datasetId` which contain
the result of the binding, ready for the admin and cleanup jobs to use.

### Step 3.2: Create the Admin Job

The admin job is executed once to initialize the data in the BigQuery instance.
It copies the public GitHub data into a table in the dataset, creating the table
if needed. Create the admin job using parameters in
[admin-job.yaml](./manifests/admin-job.yaml).

```shell
kubectl create -f ./manifests/admin-job.yaml
```

Check on completion of the job:

```shell
kubectl get job -n bigquery bigquery-admin-job
```

You can examine the BigQuery dataset and the newly created table in the
[BigQuery console](https://bigquery.cloud.google.com/).
**Note:** If you don't see any tables, you must explicitly grant yourself
the "BigQuery Data Viewer" role in your project.

The BigQuery instance is now ready to be used by the GitHub data application.

## Step 4: Deploy the Application

The GitHub data application serves user requests and queries the dataset. It
performs no administrative operations. Therefore, it only requires user-level
access and will assume the identity of a service account with user level
privileges.

Even though the GitHub data deployment only uses BigQuery, a typical
application may use a number of different Google Cloud services. For example,
the application can be extended to store values at certain times in a Cloud Spanner
instance. In this case, the application will use a single service account rather than
creating a new one for each binding.

The application follows this pattern. You will:

*   Create a service account instance by provisioning a special
    'service account' service.
*   Create a binding to the service account instance. This will:
      *    Create a service account private key.
      *    Store the private key in the Kubernetes secret as `privateKeyData`.
*   Create a binding to the BigQuery instance, referencing the service account.
    This will:
      *    Grant the service account the roles needed to use the BigQuery
           instance.
      *    Store the BigQuery instance connection information (`projectId`,
           `datasetId`) in a Kubernetes secret.

The GitHub data deployment consumes **both** secrets via environment variables
`GOOGLE_APPLICATION_CREDENTIALS`, `BIGQUERY_PROJECT`, and `BIGQUERY_DATASET`.
Review the deployment configuration in
[app-deployment.yaml](manifests/app-deployment.yaml).

### Step 4.1: Provision User Service Account

Create the user service account using the parameters in
[user-account-instance.yaml](./manifests/user-account-instance.yaml):

```shell
kubectl create -f ./manifests/user-account-instance.yaml
```

Check on the status of the service account provisioning:

```shell
svcat get instance --namespace bigquery user-service-account
```

Once the status is `Ready`, create a binding to make the service account
private key available in a secret, using the parameters in
[user-account-binding.yaml](manifests/user-account-binding.yaml).

```shell
kubectl create -f ./manifests/user-account-binding.yaml
```

Check the binding status:

```shell
svcat get binding --namespace bigquery
```

When the binding status is `Ready`, view the secret containing the service
account credentials:

```shell
kubectl get secret --namespace bigquery user-service-account -oyaml
```

Notice the `privateKeyData` value which contains the service account private
key.

### Step 4.2: Grant User Service Account Access to BigQuery

Create the user binding to the BigQuery instance using the parameters in
[user-bigquery-binding.yaml](./manifests/user-bigquery-binding.yaml):

```shell
kubectl create -f ./manifests/user-bigquery-binding.yaml
```

Check the binding status:

```shell
svcat get binding --namespace bigquery user-bigquery-binding
```

Once the `user-bigquery-binding` status is `Ready`, view the secret containing
the result of the binding (the default name of the secret is the same as the
name of the binding resource - `user-bigquery-binding`):

```shell
kubectl get secret -n bigquery $(kubectl get servicebinding -n bigquery user-bigquery-binding -o=jsonpath='{.spec.secretName}') -oyaml
```

Notice the `projectId` and `datasetId` values. They are referenced from
[app-deployment.yaml](manifests/app-deployment.yaml) and tell
the application deployment which dataset to access.

### Step 4.3: Create the Application Deployment

Create the Kubernetes deployment using configuration in
[app-deployment.yaml](manifests/app-deployment.yaml). The
configuration uses two secrets to obtain service account credentials
(`user-account-binding`) and BigQuery instance `projectId` and `datasetId`
(`user-bigquery-binding`).

```shell
kubectl create -f ./manifests/app-deployment.yaml
```

Wait for the deployment to complete, and then find the Kubernetes service
external IP address:

```shell
kubectl get service --namespace bigquery
IP= ... # External IP address of the Kubernetes load balancer.
```

or:

```shell
IP=$(kubectl get service --namespace bigquery bigquery-app-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Step 5: Access the Application

Use the IP address of the Kubernetes load balancer service along with a `curl`
command to access the application.

`GET /query` will return a list of results from running a query on the BigQuery
table of all GitHub commits, matching a specified prefix. If no parameter is
specified, the default prefix queried is: "this actually works".

```shell
# Query the commits that match the default commit prefix:
curl http://${IP}/query
{"entries":[{"name":"Blake","message":"this actually works\n"},...]}
```

`GET /query?prefix=...` will return a list of commits whose message matches
specified prefix; "fix the unit tests" in the example:

```shell
# Query the commits that match a custom commit prefix:
curl http://${IP}/query --data-urlencode prefix="fix the unit tests" --get
{"entries":[{"name":"aaronchang.tw","message":"fix the unit tests"},...]}
```

## Step 6: Cleanup

### Delete the user deployment

Delete the user deployment. It will stop serving user traffic and will cease
using the BigQuery dataset.

```shell
kubectl delete -f ./manifests/app-deployment.yaml
```

### Run the Cleanup Job

The cleanup job is executed once to delete the table from the BigQuery dataset.
BigQuery does not allow deleting a dataset with tables.

```shell
kubectl create -f ./manifests/cleanup-job.yaml
```

You can examine the dataset [here](https://bigquery.cloud.google.com)
to verify that the table has been successfully deleted.

### Cleanup Remaining Resources

To avoid incurring charges to your Google Cloud Platform account for the
resources used in this sample, delete and deprovision all resources.

An expedient way is to delete the Kubernetes namespace; however make sure that
the namespace doesn't contain any resources you want to keep:

```shell
kubectl delete namespace bigquery
```

Alternatively, delete all resources individually by running the following
commands:

**Note:** You may have to wait several minutes between steps to allow for the
previous operations to complete.

Delete the admin and cleanup Kubernetes jobs:

```shell
kubectl delete -f ./manifests/admin-job.yaml
kubectl delete -f ./manifests/cleanup-job.yaml
```

Delete the user binding to the BigQuery instance:

```shell
kubectl delete -f ./manifests/user-bigquery-binding.yaml
```

Delete the admin binding to the BigQuery instance. This also deletes
the service account created for the admin binding.

```shell
kubectl delete -f ./manifests/admin-bigquery-binding.yaml
```

Unbind the user service account:

```shell
kubectl delete -f ./manifests/user-account-binding.yaml
```

Deprovision the user service account:

```shell
kubectl delete -f ./manifests/user-account-instance.yaml
```

Deprovision the BigQuery instance:

```shell
kubectl delete -f ./manifests/bigquery-instance.yaml
```

If the `bigquery` namespace contains no resource you wish to keep,
delete it:

```shell
kubectl delete namespace bigquery
```

Remove all the roles associated with the service accounts `bigquery-admin` and
`bigquery-user` following [this
guide](https://cloud.google.com/iam/docs/granting-changing-revoking-access#revoking_access_to_team_members).

## Troubleshooting

Please find the troubleshooting information
[here](../README.md#troubleshooting).
