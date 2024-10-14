## Requirements

The [`aws-provider.yaml`](./aws-provider.yaml) file must be instantiated as a part of the `crossplane` application
kustomization rather than in the `aws` sub-application.

Otherwise, a race condition will persist relating to the `aws.upbound.io/v1beta1` CRDs. 
