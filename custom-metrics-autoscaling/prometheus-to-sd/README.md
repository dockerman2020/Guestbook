# Prometheus dummy exporter

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg)](https://ssh.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https://github.com/GoogleCloudPlatform/kubernetes-engine-samples&cloudshell_workspace=custom-metrics-autoscaling/prometheus-to-sd&cloudshell_tutorial=README.md)


A simple prometheus-dummy-exporter container exposes a single Prometheus metric with a constant value. The metric name, value and port on which it will be served can be passed by flags.

This container is then deployed in the same pod with another container, prometheus-to-sd, configured to use the same port. It scrapes the metric and publishes it to Stackdriver. This adapter isn't part of the sample code, but a standard component used by many Kubernetes applications. You can learn more about it
[here](https://github.com/GoogleCloudPlatform/k8s-stackdriver/tree/master/prometheus-to-sd).

# Build

Provided manifest files use already available images. You don't need to do
anything else to use them. The following steps are only applicable if you want to
build your own image.

If you would like to build this image locally and deploy via Arifact Registry, checkout [Artifact Registry Quickstart for Docker](https://cloud.google.com/artifact-registry/docs/docker/quickstart).
