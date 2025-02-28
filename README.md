# Furiosa Device Plugin for Kubernetes

<!-- ADD TOC HERE -->

## Overview
The furiosa device plugin implements the [Kubernetes Device Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) interface for FuriosaAI NPU devices.  
It allows:
- discovering the Furiosa NPU devices and register to a Kubernetes cluster.
- tracking the health of the devices and report to a Kubernetes cluster.
- running AI workload on the top of the Furiosa NPU devices within a Kubernetes cluster.

## Preparing NPU nodes
### Driver & Firmware
See [following guide](https://furiosa-ai.github.io/docs/latest/en/software/installation.html) to install the driver and firmware.
<!-- update this once operator is ready -->

### Kubernetes
We do support the following versions of Kubernetes and CRI runtime:
- Kubernetes: v1.20.0 or later
- CRI Runtime: [containerd](https://github.com/containerd/containerd) or [CRI-O](https://github.com/cri-o/cri-o)

> [!NOTE]  
> Docker is officially deprecated as a container runtime in Kubernetes.
> It is recommended to use containerd or CRI-O as a container runtime.
> Otherwise you may face unexpected issues with the device plugin.
> For more information, see [here](https://kubernetes.io/blog/2020/12/02/dont-panic-kubernetes-and-docker/).

## Configuration
The configuration should be written in following format as yaml and located at `/etc/config/config.yaml`.

- If resourceStrategy is not specified, the default value is `"none"`.
 - If debugMode is not specified, the default value is `false`.
- If disabledDeviceUUIDs is not specified, the default value is empty list `[]`.
```yaml
partitioning: "none"
debugMode: false
disabledDeviceUUIDs:
  - "uuid1"
  - "uuid2"
```

The Furiosa NPU can be integrated into the Kubernetes cluster in various configurations. A single NPU card can either be exposed as a single resource or partitioned into multiple resources. Partitioning into multiple resources allows for more granular control.

The following table shows the available configurations:

| Strategy/Arch | Warboy                       |                         | Renegade                    |                         |
|---------------|------------------------------|-------------------------|-----------------------------|-------------------------|
|               | Resource Name                | Resource Count Per Card | Resource Name               | Resource Count Per Card |
| generic       | furiosa.ai/warboy            | 1                       | furiosa.ai/rngd             | 1                       |
| single-core   | furiosa.ai/warboy-1core.8gb  | 2                       | furiosa.ai/rngd-1core.6gb   | 8                       |
| dual-core     | furiosa.ai/warboy-2core.16gb | 1                       | furiosa.ai/rngd-2core.12gb  | 4                       |
| quad-core     | Not Supported                | Not Supported           | furiosa.ai/rngd-4core.24gb  | 2                       |


## Deployment

### Deploy with kubectl
The deployment yaml file are available at [deployments/raw/furiosa-device-plugin-ds.yaml](deployments/raw/furiosa-device-plugin-ds.yaml).
You can deploy the device plugin by running the following command:
```bash
kubectl apply -f deployments/raw/furiosa-device-plugin-ds.yaml
```

### Deploy with Helm
The helm chart is available at [deployment/helm](deployments/helm) directory.
To configure deployment as you need, you can modify [deployments/helm/values.yaml](deployments/helm/values.yaml).

You can deploy the device plugin by running the following command:
```bash
helm install furiosa-device-plugin deployments/helm -f deployments/helm/values.yaml -n kube-system
```

<!-- add deploy with npu operator here -->


### Verify deployment
Following command should show the Furiosa NPU devices in the Kubernetes cluster.
```bash
kubectl get nodes -o json | jq -r '.items[] | .metadata.name as $name | .status.capacity | to_entries | map("    \(.key): \(.value)") | $name + ":\n  capacity:\n" + join("\n")'
```


## Notes
