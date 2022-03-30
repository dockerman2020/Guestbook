#!/usr/bin/env python
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

import os
import rediswq
from model_training import FraudDetectionModelTrainer

# Initialize variables
FILESTORE_PATH = "/mnt/fileserver/"
TESTING_DATASET_PATH = FILESTORE_PATH + "datasets/testing/test_dataset.pkl"
OUTPUT_DIR = FILESTORE_PATH + "output/"
REPORT_PATH = OUTPUT_DIR + "report.txt"
CLASS_LABEL = "TX_FRAUD_SCENARIO"
QUEUE_NAME = "datasets"
HOST = "redis"

def main():
    """
    Workload which:
      1. Claims a filename from a Redis Worker Queue
      2. Reads the dataset from the file
      3. Partially trains the model on the dataset
      4. Saves a model checkpoint and generates a report on
         the performance of the model after the partial training.
      5. Removes the filename from the Redis Worker Queue
      6. Repeats 1 through 5 till the Queue is empty
    """
    q = rediswq.RedisWQ(name="datasets", host=HOST)
    print("Worker with sessionID: " + q.sessionID())
    print("Initial queue state: empty=" + str(q.empty()))
    checkpoint_path = None
    while not q.empty():
        # Claim item in Redis Worker Queue
        item = q.lease(lease_secs=20, block=True, timeout=2)
        if item is not None:
            dataset_path = item.decode("utf-8")
            print("Processing dataset: " + dataset_path)
            training_dataset_path = FILESTORE_PATH + dataset_path

            # Initialize the model training manager class
            model_trainer = FraudDetectionModelTrainer(
                training_dataset_path,
                TESTING_DATASET_PATH,
                CLASS_LABEL,
                checkpoint_path=checkpoint_path,
            )

            # Train model and save checkpoint + report
            checkpoint_path = model_trainer.train_and_save(OUTPUT_DIR)
            model_trainer.generate_report(REPORT_PATH)

            # Remove item from Redis Worker Queue
            q.complete(item)
        else:
            print("Waiting for work")

    print("Queue empty, exiting")


# Run workload
main()
