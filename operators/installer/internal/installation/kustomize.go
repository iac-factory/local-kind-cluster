package installation

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"installer.operators.ethr.gg/installer/api/v1alpha1"
)

func kustomize(ctx context.Context, object *v1alpha1.Kustomize) error {
	var version = os.Getenv("VERSION")

	exceptions := make([]*field.Error, 0)

	// Step 1: Fetch the Kustomize manifests from the URL
	url := object.Spec.URL
	if url == "" {
		exceptions = append(exceptions, field.Required(field.NewPath("url"), "Invalid URL Format - Empty String"))

		return errors.NewInvalid(object.GroupVersionKind().GroupKind(), object.Name, exceptions)
	}

	client := &http.Client{Timeout: time.Second * 15}
	request, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		e = fmt.Errorf("unable to make http request to %s: %w", url, e)
		exceptions = append(exceptions, field.InternalError(field.NewPath("url"), e))

		return errors.NewInvalid(object.GroupVersionKind().GroupKind(), object.Name, exceptions)
	}

	response, e := client.Do(request)
	if e != nil {
		e = fmt.Errorf("unable to get http response from %s: %w", url, e)
		exceptions = append(exceptions, field.InternalError(field.NewPath("url"), e))

		return errors.NewInvalid(object.GroupVersionKind().GroupKind(), object.Name, exceptions)
	}

	defer response.Body.Close()

	data, e := io.ReadAll(response.Body)
	if e != nil {
		e = fmt.Errorf("unable to read http response body: %w", e)

		return field.InternalError(field.NewPath("url"), e)
	}

	// Step 2: Save the fetched data to a temporary directory
	if e := os.MkdirAll(os.TempDir(), 0o755); e != nil {
		e = fmt.Errorf("unable to generate temporary, root directory: %w", e)

		return field.InternalError(field.NewPath("url"), e)
	}

	temporary, e := os.MkdirTemp(os.TempDir(), "kustomize-")
	if e != nil {
		e = fmt.Errorf("unable to generate temporary directory: %w", e)

		return field.InternalError(field.NewPath("url"), e)
	}

	defer os.RemoveAll(temporary)
	if e := os.WriteFile(filepath.Join(temporary, "kustomization.yaml"), data, 0o644); e != nil {
		e = fmt.Errorf("unable to write kustomization.yaml to temporary directory: %w", e)

		return field.InternalError(field.NewPath("url"), e)
	}

	// Step 3: Run Kustomize to build the manifests
	fs := filesys.MakeFsOnDisk()
	options := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(options)
	m, e := k.Run(fs, temporary)
	if e != nil {
		e = fmt.Errorf("unable to perform kustomization: %w", e)

		return field.InternalError(field.NewPath("url"), e)
	}

	yaml, e := m.AsYaml()
	if e != nil {
		e = fmt.Errorf("failed to convert manifests to YAML: %w", e)

		return field.InternalError(field.NewPath("url"), e)
	}

	configuration, e := rest.InClusterConfig()
	if e != nil {
		home, e := os.UserHomeDir()
		if e != nil {
			e = fmt.Errorf("failed to retrieve user home directory: %w", e)

			return field.InternalError(field.NewPath("url"), e)
		}

		target := filepath.Join(home, ".kube", "config")
		content, e := os.ReadFile(target)
		if e != nil {
			e = fmt.Errorf("unable to read .kube/config file from %s: %w", target, e)

			return field.InternalError(field.NewPath("url"), e)
		}

		clientconfiguration, e := clientcmd.NewClientConfigFromBytes(content)
		if e != nil {
			e = fmt.Errorf("unable to generate configuration from .kube/config file at %s: %w", target, e)

			return field.InternalError(field.NewPath("url"), e)
		}

		configuration, e = clientconfiguration.ClientConfig()
		if e != nil {
			e = fmt.Errorf("unable to generate rest-configuration from .kube/config file at %s: %w", target, e)

			return field.InternalError(field.NewPath("url"), e)
		}
	}

	clientset, e := kubernetes.NewForConfig(configuration)
	if e != nil {
		e = fmt.Errorf("failed to create Kubernetes clientset: %w", e)

		return field.InternalError(field.NewPath("url"), e)
	}

	// Assuming the generated YAML is for multiple resources, split and apply each
	resources := bytes.Split(yaml, []byte("---"))
	for _, resource := range resources {
		if len(resource) == 0 {
			continue
		}

		// Apply each resource (you can add better handling and error checking here)
		// if _, e := clientset.RESTClient().Post().AbsPath("/api/v1/namespaces/default").Body(resource).DoRaw(context.Background()); e != nil {
		if _, e := clientset.RESTClient().Post().Body(resource).DoRaw(ctx); e != nil {
			e = fmt.Errorf("failed to apply resource: %w", e)

			return field.InternalError(field.NewPath("url"), e)
		}
	}

	object.Status.Processed = true

	slog.InfoContext(ctx, "Successfully Applied all Kubernetes Manifest(s) via Kustomize", slog.String("version", version), slog.Int("total", len(resources)))

	return nil
}
