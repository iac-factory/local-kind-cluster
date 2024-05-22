package installation

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func deserialize(data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	if e := apiextensionsv1.AddToScheme(scheme.Scheme); e != nil {
		return nil, nil, nil
	}

	if e := apiextensionsv1beta1.AddToScheme(scheme.Scheme); e != nil {
		return nil, nil, nil
	}

	decoder := scheme.Codecs.UniversalDeserializer()

	return decoder.Decode(data, nil, nil)
}
