## CloudSQL Example

This example shows how to build an application that uses [Google Cloud SQL](https://cloud.google.com/sql/docs/introduction)
using Kubernetes and [Docker](https://www.docker.com/).

The example consists of a single pod containing two containers:

- A web frontend container running Wordpress.
- A [Cloud SQL Proxy](https://github.com/GoogleCloudPlatform/cloudsql-proxy/) container providing connectivity to Cloud SQL.

### Prerequisites

This example requires a running Kubernetes cluster at version 1.2 or higher. See the [Getting Started guides](https://cloud.google.com/container-engine/docs/before-you-begin)
for how to get started. As noted above, if you have a Google Container Engine cluster set up, go [here](https://cloud.google.com/container-engine/docs/tutorials/guestbook) instead.

To check your version, run `kubectl version` and make sure both the client and server report a version of at least 1.2.

You'll also need a running Cloud SQL instance, and a Service Account that can access the instance.
See the [Cloud SQL access control](https://cloud.google.com/sql/docs/access-control) page for more information.

#### Create Secrets

You'll need to create several `Secret` resources to allow the SQL proxy to connect with your SQL instance. First, you'll need to
create a secret resource containing the Service Account credentials to allow the proxy to communicate with the Cloud SQL API.

First, create and download the JSON Service Account credentials following [these steps](https://developers.google.com/identity/protocols/OAuth2ServiceAccount#creatinganaccount).
Then run this command, making sure to replace PATH_TO_CREDENTIAL_FILE with the correct location of the JSON file:

```
kubectl create secret generic cloudsql-oauth-credentials --from-file=credentials.json=<PATH_TO_CREDENTIAL_FILE>
```

Next, you'll need to create another pair of secrets to allow the proxy to connect to the actual SQL instance. This will contain
the SQL username and password you'd like to connect as. Make sure to replace the USERNAME and PASSWORD values.

```
kubectl create secret generic cloudsql --from-literal=username=<USERNAME> --from-literal=password=<PASSWORD>
```

#### Create Pod

Next, open the cloudsql_deployment.yaml file in this repository and replace the values `$PROJECT`, `REGION` and `$INSTANCE`
with the correct values for your SQL instance.

Then, run:

```
kubectl create -f cloudsql_deployment.yaml
```

to bring up the pod.

#### Access the Wordpress Installation

You can setup port-forwarding to your wordpress installation to access it over localhost.

First find the pod name:

```
$ kubectl get pods
NAME                         READY     STATUS    RESTARTS   AGE
wordpress-2668199741-wvaup   2/2       Running   1          1m
```

Then use `kubectl port-forward`:

```
kubectl port-forward wordpress-2668199741-wvaup 8080:80
```

Then open `localhost:8080` in your browser. You should see the Wordpress installation screen.
