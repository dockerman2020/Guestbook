# Workload Metrics example

This tutorial demonstrates how to automatically scale your Google Kubernetes Engine (GKE)
workloads based on Prometheus-style metrics emitted by your application.  It uses the [GKE workload
metrics](https://cloud.google.com/stackdriver/docs/solutions/gke/managing-metrics#workload-metrics)
pipeline to collect the metrics emitted from the example application and send them to
[Cloud Monitoring](https://cloud.google.com/monitoring), and then uses the
[HorizontalPodAutoscaler](https://cloud.google.com/kubernetes-engine/docs/concepts/horizontalpodautoscaler)
along with the [Custom Metrics Adapter](https://github.com/GoogleCloudPlatform/k8s-stackdriver/tree/master/custom-metrics-stackdriver-adapter) to scale the application.