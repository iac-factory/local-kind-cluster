## Prerequisites

A Kubernetes Secret must be created with the following specification:

- Name: `aws-crossplane-credentials`
- Namespace: `flux-system`
- Type: `Opaque`
- Fields:
    - `profile`: base64-value

## References

- [Provider Marketplace](https://marketplace.upbound.io)
- [AWS Provider(s)](https://marketplace.upbound.io/providers/upbound/provider-family-aws/latest/providers)
