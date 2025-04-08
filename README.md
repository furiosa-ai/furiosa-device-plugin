# Furiosa Device Plugin

The Furiosa device plugin implements the [Kubernetes Device Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) interface for FuriosaAI NPU devices, and its features are as follows:
* Discovering the Furiosa NPU devices and registering to a Kubernetes cluster.
* Tracking the health of the devices and reporting to a Kubernetes cluster.
* Running AI workload on the top of the Furiosa NPU devices within a Kubernetes cluster.

## Configuration

The configuration structure is as follows:
```yaml
config:
  partitioning: none
  debugMode: false
```

### Device Partitioning

The Furiosa NPU can be integrated into the Kubernetes cluster in various configurations.
A single NPU card can either be exposed as a single resource or partitioned into multiple resources.
`partitioning` determines the resource unit of partitioned resource and name of the resource.

The following table shows the available partitioning configurations and expected resource names:

> [!NOTE]  
> Only RNGD architecture supports partitioning.
> If the partitioning configuration is configured on the Warboy, the device plugin will not start.

| NPU Configuration | Resource Name                | Resource Count Per Card |
|-------------------|------------------------------|-------------------------|
| `none`            | `furiosa.ai/rngd`            | 1                       |
| `2core.12gb`      | `furiosa.ai/rngd-2core.12gb` | 4                       |
| `4core.24gb`      | `furiosa.ai/rngd-4core.24gb` | 2                       |

### Debug Mode

`debugMode` enables or disables debug mode. The default value is `false`.

### Disable Device (Experimental)

`disabledDeviceUUIDs` allows disabling specific devices on a per-node basis.
The UUID of the device can be obtained using the [furiosa-smi info](https://developer.furiosa.ai/latest/en/device_management/system_management_interface/furiosa_smi_cli.html#furiosasmicli) command.

The following is an example configuration:

```yaml
config:
  partitioning: none
  debugMode: false
  disabledDeviceUUIDs:
    node_a:
      - "uuid1"
      - "uuid2"
    node_b:
      - "uuid3"
      - "uuid4"
```

## Deploying Furiosa Device Plugin with Helm

The Furiosa device plugin helm chart is available at [furiosa-ai/helm-charts](https://github.com/furiosa-ai/helm-charts).
To configure deployment as you need, you can modify charts/furiosa-device-plugin/values.yaml.
* If partitioning is not specified, the default value is `"none"`.
* If debugMode is not specified, the default value is `false`.
* If disabledDeviceUUIDs is not specified, the default value is empty list `[]`.

You can deploy the Furiosa Device Plugin by running the following commands:

```shell
helm repo add furiosa https://furiosa-ai.github.io/helm-charts
helm repo update
helm install furiosa-device-plugin furiosa/furiosa-device-plugin -n kube-system
```
