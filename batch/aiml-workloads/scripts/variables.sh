#!/usr/bin/env sh
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "*********************************"
echo "Initializing variables"
echo "*********************************"
# Replace the following with values for your own  project.
export PROJECT_ID="<YOUR_PROJECT_ID>"
export REGION="<YOUR_REGION>"
export ZONE="<YOUR_ZONE>"

export CLUSTER_NAME="batch-aiml"
export AR_REPO_ID="batch-aiml-docker-repo"
export FILESTORE_ID="batch-aiml-filestore"

export DATASETS_DIR="datasets"
export QUEUE_NAME="datasets"
export POD_NAME="redis-leader"
export PVC_PATH="/mnt/fileserver"

echo "PROJECT_ID=${PROJECT_ID}"
echo "REGION=${REGION}"
echo "ZONE=${ZONE}"
echo "CLUSTER_NAME=${CLUSTER_NAME}"
echo "AR_REPO_ID=${AR_REPO_ID}"
echo "FILESTORE_ID=${FILESTORE_ID}"
echo "DATASETS_DIR=${DATASETS_DIR}"
echo "QUEUE_NAME=${QUEUE_NAME}"
echo "POD_NAME=${POD_NAME}"
echo "PVC_PATH=${PVC_PATH}"
echo "*********************************"
