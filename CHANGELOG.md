# Changelog

All notable changes to the Liqo Security Engine will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] ([0.1.0] - Initial Development)

### Added

- Comprehensive code documentation for all Go source files
- Professional README with architecture overview and examples
- CONTRIBUTING.md with development guidelines
- CODE_OF_CONDUCT.md following Contributor Covenant 2.1
- Apache 2.0 LICENSE file
- Enhanced examples with detailed comments and use cases
- Examples README with usage scenarios and best practices
- Additional example configurations for various deployment patterns
- PeeringConnectivity Custom Resource Definition (CRD)
- Controller for reconciling PeeringConnectivity resources
- Support for five resource groups:
  - local-cluster: Local cluster pod CIDR
  - remote-cluster: Remote cluster pod CIDR
  - offloaded: Pods offloaded from consumer to provider
  - vc-local: Local pods in offloaded namespaces
  - vc-remote: Shadow pods on consumer cluster
- Automatic FirewallConfiguration generation
- Dynamic pod tracking and firewall rule updates
- Watches for Pods, Networks, and NamespaceOffloadings
- Allow and deny actions for traffic rules
- Integration with Liqo fabric network
- Support for nftables-based firewall rules
- Status conditions and event recording
- RBAC configurations
- Helm chart for deployment
- Combined install manifest (install.yaml)
- Basic examples for consumer and provider clusters
- Unit and integration tests
- End-to-end test framework

### Changed

- Improved code comments throughout the codebase
- Enhanced API documentation with detailed field descriptions
- Updated package-level documentation for all packages

### Documentation

- Added comprehensive README with:
  - Project overview and key features
  - Architecture diagram
  - Installation instructions (Helm and kubectl)
  - Quick start guide
  - Resource groups reference table
  - Multiple examples
  - API reference
  - Troubleshooting guide
  - Development guide
  - Project structure
  - Roadmap
- Created examples for:
  - Consumer cluster configuration
  - Provider cluster configuration
  - Isolated cluster setup
  - Selective access control
  - Open policy for development
  - Multi-tenant provider setup

### Technical Details

- Built with Kubebuilder v4.10.1
- Go 1.24.6 support
- Kubernetes 1.11.3+ compatibility
- Integration with Liqo networking APIs:
  - networkingv1beta1.FirewallConfiguration
  - ipamv1alpha1.Network
  - offloadingv1beta1.NamespaceOffloading

[Unreleased]: https://github.com/riccardotornesello/liqo-security-engine/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/riccardotornesello/liqo-security-engine/releases/tag/v0.1.0
