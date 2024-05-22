## Usage

```bash
kubectl apply --kustomize ./overlays/development
```

## Namespaces

- https://kubernetes.io/docs/tasks/administer-cluster/namespaces

### [Understanding Namespaces and DNS](https://kubernetes.io/docs/tasks/administer-cluster/namespaces/#understanding-namespaces-and-dns)

When you create a Service, it creates a corresponding DNS entry. This entry is of the form
`<service-name>.<namespace-name>.svc.cluster.local`, which means that if a container uses `<service-name>`
it will resolve to the service which is local to a namespace. This is useful for using the same
configuration across multiple namespaces such as `Development`, `Staging` and `Production`. If you want
to reach across namespaces, you need to use the fully qualified domain name (FQDN).

Kubernetes annotations and labels are both key-value pairs associated with Kubernetes objects, such as pods and
deployments. However, they serve different purposes and have different uses within a Kubernetes environment.

## Labels vs Annotations

### Labels

- **Purpose**: Labels are used to organize, select, and group Kubernetes objects. They are the primary mechanism for
  querying and performing actions on a set of objects.
- **Use Cases**:
    - **Filtering resources** during queries with `kubectl` or within client libraries.
    - **Specifying constraints** and requirements for scheduling pods on nodes.
    - **Service Discovery**: Labels can be used by a `Service` to select a group of pods to expose through a common
      interface.
    - **Load Balancing**: Distributing network traffic to multiple instances of an application.

### Annotations

- **Purpose**: Annotations are used to store additional, non-identifying information that can be used by tools and
  libraries. They are not used to identify and select objects.
- **Use Cases**:
    - **Storing build/release versions**, URLs, release notes, or other information about the object that doesn't fit
      into labels.
    - **Management tools**: Annotations can provide hints to tools about how to behave or store metadata that is not
      used for selection purposes.
    - **Integrations and extensions**: Annotations are often used by extensions and third-party tools to store
      configuration specifics or metadata.

### Key Differences

- **Selection**: Labels can be used in selectors (e.g., in a `Deployment` to select a set of pods), whereas annotations
  cannot.
- **Purpose and Use**: Labels are primarily for organizing and finding objects. Annotations are for storing additional
  information that doesn't fit into labels.
- **Syntax and Constraints**: Both use key-value pairs, but labels have stricter naming constraints because they are
  intended for use in selectors.

In summary, while both labels and annotations are key-value pairs attached to Kubernetes objects, labels are designed
for identifying and organizing resources, whereas annotations are meant for storing additional, non-identifying
information that can be used by tools and libraries.
