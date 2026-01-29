.. _DevicePlugin:

################################
Installing Furiosa Device Plugin
################################


Furiosa Device Plugin
================================================================
The Furiosa device plugin implements the `Kubernetes Device Plugin <https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/>`_
interface for FuriosaAI NPU devices, and its features are as follows:

* Discovering the Furiosa NPU devices and registering to a Kubernetes cluster.
* Tracking the health of the devices and reporting to a Kubernetes cluster.
* Running AI workload on the top of the Furiosa NPU devices within a Kubernetes cluster.

Deploying Furiosa Device Plugin with Helm
-----------------------------------------

The Furiosa device plugin helm chart is available at https://github.com/furiosa-ai/helm-charts.

You can deploy the Furiosa Device Plugin by running the following commands:

.. code-block:: sh

  helm repo add furiosa https://furiosa-ai.github.io/helm-charts
  helm repo update
  helm install furiosa-device-plugin furiosa/furiosa-device-plugin -n furiosa-system


Request Furiosa NPU Resource in Pod
----------------------------------------------

Furiosa NPU devices can be integrated into a Kubernetes cluster.
Each NPU card is exposed as a single resource.

The following table shows the expected resource names:

.. note::

.. list-table::
   :align: center
   :widths: 200 200
   :header-rows: 1

   * - Resource Name
     - Resource Count Per Card
   * - ``furiosa.ai/rngd``
     - ``1``


License
-------

.. code-block:: text

   Copyright 2023 FuriosaAI, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
