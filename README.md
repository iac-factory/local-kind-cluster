# local-kind-cluster

Local Kubernetes Cluster(s) via Kind

## Contribution

Please see the [**Contributing Guide**](./CONTRIBUTING.md) file for additional details.

## Build Attestations

Build attestations are attached to the final image as metadata.

The purpose of attestations is to make it possible to inspect an image and see where it comes from, who created it and how, and what it contains.

Such a concept enables the use of policy engines for validating images based on policy rules.

Two types of build annotations are available:

- **Software Bill of Material (SBOM)**: list of software artifacts that an image contains, or that were used to build the image.
- **Provenance**: how an image was built.

### Including Attestations with Docker

```bash
docker buildx build --sbom=true --provenance=true .
```

## Multi-Cluster GitOps
https://github.com/aws-samples/eks-multi-cluster-gitops
https://aws.amazon.com/blogs/containers/part-1-build-multi-cluster-gitops-using-amazon-eks-flux-cd-and-crossplane/

The local cluster acts as the management cluster.
