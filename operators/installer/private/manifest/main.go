package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"

	yamlstandard "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed manifest.yaml
var manifest []byte

func main() {
	var buffer bytes.Buffer

	buffer.Write(manifest)

	var documents = make([][]byte, 0)

	d := yamlstandard.NewDecoder(bytes.NewReader(manifest))

	var total = 0

	for {
		total++

		var specification interface{}

		e := d.Decode(&specification)

		if errors.Is(e, io.EOF) {
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

		documents = append(documents, output)

		fmt.Println("Parsed Total YAML Document(s):", total)
	}

	for document := range documents {
		fmt.Printf("Converting YAML Document (%d/%d) to JSON", document+1, total)
		resource := documents[document]

		content, e := yaml.ToJSON(resource)
		if e != nil {
			panic(e)
		}

		_ = content
	}
}
