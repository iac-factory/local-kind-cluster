package installation

import (
	"bytes"
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	yamlstandard "gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	k8sresource "k8s.io/cli-runtime/pkg/resource"

	"installer.operators.ethr.gg/installer/api/v1alpha1"
	"installer.operators.ethr.gg/installer/internal/configuration"
)

// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#Client

func manifest(ctx context.Context, object *v1alpha1.Manifest) error {
	var version = os.Getenv("VERSION")

	exceptions := make([]*field.Error, 0)

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

	var buffer bytes.Buffer
	if _, e := io.Copy(&buffer, response.Body); e != nil {
		e = fmt.Errorf("unable to read http response body: %w", e)

		return field.InternalError(nil, e)
	}

	settings, e := configuration.Configuration()
	if e != nil {
		return e
	}

	dynamics, e := dynamic.NewForConfig(settings)
	if e != nil {
		e = fmt.Errorf("failed to create dynamic client: %v", e)
		return field.InternalError(nil, e)
	}

	clientset, e := kubernetes.NewForConfig(settings)
	if e != nil {
		e = fmt.Errorf("failed to create clientset: %v", e)
		return field.InternalError(nil, e)
	}

	grs, e := restmapper.GetAPIGroupResources(clientset.Discovery())
	if e != nil {
		e = fmt.Errorf("failed to instantiate rest-mapper: %v", e)
		return field.InternalError(nil, e)
	}

	rm := restmapper.NewDiscoveryRESTMapper(grs)

	/***
	The dynamic client in client-go can deal with both unstructured.Unstructured objects; runtime.Object can be converted
	to unstructured.Unstructured objects.

	https://127.0.0.1:65300/apis/apiextensions.k8s.io/v1/customresourcedefinitions
	*/

	// Assuming the generated YAML is for multiple resources, split and apply each

	var resources = make([][]byte, 0)

	d := yamlstandard.NewDecoder(&buffer)
	for {
		var specification interface{}

		e := d.Decode(&specification)

		if stderrors.Is(e, io.EOF) {
			break
		} else if specification == nil {
			continue
		}

		if e != nil {
			panic(e)
		}

		output, e := yamlstandard.Marshal(specification)
		if e != nil {
			panic(e)
		}

		resources = append(resources, output)
	}

	total := len(resources)

	object.SetTotal(total) // --> update status when total amount of resource(s) is known

	const crd = "customresourcedefinitions"

	for index, resource := range resources {
		if len(resource) == 0 {
			continue
		}

		content, e := yaml.ToJSON(resource)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Convert YAML to JSON", slog.String("error", e.Error()))
			e = fmt.Errorf("error converting yaml to json: %w", e)
			return field.InternalError(nil, e)
		}

		slog.DebugContext(ctx, "Successfully Converted YAML To JSON")

		var raw unstructured.Unstructured
		if e := raw.UnmarshalJSON(content); e != nil {
			slog.ErrorContext(ctx, "Unable to Decode Resource into Unstructured K8s Type", slog.String("error", e.Error()))
			e = fmt.Errorf("unable to decode resource into unstructured: %w", e)
			return field.InternalError(nil, e)
		}

		iterator := slog.Group(raw.GetName(), slog.Int("index", index+1), slog.Int("total", total))

		slog.DebugContext(ctx, "Successfully Unmarshalled JSON into Unstructured Resource", iterator, slog.Group(raw.GetName(), slog.String("namespace", raw.GetNamespace()), slog.String("kind", raw.GetKind()), slog.String("group", raw.GroupVersionKind().Group)))

		var kind = raw.GetKind()
		switch raw.GroupVersionKind().GroupVersion() {
		case apiextensionsv1.SchemeGroupVersion: // !!! => Do not update CRDs in the case of existence
			kind = crd

			slog.DebugContext(ctx, "Evaluating a CRD", iterator, slog.String("name", raw.GetName()), slog.String("group", raw.GroupVersionKind().Group))

			gvr := schema.GroupVersionResource{Group: raw.GroupVersionKind().Group, Version: raw.GroupVersionKind().Version, Resource: kind}

			pointer, e := dynamics.Resource(gvr).Get(ctx, raw.GetName(), metav1.GetOptions{})
			if e != nil && errors.IsNotFound(e) {
				slog.InfoContext(ctx, "Existing CRD Not Found - Creating")

				pointer, e = dynamics.Resource(gvr).Create(ctx, &raw, metav1.CreateOptions{})
				if e != nil {
					e = fmt.Errorf("unable to create crd: %v", e)

					return field.InternalError(nil, e)
				}
			}

			// @TODO implement a CRD drift status check
			// if !(maps.Equal(raw.Object, pointer.Object)) {
			//
			// }

			raw = *(pointer)

			slog.DebugContext(ctx, "Successfully Evaluated Custom-Resource-Definition (CRD)", iterator)
		default:
			slog.DebugContext(ctx, "Evaluating a Resource", iterator, slog.String("name", raw.GetName()), slog.String("group", raw.GroupVersionKind().Group), slog.String("kind", raw.GetKind()))

			singleton, gvk, e := deserialize(resource)
			if e != nil {
				slog.ErrorContext(ctx, "Unable to Deserialize Manifest into Runtime.Object", iterator, slog.String("error", e.Error()))
				e = fmt.Errorf("unable to deserialize resource into runtime.Object: %w", e)
				return field.InternalError(nil, e)
			}

			gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
			mapping, e := rm.RESTMapping(gk, gvk.Version)
			if e != nil {
				slog.ErrorContext(ctx, "Unable to Establish a Rest-Mapping", iterator, slog.String("error", e.Error()))
				e = fmt.Errorf("unable to establish a restmapping, mapping: %w", e)
				return field.InternalError(nil, e)
			}

			gv := mapping.GroupVersionKind.GroupVersion()

			namespace, e := meta.NewAccessor().Namespace(singleton)
			if e != nil {
				slog.ErrorContext(ctx, "Unable to Call the Namespace Accessor", iterator, slog.String("error", e.Error()))
				e = fmt.Errorf("unable to call namespace accessor: %w", e)
				return field.InternalError(nil, e)
			}

			settings.ContentConfig = k8sresource.UnstructuredPlusDefaultContentConfig()
			settings.GroupVersion = &gv

			if len(gv.Group) == 0 {
				settings.APIPath = "/api"
			} else {
				settings.APIPath = "/apis"
			}

			restclient, e := rest.RESTClientFor(settings)
			if e != nil {
				slog.ErrorContext(ctx, "Unable to Instantiate Customized REST-Client", iterator, slog.String("error", e.Error()))
				e = fmt.Errorf("unable to instantiate customized rest-client: %w", e)
				return field.InternalError(nil, e)
			}

			helper := k8sresource.NewHelper(restclient, mapping)

			singleton, e = helper.Create(namespace, true, singleton)
			if e != nil && errors.IsAlreadyExists(e) {
				slog.DebugContext(ctx, "Resource Already Exists - Skipping", iterator)
				continue
			} else if e != nil {
				slog.ErrorContext(ctx, "Unable to Apply Manifest", iterator)
				e = fmt.Errorf("unable to apply manifest via k8sresource, helper utility: %w", e)
				return field.InternalError(nil, e)
			}

			slog.DebugContext(ctx, "Successfully Evaluated Resource", iterator)
		}

		slog.InfoContext(ctx, "Successfully Applied Manifest", iterator)
	}

	object.SetError(nil)
	object.SetProcessed(true)

	slog.InfoContext(ctx, "Successfully Applied all Kubernetes Manifest(s)", slog.String("version", version), slog.Int("total", len(resources)))

	return nil
}
