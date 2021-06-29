// Some of the Docker images from this repository are stored in Google Cloud's Artifact Registry (e.g., images referenced by GKE tutorials).
// Such images are rebuilt and repushed to Artifact Registry whenever related changes occur.
// The rebuilding/repushing is done by Google Cloud Build Triggers that we have set up in the "google-samples" Google Cloud project.
// This .tf file describes those Google Cloud Build Triggers and can be used to recreate them (e.g., in case they're accidentally deleted).
// How to use this file:
//     1. Install Terraform.
//     2. From this directory, run "terraform init". This will download the Google Terraform plugin.
//        If you get an error similar to "querying Cloud Storage failed: storage: bucket doesn't exist",
//        try running: gcloud auth application-default login
//     3. Finally, run "terraform apply" to create any missing Google Cloud Build Triggers.

terraform {
  backend "gcs" {
    bucket = "kubernetes-engine-samples"
    prefix = "terraform-state"
  }
}

provider "google" {
    project = "google-samples"
    region = "us-central1"
    zone = "us-central1-b"
}

locals {
    trigger_description = "This Cloud Build Trigger was created using Terraform (see github.com/GoogleCloudPlatform/kubernetes-engine-samples/tree/master/terraform)."
}

resource "google_cloudbuild_trigger" "cloud-pubsub" {
    name = "kubernetes-engine-samples-cloud-pubsub"
    filename = "cloud-pubsub/cloudbuild.yaml"
    included_files = ["cloud-pubsub/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "custom-metrics-direct-to-sd" {
    name = "kubernetes-engine-samples-custom-metrics-direct-to-sd"
    filename = "custom-metrics-autoscaling/direct-to-sd/cloudbuild.yaml"
    included_files = ["custom-metrics-autoscaling/direct-to-sd/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "custom-metrics-prometheus-to-sd" {
    name = "kubernetes-engine-samples-custom-metrics-prometheus-to-sd"
    description = local.trigger_description
    filename = "custom-metrics-autoscaling/prometheus-to-sd/cloudbuild.yaml"
    included_files = ["custom-metrics-autoscaling/prometheus-to-sd/**"]

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "guestbook-php-redis" {
    name = "kubernetes-engine-samples-guestbook-php-redis"
    filename = "guestbook/php-redis/cloudbuild.yaml"
    included_files = ["guestbook/php-redis/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "guestbook-redis-follower" {
    name = "kubernetes-engine-samples-guestbook-redis-follower"
    filename = "guestbook/redis-follower/cloudbuild.yaml"
    included_files = ["guestbook/redis-follower/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "hello-app" {
    name = "kubernetes-engine-samples-hello-app"
    filename = "hello-app/cloudbuild.yaml"
    included_files = ["hello-app/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "hello-app-cdn" {
    name = "kubernetes-engine-samples-hello-app-cdn"
    filename = "hello-app-cdn/cloudbuild.yaml"
    included_files = ["hello-app-cdn/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "hello-app-redis" {
    name = "kubernetes-engine-samples-hello-app-redis"
    filename = "hello-app-redis/cloudbuild.yaml"
    included_files = ["hello-app-redis/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "hello-app-tls" {
    name = "kubernetes-engine-samples-hello-app-tls"
    filename = "hello-app-tls/cloudbuild.yaml"
    included_files = ["hello-app-tls/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}

resource "google_cloudbuild_trigger" "whereami" {
    name = "kubernetes-engine-samples-whereami"
    filename = "whereami/cloudbuild.yaml"
    included_files = ["whereami/**"]
    description = local.trigger_description

    github {
        owner = "GoogleCloudPlatform"
        name = "kubernetes-engine-samples"
        push {
            branch = "^master$"
        }
    }
}
