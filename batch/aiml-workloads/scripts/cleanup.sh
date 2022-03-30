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

source scripts/variables.sh

echo "*********************************"
echo "Deleting GCP resources"
echo "*********************************"

echo "Deleting GKE cluster..."
gcloud container clusters delete ${CLUSTER_NAME} \
    --project=${PROJECT_ID} --zone=${ZONE}
echo "GKE cluster '${CLUSTER_NAME}' deleted."

echo "Deleting Filestore instance..."
gcloud filestore instances delete ${FILESTORE_ID} --zone=${ZONE}
echo "Filestore instance '${FILESTORE_ID}' deleted."

