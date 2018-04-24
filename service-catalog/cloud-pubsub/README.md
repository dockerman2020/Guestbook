# Service Catalog Sample - Cloud Pub/Sub

This sample demonstrates how to build a simple Kubernetes web application using
the [Kubernetes Service
Catalog](https://kubernetes.io/docs/concepts/service-catalog/) and the [Google
Cloud Platform Service
Broker](https://cloud.google.com/kubernetes-engine/docs/concepts/add-on/service-broker),
an implementation of the [Open Service Broker
API](https://www.openservicebrokerapi.org/).

The sample highlights a number of Kubernetes and Open Service Broker concepts:

*   Using the Service Catalog and the Service Broker to ***provision*** a
    *service instance*.
*   ***Binding*** the provisioned *service instance* to a Kubernetes
    application.
*   Use of the *service binding* by the application to access the *service
    instance*.

The sample contains two web applications using [Cloud
Pub/Sub](https://cloud.google.com/pubsub/); it allows users to publish messages,
and then to receive the published messages.

At the core of the sample is a Cloud Pub/Sub topic and a subscription to the
topic accessed by different components of the sample. The Cloud Pub/Sub topic,
which is a corresponding resource of Cloud Pub/Sub service instance, is
provisioned by the Service Broker and the subscription is completed by the
binding to the service instance. The service instance is accessed by two
Kubernetes applications: a publisher web deployment and a subscriber web
deployment.

The `publisher` web deployment uses a binding which allows it to publish
messages to the Cloud Pub/Sub topic.

The `subscriber` web deployment uses a binding which not only creates a
subscription to the topic, but also grants subscriber-level access to the
Pub/Sub topic - to receive the messages from its subscribed topic.

## Objectives

To deploy and run the sample application, you must:

1.  Create a new Kubernetes namespace.
2.  Create a Cloud Pub/Sub topic.
3.  Deploy the publisher application:
    1.  Create a Cloud Pub/Sub service binding for publisher.
    2.  Create the Kubernetes publisher deployment.
4.  Deploy the subscriber application:
    1.  Provision a Cloud IAM service account instance and bind to it.
    2.  Create a Cloud Pub/Sub service binding for subscriber.
    3.  Create the Kubernetes subscriber deployment.
5.  Use the publisher and subscriber applications.
6.  Clean up.

## Before you begin

Review the [information](../README.md) applicable to all Service Catalog
samples, including [prerequisites](../README.md#prerequisites) and
[troubleshooting](../README.md#troubleshooting).

Successful use of this sample requires:

*   A Kubernetes cluster, minimum version 1.8.x.
*   Kubernetes Service Catalog and the Service Broker [installed](
    https://cloud.google.com/kubernetes-engine/docs/how-to/add-on/service-broker/install-service-catalog).
*   The Service Catalog CLI (`svcat`) [installed](
    https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md#installing-the-service-catalog-cli).
*   The [Cloud Pub/Sub API](
    https://console.cloud.google.com/apis/library/pubsub.googleapis.com)
    enabled.

## Step 1: Create a New Kubernetes Namespace

```shell
kubectl create namespace pubsub
```

**TIP:** For convenience, also set the `NAMESPACE` environment variable in your
shell: `export NAMESPACE=pubsub`. `svcat` will use this namespace by default so
you don't have to use `--namespace pubsub` on every command below.

## Step 2: Create a Cloud Pub/Sub Topic

To create a Cloud Pub/Sub topic, run:

```shell
svcat provision pubsub-instance \
    --class cloud-pubsub \
    --plan beta \
    --param topicId=comments \
    --namespace pubsub
```

This command will use the Kubernetes Service Catalog to provision a Cloud
Pub/Sub service instance with the parameters set by `--param` flag. The
corresponding resource is a Pub/Sub topic called "comments".

Check on the status of the service instance provisioning:

```shell
svcat get instance pubsub-instance --namespace pubsub
```

The service instance is provisioned when its status is `Ready`. You can also
examine the Cloud Pub/Sub topic in the [Google Cloud console](
https://console.cloud.google.com/cloudpubsub/topicList).

## Step 3: Deploy the Publisher Application

The `publisher` deployment checks the existence of the Pub/Sub topic and
publishes messages to the topic. To do so, it requires permissions from the
Pub/Sub `publisher` and `viewer` roles.

To express the intent of granting the permissions to `publisher`, you will
create a *service binding* using the parameters in
[publisher-binding.yaml](manifests/publisher-binding.yaml). Creating the
service binding will:

*   Create a service account to authenticate with Cloud Pub/Sub.
*   Grant the service account the `roles/pubsub.publisher` and
    `roles/pubsub.viewer` roles.
*   Store the service account private key (`privateKeyData`) and the Pub/Sub
    topic information (`projectId`, `topicId`) in a Kubernetes secret.

The `publisher` deployment consumes the information stored in the secret via
environment variables `GOOGLE_APPLICATION_CREDENTIALS`, `GOOGLE_CLOUD_PROJECT`,
and `PUBSUB_TOPIC`. Review the publisher deployment configuration in
[publisher-deployment.yaml](manifests/publisher-deployment.yaml).

### Step 3.1: Create a Cloud Pub/Sub Service Binding for Publisher

Create the publisher binding:

```shell
kubectl create -f manifests/publisher-binding.yaml
```

The command will use the Kubernetes Service Catalog to create a binding to the
Pub/Sub service instance provisioned earlier.

Check on the status of the binding operation:

```shell
svcat get binding publisher --namespace pubsub
```

Once the binding status is `Ready`, view the Kubernetes secret containing the
result of the binding (the default name of the secret is the same as the name
of the binding resource - `publisher`).

```shell
kubectl get secret --namespace pubsub publisher -oyaml
```

Notice the values `privateKeyData`, `projectId`, and `topicId` which contain
the result of the binding, ready for the publisher deployment to use.

### Step 3.2: Create the Kubernetes Publisher Deployment

Create the Kubernetes deployment using configuration in
[publisher-deployment.yaml](manifests/publisher-deployment.yaml):

```shell
kubectl create -f ./manifests/publisher-deployment.yaml
```

Wait for the deployment to complete and then find the the Kubernetes service
external IP address:

```shell
kubectl get service publisher-service --namespace pubsub
PUB_IP= ... # External IP address of the Kubernetes load balancer.
```

or:

```shell
PUB_IP=$(kubectl get service --namespace pubsub publisher-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Step 4: Deploy the Subscriber Application

The `subscriber` deployment uses the subscription to the Pub/Sub topic, pulls
the messages published to the topic, and displays the received messages. To do
so, it requires permissions from the `subscriber` and `viewer` roles.

In order to acquire the permissions, `subscriber` will run as a single service
account with `subscriber` and `viewer` roles granted. Next, you will:

*   Create a service account by provisioning a service instance for "Cloud IAM
    service account".
*   Create a binding to the service account instance. This will:
      *   Create a service account private key.
      *   Store the private key in the Kubernetes secret as `privateKeyData`.
*   Create a binding to Cloud Pub/Sub service instance, using the subscription
    ID and referencing the service account. This will:
      *   Create a subscription to the corresponding topic of the Pub/Sub
          service instance.
      *   Grant the service account the roles needed to subcribe to the topic
          and view the published messages.
      *   Store the Cloud Pub/Sub subscription information (`projectId`,
           `subscriptionId`) in a Kubernetes secret.

The `subscriber` deployment consumes **both** secrets via environment variables
`GOOGLE_APPLICATION_CREDENTIALS`, `GOOGLE_CLOUD_PROJECT`, and
`PUBSUB_SUBSCRIPTION`. Review the subscriber deployment configuration in
[subscriber-deployment.yaml](manifests/subscriber-deployment.yaml).

### Step 4.1: Provision a Cloud IAM Service Account Instance and Bind to It

Create a subscriber service account:

```shell
svcat provision subscriber-account \
    --class cloud-iam-service-account \
    --plan beta \
    --param accountId=pubsub-subscriber \
    --namespace pubsub
```

Check on the status of the service account provisioning:

```shell
svcat get instance --namespace pubsub subscriber-account
```

Once the status is `Ready`, create a binding to make the service account
private key available in a Kubernetes secret:

```shell
svcat bind subscriber-account \
    --namespace pubsub \
    --name subscriber-account-credentials
```

Check the binding status:

```shell
svcat get binding --namespace pubsub subscriber-account-credentials
```

When the binding status is `Ready`, view the secret containing the service
account credentials:

```shell
kubectl get secret --namespace pubsub subscriber-account-credentials -oyaml
```

Notice the `privateKeyData` value which contains the service account private
key.

### Step 4.2: Create a Cloud Pub/Sub Service Binding for Subscriber

Create the subscriber binding to the Cloud Pub/Sub service instance using the
configuration in [subscriber-binding.yaml](manifests/subscriber-binding.yaml):

```shell
kubectl create -f ./manifests/subscriber-binding.yaml
```

Check the binding status:

```shell
svcat get binding --namespace pubsub subscriber
```

Once the binding status is `Ready`, view the secret containing the result of the
binding (the default name of the secret is the same as the name of the binding
resource - `subscriber`):

```shell
kubectl get secret --namespace pubsub subscriber -oyaml
```

Notice the `projectId` and `subscriptionId` values. They are referenced from
[subscriber-deployment.yaml](manifests/subscriber-deployment.yaml) and tell
the subscriber deployment which subscription to use.

### Step 4.3: Create the Kubernetes Subscriber Deployment

Create the Kubernetes deployment using configuration in
[subscriber-deployment.yaml](manifests/subscriber-deployment.yaml). The
configuration uses two secrets to obtain service account credentials
(from secret `subscriber-account-credentials` ) and Pub/Sub subscription
`projectId` and `subscriptionId` (from secrect `subscriber`).

```shell
kubectl create -f ./manifests/subscriber-deployment.yaml
```

Wait for the deployment to complete and then find the the Kubernetes service
external IP address:

```shell
kubectl get service subscriber-service --namespace pubsub
SUB_IP= ... # External IP address of the Kubernetes load balancer.
```

or:

```shell
SUB_IP=$(kubectl get service --namespace pubsub subscriber-service -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Step 5: Use the Publisher and Subscriber Applications

Use the IP addresses of the Kubernetes load balancer services in the web browser
to access the application.

Enter the IP address of `subscriber-service`, `${SUB_IP}`, into your browser. No
messages have been received.

Enter the IP address of `publisher-service`, `${PUB_IP}`, into your browser and,
in the simple UI, enter messages to send to the subscriber.

Refresh the subscriber browser window - the messages have now been received.

## Step 6: Clean up

To avoid incurring charges to your Google Cloud Platform account for the
resources used in this sample, delete and deprovision all resources.

An expedient way is to delete the Kubernetes namespace; however make sure that
the namespace doesn't contain any resources you want to keep:

```shell
kubectl delete namespace pubsub
```

Alternatively, delete all resources individually by running the following
commands:

**Note:** You may have to wait several minutes between steps to allow for the
previous operations to complete.

Delete the publisher deployment and subscriber deployment. They will stop
serving user traffic and will cease using the Cloud Pub/Sub topic.

```shell
kubectl delete -f ./manifests/publisher-deployment.yaml
kubectl delete -f ./manifests/subscriber-deployment.yaml
```

Delete the subscriber binding to the Cloud Pub/Sub service instance:

```shell
kubectl delete -f ./manifests/subscriber-binding.yaml
```

Delete the publisher binding to the Cloud Pub/Sub service instance. This also
deletes the service account created for the publisher binding.

```shell
kubectl delete -f ./manifests/publisher-binding.yaml
```

Unbind the subscriber service account instance:

```shell
svcat unbind subscriber-account \
    --name subscriber-account \
    --namespace pubsub
```

Deprovision the subscriber service account instance:

```shell
svcat deprovision subscriber-account --namespace pubsub
```

Deprovision the Cloud Pub/Sub service instance:

```shell
svcat deprovision pubsub-instance --namespace pubsub
```

If the `pubsub` namespace contains no resource you wish to keep, delete it:

```shell
kubectl delete namespace pubsub
```

## Troubleshooting

Please find the troubleshooting information
[here](../README.md#troubleshooting).
