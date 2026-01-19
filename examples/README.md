# Examples

This directory contains example PeeringConnectivity configurations for various use cases.

## Quick Reference

| Example                        | Use Case         | Description                                                    |
| ------------------------------ | ---------------- | -------------------------------------------------------------- |
| [consumer.yaml](consumer.yaml) | Consumer cluster | Allow shadow pods, deny direct remote access                   |
| [provider.yaml](provider.yaml) | Provider cluster | Isolate offloaded workloads, allow bidirectional with consumer |

## Understanding Resource Groups

Before using these examples, understand the resource groups:

- **local-cluster**: Pods in the local cluster's pod CIDR
- **remote-cluster**: Pods in the remote cluster's pod CIDR
- **offloaded**: Pods that were offloaded from consumer to provider
- **vc-local**: Local pods in namespaces with offloading enabled
- **vc-remote**: Shadow pods representing offloaded workloads

## Basic Usage

1. Choose an example that matches your use case
2. Replace the cluster ID and namespace with your values:
   ```yaml
   metadata:
     name: <your-cluster-id>
     namespace: liqo-tenant-<your-cluster-id>
   ```
3. Apply the configuration:
   ```bash
   kubectl apply -f <example-file>.yaml
   ```
4. Verify the status:
   ```bash
   kubectl get peeringconnectivity -n liqo-tenant-<your-cluster-id>
   kubectl describe peeringconnectivity <your-cluster-id> -n liqo-tenant-<your-cluster-id>
   ```

## Example Scenario: Consumer-Provider Setup

**Consumer side** (consumer.yaml):

- Allows shadow pods to communicate freely
- Denies direct access from remote cluster

**Provider side** (provider.yaml):

- Allows bidirectional communication between offloaded pods and remote cluster
- Denies offloaded pods from accessing provider resources
- Denies remote cluster from accessing provider resources

## Customizing Examples

### Adding Granular Rules

You can combine multiple rules for fine-grained control:

```yaml
spec:
  rules:
    # Allow specific source to specific destination
    - source:
        group: vc-local
      destination:
        group: offloaded
      action: "allow"

    # Allow specific source to any destination
    - source:
        group: vc-remote
      action: "allow"

    # Deny traffic to specific destination
    - destination:
        group: local-cluster
      action: "deny"
```

### Rule Evaluation Order

Rules are evaluated in order from top to bottom. The first matching rule determines the action:

```yaml
spec:
  rules:
    # More specific rules should come first
    - source:
        group: vc-remote
      destination:
        group: remote-cluster
      action: "allow"

    # More general rules should come later
    - source:
        group: vc-remote
      action: "deny"
```

## Testing Your Configuration

After applying a PeeringConnectivity resource:

1. **Check the resource status**:

   ```bash
   kubectl get peeringconnectivity -A
   ```

2. **Verify FirewallConfiguration was created**:

   ```bash
   kubectl get firewallconfiguration -A
   ```

3. **Test connectivity**:

   ```bash
   # Deploy test pods in different resource groups
   # Try to communicate between them
   # Verify that allowed traffic works and denied traffic fails
   ```

4. **Check controller logs**:
   ```bash
   kubectl logs -n liqo-system -l app=liqo-security-engine
   ```

## Troubleshooting

### Rules Not Taking Effect

1. Verify the namespace matches the pattern `liqo-tenant-<cluster-id>`
2. Check the PeeringConnectivity status for errors
3. Verify the FirewallConfiguration was created
4. Check controller logs for reconciliation errors

### Unexpected Traffic Blocking

1. Review rule order (first match wins)
2. Check if a deny rule is too broad
3. Verify resource group memberships (check pod labels)
4. Look for established connection exceptions

### Unable to Determine Resource Group

Some pods may not fall into the expected resource groups:

- Check pod labels (liqo.io/local-pod, liqo.io/origin-cluster-id)
- Verify namespace has NamespaceOffloading resource (for vc-local)
- Check pod scheduling node (for vc-remote)

## Best Practices

1. **Start Restrictive**: Begin with deny-all and selectively allow
2. **Test Thoroughly**: Test all use cases before production deployment
3. **Document Intent**: Add comments explaining why each rule exists
4. **Regular Review**: Periodically review and update policies
5. **Monitor Impact**: Watch metrics and logs after policy changes
6. **Version Control**: Keep policies in git with change history

## Additional Resources

- [Main README](../README.md) - Project overview and getting started
- [API Reference](../README.md#api-reference) - Complete API documentation
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute
- [Liqo Documentation](https://docs.liqo.io) - Learn more about Liqo
