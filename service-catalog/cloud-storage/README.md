# Service Catalog Sample - Cloud Storage

This sample demonstrates how to build a simple Kubernetes web application using
[Kubernetes Service
Catalog](https://kubernetes.io/docs/concepts/service-catalog/) and a [Google
Cloud Platform Service
Broker](https://cloud.google.com/kubernetes-engine/docs/concepts/add-on/service-broker),
an implementation of the [Open Service
Broker](https://www.openservicebrokerapi.org/) standard.

The sample highlights a number of Kubernetes and Open Service Broker concepts:

*   Using Service Catalog and the Service Broker to *provision*
    a *service instance*.
*   *Binding* the provisioned *service instance* to a Kubernetes application.
*   Use of the *binding* by the application to access the *service instance*.

The sample application exposes a simple web API which allows its clients to
store and retrieve quotes by famous people. The application uses
[Cloud Storage](https://cloud.google.com/storage/) to store the data.

An instance of a Cloud Storage service is provisioned in your project by the
Service Broker. Then, two separate components access the Cloud Storage instance
using *bindings*. These components are the `quotes` application and an
administrative `cleanup` job.

The `quotes` application uses a binding which will allow it to create and read
objects from the Cloud Storage bucket.

The `cleanup` job uses a binding which will allow it to delete objects from the
Cloud Storage bucket in preparation for deprovisioning.

## Objectives

To deploy and run the sample `quotes` application, you must:

1.  Create a new Kubernetes namespace.
2.  Provision a new Cloud Storage instance using Kubernetes Service Catalog.
3.  Deploy the `quotes` application:
4.  Interact with the  `quotes` application.
5.  Deprovision and delete all resources used by the sample.
    1.  Delete the `quotes` application.
    2.  Deploy the `cleanup` job to prepare instance for deprovisioning.
    3.  Deprovision all resources and delete namespace

## Before you begin

Review the [information](../README.md) applicable to all Service Catalog
samples, including [prerequisites](../README.md#prerequisites):

*   A Kubernetes cluster, minimum version 1.8.x.
*   Kubernetes Service Catalog and the Service Broker [installed](
    https://cloud.google.com/kubernetes-engine/docs/how-to/add-on/service-broker/install-service-catalog).
*   The Service Catalog CLI (`svcat`) [installed](
    https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#installing-the-service-catalog-cli).

## Step 1: Create a Kubernetes namespace

```shell
kubectl create namespace storage-quotes
```

## Step 2: Provisioning Cloud Storage

Provision an instance of Cloud Storage:

```shell
svcat provision storage-instance --namespace storage-quotes \
    --class cloud-storage --plan beta \
    --param bucketId=quotes-$(uuidgen | tr A-Z a-z) \
    --param location=us-central1
```

This command will use the Kubernetes Service Catalog to provision an instance
of a Cloud Storage service, which will create a Cloud Storage bucket.

Check on the provisioning status:

```shell
svcat get instance --namespace storage-quotes storage-instance
```

The instance is provisioned when status is `Ready`.

## Step 3: Deploy the Application

The `quotes` deployment reads data from and writes data to the Cloud Storage
bucket. It creates new objects in the bucket, lists objects in the bucket, and
reads contents of the objects. It does not delete objects from the bucket.

To perform these operations, the `quotes` deployment will assume an identity
of a service account with sufficient privileges.

Because the `quotes` application only uses a single Google Cloud service
(Cloud Storage), it can can create a service account as part of binding
to the Cloud Storage instance.

**Note**: If the sample were extended to use another Google Cloud service, for
example Pub/Sub to notify subscribers when new quote was created, it would
be appropriate to create a single service account for the application. The
application would use it to authenticate with all Google Cloud services on which
it depends.

To deploy the `quotes` application, you will use the
[quotes-deployment.yaml](./manifests/quotes-deployment.yaml) manifest.

Applying the manifest to your Kubernetes cluster will:

*  Create a binding to the Cloud Storage instance. This will:
   *  Create a new service account for the binding.
   *  Grant the service account requested roles.
   *  Create a service account private key.
*  Store the Bucket information (`projectId` and `bucketId`) and service account
   private key (`privateKeyData`) in a Kubernetes secret `user-storage-binding`.
*  Create a Kubernetes deployment which uses the Kubernetes secret as input
   parameters.

Create the binding and the `quotes` deployment using parameters in
[quotes-deployment.yaml](./manifests/quotes-deployment.yaml):

```shell
kubectl create -f ./manifests/quotes-deployment.yaml
```

Check the binding status:

```shell
svcat get binding --namespace storage-quotes user-storage-binding
```

Once the `user-storage-binding` status is `Ready`, view the secret containing
the result of the binding. The default name of the secret is the same as the
name of the binding resource - `user-storage-binding`:

```shell
kubectl get secret --namespace storage-quotes user-storage-binding -oyaml
```

Notice the following values in particular:

| Value            | Contains                                                |
| ---------------- | ------------------------------------------------------- |
| `privateKeyData` | service account private key                             |
| `bucketId`       | Cloud Storage bucket to which the binding grants access |

These values are used by the `quotes` deployment.

As soon as the secret created by the binding exists, Kubernetes will proceed
creating the deployment pods. Check on the status of the deployment:

```shell
kubectl get deployment --namespace storage-quotes
```

Wait for the deployment to complete and find the external IP address of the
`quotes-service` load balancer:

```shell
kubectl get service --namespace storage-quotes quotes-service
```

Save the external IP address in an `IP` environment variable:

```shell
IP=$(kubectl get service --namespace storage-quotes quotes-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

You are now ready to access the application's web API.

## Step 4: Use the Quotes Application

Use the IP address of the Kubernetes load balancer service along with a `curl`
command to acess the application.

`GET /quotes` will return a list of quotes in JSON format:

```shell
# Query the quotes:
curl http://${IP}/quotes
{"quotes":[]}
```

An `HTTP POST` is used to add a new quote:

```shell
# Create a new quote:
curl http://${IP}/quotes -d '{"person": "Dalai Lama", "quote": "Be kind whenever possible. It is always possible."}'

# Query the quotes again:
curl http://${IP}/quotes
{"quotes":[{"person":"Dalai Lama","quote":"Be kind whenever possible. It is always possible."}]}
```

Congratulations! You have just deployed an application which accesses services
provisioned by the Service Broker.

## Step 5: Cleanup

### Step 5.1: Delete the Quotes Deployment

Delete the `quotes` deployment. It will stop serving user traffic and stop
using the Cloud Storage instance:

```shell
kubectl delete -f ./manifests/quotes-deployment.yaml
```

### Step 5.2: Run the Cleanup Job

The cleanup job is executed once to delete any leftover objects in the Cloud
Storage instance. It lists all objects in the bucket and deletes them.
The cleanup job requires privileges to delete objects from the Cloud Storage
bucket, as part of deploying the cleanup job, new binding is created with
sufficient privileges.

The `cleanup` job also uses simplified flow of creating binding which creates
a new service account as part of binding.

Create the cleanup binding and the cleanup job using the configuration in
[cleanup-job.yaml](./manifests/cleanup-job.yaml):

```shell
kubectl create -f ./manifests/cleanup-job.yaml
```

Check on the status of the binding creation:

```shell
svcat get binding --namespace storage-quotes cleanup-storage-binding
```

Once the binding status is `Ready`, Kubernetes will automatically execute
the cleanup job itself.

Check on the status of the cleanup job:

```
kubectl get job --namespace storage-quotes storage-cleanup-job
```

You can examine the bucket in the [Google Cloud Console](https://console.cloud.google.com/storage/browser)
to verify that the bucket is now empty.

**Note:** If you cannot list the buckets, you must explicitly grant yourself
the "Storage Admin" role in your project.

### Step 5.3: Cleanup Remaining Resources

To avoid incurring charges to your Google Cloud Platform account for the
resources used in this sample, delete and deprovision all resources.

An expedient way is to delete the Kubernetes namespace; however make sure that
the namespace doesn't contain any resources you want to keep:

```shell
kubectl delete namespace storage-quotes
```

Alternatively, delete all resources individually by running the following
commands:

**Note:** You may have to wait several minutes between steps to allow for the
previous operations to complete.

Delete the cleanup Kubernetes job. This also deletes the binding used by the
cleanup job, as well as the service account created for the cleanup binding:

```shell
kubectl delete -f ./manifests/cleanup-job.yaml
```

Deprovision the Cloud Storage instance. This will delete the bucket:

```shell
svcat deprovision --namespace storage-quotes storage-instance
```

If the `storage-quotes` namespace contains no resources you wish to keep,
delete it:

```shell
kubectl delete namespace storage-quotes
```

## Troubleshooting

Please find the troubleshooting information
[here](../README.md#troubleshooting).
