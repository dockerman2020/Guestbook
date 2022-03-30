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

from pathlib import Path
import pandas as pd
import os
import datetime
import pickle
from sklearn.linear_model import SGDClassifier
from sklearn.metrics import accuracy_score


class FraudDetectionModelTrainer:
    """
    Machine Learning Model Training Management

    This object is used to manage the lifecycle of a fraudulent transaction
    detection model. It partially trains the model on batched datasets and
    generates model checkpoints and accuracy reports with each batch job.
    """
    def __init__(
        self,
        train_dataset_path,
        test_dataset_path,
        label,
        checkpoint_path=None
    ):
        # self._classes define the possible labels the model can expect to
        # encounter as it evaluates partial fitting with each batch of new data.
        #  0 = transaction is not fraudulent
        #  1, 2, 3 = transaction marked as fraudulent through different methods
        self._classes = [0, 1, 2, 3]        # 4 different fraud scenarios
        self._train_dataset_path = train_dataset_path
        self._test_dataset_path = test_dataset_path
        self._label = label

        # If a filepath to a model checkpoint is provided, load the model with
        # the parameters from the checkpoint.
        # Otherwise, instantiate a new model.
        if checkpoint_path:
            with open(checkpoint_path, 'rb') as f:
                self._model = pickle.load(f)
        else:
            self._model = SGDClassifier(warm_start=True)

    def get_model(self):
        """
        Return the model object.
        """
        return self._model

    def get_features_and_labels(self, dataframe):
        """
        Partition the dataframe into arrays of features and labels.
        """
        features = dataframe.drop(columns=self._label, axis=1)
        labels = dataframe[self._label]

        return features, labels

    def get_model_accuracy(self, features, labels):
        """
        Accepts a dataset to partially train the model and returns a
        score for the model's prediction accuracy.
        """
        features_prediction = self._model.predict(features)
        accuracy = accuracy_score(features_prediction, labels)
        return accuracy

    def _read_dataset(self, dataset_path):
        """
        Accepts a filepath containing the dataset and
        returns the deserializes pandas.DataFrame.
        """
        dataset = pd.read_pickle(dataset_path)
        return dataset

    def _get_checkpoint_name(self):
        """
        Returns a filename for the model checkpoint.
        """
        dataset_basename = Path(self._train_dataset_path).resolve().stem
        filename = "model_cpt_{}.pkl".format(dataset_basename)
        return filename

    def _save_model(self, checkpoint_dir):
        """
        Accepts a directory path and returns the filepath where
        the checkpoint for the model is saved.
        Creates the destination directory if it does not exist.
        """
        # Check whether the specified path exists or not
        isExist = os.path.exists(checkpoint_dir)

        if not isExist:
            # Create a new directory because it does not exist
            os.makedirs(checkpoint_dir)

        filename = self._get_checkpoint_name()
        path = checkpoint_dir + filename

        # Serialize the model checkpoint in to a Python Pickle file
        with open(path, 'wb') as f:
            pickle.dump(self._model, f)
        return path

    def train_and_save(self, checkpoint_dir):
        """
        Accepts a directory path.
        Returns the filepath where the model checkpoint is saved after being
        partially trained.
        """
        dataset = self._read_dataset(self._train_dataset_path)
        features, labels = self.get_features_and_labels(dataset)
        self._model.partial_fit(features, labels, classes=self._classes)
        checkpoint_path = self._save_model(checkpoint_dir)
        return checkpoint_path

    def generate_report(self, output_path):
        """
        Accepts a output filepath.
        Trains the model and appends the training report to the provided file.
        """
        generated_on = str(datetime.datetime.now())
        checkpoint_name = self._get_checkpoint_name()
        dataset_name = Path(self._train_dataset_path).resolve().name
        train_features, train_labels = self.get_features_and_labels(
            self._read_dataset(self._train_dataset_path)
        )
        test_features, test_lables = self.get_features_and_labels(
            self._read_dataset(self._test_dataset_path)
        )
        training_accuracy = self.get_model_accuracy(
            train_features,
            train_labels
        )
        test_accuracy = self.get_model_accuracy(
            test_features,
            test_lables,
        )
        with open(output_path, 'a') as f:
            report = (
                "*****************************************************\n"
                "Report generated on: {}\n"
                "Training dataset: {}\n"
                "Model checkpoint: {}\n"
                "---\n"
                "Accuracy on training data: {}\n"
                "Accuracy on testing data: {}\n"
                "\n"
            ).format(
                generated_on,
                dataset_name,
                checkpoint_name,
                training_accuracy,
                test_accuracy,
            )
            f.writelines(report)
