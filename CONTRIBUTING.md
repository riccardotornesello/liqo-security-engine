# Contributing to Liqo Security Engine

Thank you for your interest in contributing to the Liqo Security Engine! We welcome contributions from the community and are grateful for your support.

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** to demonstrate the steps
- **Describe the behavior you observed** and what you expected to see
- **Include logs and error messages** if applicable
- **Specify your environment**:
  - Kubernetes version
  - Liqo version
  - Liqo Security Engine version
  - Operating system and version

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful** to most users
- **List any similar features** in other projects if applicable

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the coding standards below
3. **Add or update tests** as appropriate
4. **Update documentation** to reflect your changes
5. **Ensure the test suite passes**
6. **Make sure your code lints** using `make lint`
7. **Submit your pull request**

#### Pull Request Guidelines

- Fill in the PR template completely
- Link to any related issues
- Update the README.md if you're changing functionality
- Add examples if you're adding new features
- Include tests for new functionality
- Follow the Go coding style
- Keep commits focused and atomic
- Write clear commit messages

## Development Workflow

### Setting Up Your Development Environment

1. **Install prerequisites**:
   - Go 1.24.6 or higher
   - Docker
   - kubectl
   - kind or a Kubernetes cluster
   - Liqo installed on your cluster

2. **Clone the repository**:
   ```bash
   git clone https://github.com/riccardotornesello/liqo-security-engine.git
   cd liqo-security-engine
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Install the CRDs**:
   ```bash
   make install
   ```

### Building and Testing

```bash
# Build the project
make build

# Run unit tests
make test

# Run linter
make lint

# Run end-to-end tests
make test-e2e

# Generate code (after changing API)
make generate

# Generate manifests (after changing API or RBAC)
make manifests
```

### Running Locally

To run the controller locally against your Kubernetes cluster:

```bash
make run
```

This will run the controller using your current kubectl context.

### Testing Your Changes

1. **Unit tests**: Add unit tests for new functionality in `*_test.go` files
2. **Integration tests**: Add integration tests in the `test/` directory
3. **Manual testing**: 
   - Deploy to a test cluster
   - Create test PeeringConnectivity resources
   - Verify FirewallConfiguration resources are created correctly
   - Check that firewall rules work as expected

## Coding Standards

### Go Style Guide

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Use meaningful variable and function names
- Write comments for exported functions and types
- Keep functions small and focused
- Handle errors appropriately

### Code Comments

- Add package-level comments to all packages
- Document all exported types, functions, and constants
- Use complete sentences for comments
- Explain the "why" not just the "what"

Example:
```go
// GetCurrentClusterPodCIDR retrieves the pod CIDR for the current (local) cluster.
// It reads the Network resource in the liqo namespace to obtain the CIDR.
func GetCurrentClusterPodCIDR(ctx context.Context, cl client.Client) (string, error) {
    // implementation
}
```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(controller): add support for IPv6 CIDRs

Add IPv6 support to resource group functions and firewall rule generation.

Closes #123
```

```
fix(forge): prevent nil pointer when destination is omitted

Add nil check for destination party before accessing fields.

Fixes #456
```

## API Changes

When making changes to the API (CRDs):

1. Update the types in `api/v1/`
2. Run `make generate` to update generated code
3. Run `make manifests` to update CRD manifests
4. Update examples to reflect API changes
5. Update documentation
6. Consider backward compatibility

## Documentation

### Updating Documentation

When making changes, update the following as needed:

- **README.md**: High-level overview and quick start
- **API documentation**: Update if changing the API
- **Examples**: Add or update examples in `examples/`
- **Code comments**: Ensure all exported types and functions are documented

### Writing Good Documentation

- Use clear, concise language
- Provide examples
- Explain use cases and scenarios
- Include troubleshooting tips
- Keep it up to date

## Release Process

Releases are managed by the maintainers. The process includes:

1. Update version numbers
2. Update CHANGELOG.md
3. Create a git tag
4. Build and push container images
5. Create GitHub release
6. Update Helm chart

## Getting Help

If you need help:

- Check the [README.md](README.md) and documentation
- Search existing [issues](https://github.com/riccardotornesello/liqo-security-engine/issues)
- Ask questions in [discussions](https://github.com/riccardotornesello/liqo-security-engine/discussions)
- Join the Liqo community channels

## Recognition

Contributors will be recognized in:

- The project's README
- Release notes
- GitHub contributors page

Thank you for contributing to Liqo Security Engine!
