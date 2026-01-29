# Liqo Connectivity Engine

[![Go Version](https://img.shields.io/github/go-mod/go-version/riccardotornesello/liqo-connectivity-engine)](go.mod)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

The Liqo Connectivity Engine is a Kubernetes operator that provides fine-grained network security policies for [Liqo](https://liqo.io) multi-cluster environments. It enables administrators to control network traffic between different resource groups in peered clusters, offering security controls for workload offloading and multi-cluster scenarios.

## Overview

In a Liqo multi-cluster deployment, workloads can be offloaded from one cluster (consumer) to another (provider). The Liqo Connectivity Engine allows you to define security policies that control network connectivity between:

- **Local cluster pods**: Pods running in the local cluster's own pod CIDR
- **Remote cluster pods**: Pods running in a remote peered cluster's pod CIDR
- **Offloaded pods**: Workloads that have been offloaded from a consumer to a provider cluster
- **Virtual cluster pods**: Local and remote pods in namespaces configured for offloading

The connectivity engine translates high-level security rules into low-level firewall configurations that are applied to the Liqo fabric network using nftables.

## Key Features

- **Fine-grained traffic control**: Define allow/deny rules for traffic between different resource groups
- **Dynamic pod tracking**: Automatically tracks and updates firewall rules as pods are created, deleted, or offloaded
- **Multi-cluster aware**: Understands Liqo's multi-cluster topology and virtual cluster concepts
- **Kubernetes-native**: Uses Custom Resource Definitions (CRDs) for policy configuration
- **Integration with Liqo**: Seamlessly integrates with Liqo's networking stack

## Architecture

The Liqo Connectivity Engine consists of:

1. **PeeringConnectivity CRD**: Defines security policies for a peered cluster
2. **Controller**: Watches PeeringConnectivity resources and translates them into FirewallConfiguration resources
3. **Resource watchers**: Monitors Pods, Networks, and NamespaceOffloadings to keep firewall rules up-to-date

```
┌─────────────────────────────────────────────────────────────┐
│                    Liqo Connectivity Engine                 │
│                                                             │
│  ┌────────────────┐      ┌──────────────────────────┐       │
│  │ PeeringConn    │─────>│ Controller               │       │
│  │ ectivity CRD   │      │                          │       │
│  └────────────────┘      │ - Reconciles policies    │       │
│                          │ - Watches resources      │       │
│                          │ - Creates FirewallConfig │       │
│                          └──────────────────────────┘       │
│                                    │                        │
└────────────────────────────────────┼────────────────────────┘
                                     │
                                     ▼
                          ┌─────────────────────┐
                          │ Liqo Fabric Network │
                          │ (nftables rules)    │
                          └─────────────────────┘
```

## Getting Started

### Prerequisites

- **Kubernetes cluster**: version 1.30.0 or higher
- **Liqo**: version 1.1.0 or higher

### Installation

#### Using Helm Chart

The easiest way to install the Liqo Connectivity Engine is using Helm:

```bash
helm install liqo-connectivity-engine ./dist/chart \
  --namespace liqo-system \
  --create-namespace
```

#### Using kubectl

Alternatively, you can install using kubectl:

```bash
kubectl apply -f https://raw.githubusercontent.com/riccardotornesello/liqo-connectivity-engine/main/dist/install.yaml
```

### Quick Start

1. **Create a PeeringConnectivity resource** for your peered cluster:

```yaml
apiVersion: security.liqo.io/v1
kind: PeeringConnectivity
metadata:
  name: remote-cluster-id
  namespace: liqo-tenant-remote-cluster-id
spec:
  rules:
    # Allow traffic from remote cluster pods to offloaded pods
    - source:
        group: remote-cluster
      destination:
        group: offloaded
      action: allow

    # Allow offloaded pods to communicate back to remote cluster
    - source:
        group: offloaded
      destination:
        group: remote-cluster
      action: allow

    # Deny all other traffic from offloaded pods
    - source:
        group: offloaded
      action: deny
```

2. **Apply the configuration**:

```bash
kubectl apply -f peeringconnectivity.yaml
```

3. **Verify the status**:

```bash
kubectl get peeringconnectivity -n liqo-tenant-remote-cluster-id
kubectl describe peeringconnectivity remote-cluster-id -n liqo-tenant-remote-cluster-id
```

More examples can be found in the [examples/](examples/) directory.

## Resource Groups

The following resource groups can be used in security rules:

| Group             | Description                                      | Usage Scenario                               |
| ----------------- | ------------------------------------------------ | -------------------------------------------- |
| `local-cluster`   | Pods in the local cluster's pod CIDR             | Restrict access from local workloads         |
| `remote-cluster`  | Pods in the remote cluster's pod CIDR            | Control access from remote cluster           |
| `offloaded`       | Pods offloaded from consumer to provider         | Isolate offloaded workloads on provider      |
| `vc-local`        | Local pods in namespaces with offloading enabled | Control access to potentially offloaded pods |
| `vc-remote`       | Shadow pods representing offloaded workloads     | Manage traffic to remote offloaded pods      |
| `private-subnets` | Private subnet IPs according to RFC1918          | Restrict access to private networks          |

## Examples

The examples and their description can be found in the [examples/](examples/) directory.

## Development

### Building from Source

1. **Clone the repository**:

```bash
git clone https://github.com/riccardotornesello/liqo-connectivity-engine.git
cd liqo-connectivity-engine
```

2. **Build the binary**:

```bash
make build
```

3. **Run tests**:

```bash
make test
```

4. **Build and push the Docker image**:

```bash
make docker-build docker-push IMG=<your-registry>/liqo-connectivity-engine:tag
```

### Running Locally

To run the controller locally against a Kubernetes cluster:

```bash
make install  # Install CRDs
make run      # Run controller locally
```

## API Reference

### PeeringConnectivity

The `PeeringConnectivity` custom resource defines security policies for a peered cluster.

#### Spec

| Field   | Type     | Required | Description                    |
| ------- | -------- | -------- | ------------------------------ |
| `rules` | `[]Rule` | No       | Ordered list of security rules |

#### Rule

| Field         | Type     | Required | Description                                             |
| ------------- | -------- | -------- | ------------------------------------------------------- |
| `action`      | `string` | No       | Action to take: `allow` or `deny`                       |
| `source`      | `Party`  | No       | Source party (if omitted, matches any source)           |
| `destination` | `Party`  | No       | Destination party (if omitted, matches any destination) |

#### Party

| Field   | Type     | Required | Description                                               |
| ------- | -------- | -------- | --------------------------------------------------------- |
| `group` | `string` | No       | Resource group, as defined in the Resource Groups section |

#### Status

| Field                | Type          | Description              |
| -------------------- | ------------- | ------------------------ |
| `conditions`         | `[]Condition` | Current state conditions |
| `observedGeneration` | `int64`       | Last observed generation |

## Troubleshooting

### PeeringConnectivity not taking effect

1. Check the status of the PeeringConnectivity resource:

   ```bash
   kubectl describe peeringconnectivity <name> -n <namespace>
   ```

2. Check the FirewallConfiguration was created:

   ```bash
   kubectl get firewallconfiguration -n <namespace>
   ```

3. Check controller logs:
   ```bash
   kubectl logs -n liqo-system -l app=liqo-connectivity-engine
   ```

### Unable to extract cluster ID

Ensure the PeeringConnectivity is created in the correct namespace. The namespace must follow the pattern: `liqo-tenant-<cluster-id>`.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on:

- Code of conduct
- Development workflow
- Submitting pull requests
- Reporting issues

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/riccardotornesello/liqo-connectivity-engine/issues)
- **Discussions**: [GitHub Discussions](https://github.com/riccardotornesello/liqo-connectivity-engine/discussions)
- **Liqo Documentation**: [https://docs.liqo.io](https://docs.liqo.io)

## Acknowledgments

This project is built on top of [Liqo](https://liqo.io), an open-source project that enables dynamic and seamless Kubernetes multi-cluster topologies.
