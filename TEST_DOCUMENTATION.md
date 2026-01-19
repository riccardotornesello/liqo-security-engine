# Test Documentation

This document describes the comprehensive test suite implemented for the Liqo Security Engine.

## Overview

The test suite includes:
- **Unit Tests**: Testing individual utility functions and components
- **Controller Tests**: Testing the PeeringConnectivity controller logic
- **Integration Tests**: Testing the complete workflow with Kubernetes resources
- **E2E Tests**: End-to-end tests validating the entire system in a real Kubernetes cluster

## Running Tests

### Unit Tests

Run all unit tests (utils package):
```bash
go test ./internal/controller/utils/... -v
```

Expected output: All 17 specs should pass successfully.

### Controller Tests

Note: Controller tests require liqo CRDs to be installed in the test environment.
```bash
make test
```

The controller tests validate:
- PeeringConnectivity resource creation
- FirewallConfiguration generation
- Status updates
- Error handling
- Resource updates

### Integration Tests

Integration tests are located in `test/integration/` and require:
- A test Kubernetes cluster (envtest)
- Liqo CRDs installed
- The controller running

Run integration tests:
```bash
go test ./test/integration/... -v
```

These tests validate:
- Complete reconciliation workflow
- Multiple PeeringConnectivity resources
- Dynamic updates
- Owner reference cleanup
- Multiple firewall rules

### E2E Tests

E2E tests are located in `test/e2e/` and require:
- A real Kubernetes cluster (Kind recommended)
- Liqo installed
- The controller deployed

Run e2e tests:
```bash
make test-e2e
```

These tests validate:
- Real cluster deployment
- PeeringConnectivity creation and updates
- Metrics availability
- Complete system behavior

## Test Coverage

### Unit Tests (internal/controller/utils/)

#### common_test.go
- ✅ `GetClusterNamespace` - Tests namespace generation from cluster ID
  - Standard cluster IDs
  - Short cluster IDs
  - Cluster IDs with hyphens
- ✅ `ExtractClusterID` - Tests cluster ID extraction from namespace
  - Valid formats
  - Invalid formats
  - Edge cases (empty, partial matches)

#### Sets Testing (included in common_test.go)
- ✅ `ForgePodIpsSet` - Tests firewall set creation from pod IPs
  - Multiple pods with IPs
  - Empty pod lists
  - Pods without IP addresses
  - IPv6 addresses
  - Special characters in set names

### Controller Tests (internal/controller/)

#### peeringsecurity_controller_test.go
- ✅ Basic resource creation with allow rules
- ✅ Resource creation with deny rules
- ✅ Empty rules handling
- ✅ Invalid namespace format error handling
- ✅ Resource updates and FirewallConfiguration synchronization
- ✅ Status condition updates
- ✅ FirewallConfiguration lifecycle management

### Integration Tests (test/integration/)

#### integration_test.go
- ✅ FirewallConfiguration creation with allow rules
- ✅ FirewallConfiguration creation with deny rules
- ✅ Resource updates trigger FirewallConfiguration updates
- ✅ Multiple rules handling
- ✅ Owner reference and cleanup on deletion

### E2E Tests (test/e2e/)

#### e2e_test.go
- ✅ Controller deployment and health
- ✅ Metrics endpoint availability
- ✅ PeeringConnectivity creation and reconciliation
- ✅ FirewallConfiguration generation
- ✅ Resource updates
- ✅ Status verification
- ✅ Metrics verification

## Test Infrastructure

### Dependencies Added to Local Liqo

To enable testing with the development version of liqo, the following types were added to the local liqo repository (`../liqo`):

1. **Set Types** (`apis/networking/v1beta1/firewall/set_types.go`):
   - `SetDataType` - Enum for set element types
   - `SetElement` - Individual set elements
   - `Set` - Named collection of elements

2. **Filter Actions** (`apis/networking/v1beta1/firewall/filterrule_types.go`):
   - `ActionAccept` - Allow packets
   - `ActionDrop` - Drop packets

3. **Match Types** (`apis/networking/v1beta1/firewall/match_types.go`):
   - `CtStateValue` - Connection tracking states
   - `MatchCtState` - Connection tracking match
   - Added `CtState` field to `Match` struct

4. **Table Types** (`apis/networking/v1beta1/firewall/table_types.go`):
   - Added `Sets` field to `Table` struct

### Test Utilities

- **envtest**: Used for controller and integration tests
- **Ginkgo/Gomega**: BDD-style testing framework
- **Kind**: Kubernetes in Docker for e2e tests

## Known Limitations

1. **Controller and Integration Tests**: Require liqo CRDs to be available. Currently, these tests are designed for environments where liqo is fully installed.

2. **Local Liqo Dependency**: The project uses a local liqo repository (`../liqo`) with custom modifications to support features not yet in the released version.

3. **Test Isolation**: Some tests create namespaces and resources that require cleanup. Ensure proper cleanup between test runs.

## Future Improvements

1. **Pod Utilities Tests**: Add tests for:
   - `GetPodsOffloadedToProvider`
   - `GetPodsFromConsumer`
   - `GetPodsInOffloadedNamespaces`

2. **Resource Groups Tests**: Add tests for all resource group functions

3. **Forge Package Tests**: Add tests for FirewallConfiguration creation logic

4. **CIDR Utilities Tests**: Add tests for:
   - `GetCurrentClusterPodCIDR`
   - `GetRemoteClusterPodCIDR`

5. **Mock Support**: Add better mocking for Kubernetes resources to enable more isolated unit tests

## Contributing

When adding new tests:
1. Follow the existing test patterns (Ginkgo BDD style)
2. Ensure tests are isolated and don't depend on external state
3. Add documentation to this file
4. Run tests locally before submitting

## Questions or Issues

For questions about the test suite or to report issues, please open an issue on the GitHub repository.
