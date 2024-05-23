package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/types"
)

var (
	registry = flag.String("registry", "localhost:5050", "container registry")
	service  = flag.String("service", "", "the service name")
	version  = flag.String("version", "", "version to overwrite kustomization file")
)

func main() {
	flag.Parse()

	if version == nil || *version == "" {
		log.Fatal("--version is required")
	}

	if service == nil || *service == "" {
		log.Fatal("--service is required")
	}

	if registry == nil || *registry == "" {
		log.Fatal("--registry is required")
	}

	cwd, e := os.Getwd()
	if e != nil {
		panic(e)
	}

	target := filepath.Join(cwd, "kustomize", "kustomization.yaml")

	buffer, e := os.ReadFile(target)
	if e != nil {
		panic(e)
	}

	var kustomization types.Kustomization
	if e := yaml.Unmarshal(buffer, &kustomization); e != nil {
		panic(e)
	}

	if len(kustomization.Images) == 0 {
		kustomization.Images = []types.Image{
			{
				Name:    "service:latest",
				NewName: fmt.Sprintf("%s/%s", *registry, *service),
				NewTag:  *(version),
			},
		}
	} else {
		kustomization.Images[0].Name = "service:latest"
		kustomization.Images[0].NewName = fmt.Sprintf("%s/%s", *registry, *service)
		kustomization.Images[0].NewTag = *(version)
	}

	content, e := yaml.Marshal(kustomization)
	if e != nil {
		panic(e)
	}

	if e := os.WriteFile(target, content, 0o644); e != nil {
		panic(e)
	}
}
