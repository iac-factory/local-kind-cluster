package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

//go:embed crd.yaml
var manifest []byte

func main() {
	// Load Kubernetes configuration
	kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// Create a dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %s", err.Error())
	}

	// Parse the YAML file to JSON
	jsonData, err := yaml.ToJSON(manifest)
	if err != nil {
		log.Fatalf("Error converting YAML to JSON: %s", err.Error())
	}

	// Decode the JSON into an unstructured.Unstructured object
	var crd unstructured.Unstructured
	if err := crd.UnmarshalJSON(jsonData); err != nil {
		log.Fatalf("Error unmarshaling JSON into unstructured: %s", err.Error())
	}

	fmt.Println(crd.GroupVersionKind().Group, crd.GroupVersionKind().Version, crd.GetKind())

	var resource = crd.GetKind()
	if crd.GroupVersionKind().GroupVersion() == apiextensionsv1.SchemeGroupVersion {
		resource = "customresourcedefinitions"
	}

	// Get the GVR (GroupVersionResource) for CRDs
	gvr := schema.GroupVersionResource{
		Group:    crd.GroupVersionKind().Group,
		Version:  crd.GroupVersionKind().Version,
		Resource: resource,
	}

	// Apply the CRD to the cluster
	_, err = dynamicClient.Resource(gvr).Create(context.TODO(), &crd, v1.CreateOptions{})
	if err != nil {
		log.Printf("Error creating CRD: %s\n", err.Error())
	}

	fmt.Println(crd.GetName())

	err = dynamicClient.Resource(gvr).Delete(context.TODO(), crd.GetName(), v1.DeleteOptions{})
	if err != nil {
		log.Fatalf("Error creating CRD: %s", err.Error())
	}

	fmt.Println("CRD created successfully")
}
