package configuration

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Configuration() (*rest.Config, error) {
	configuration, e := rest.InClusterConfig()
	if e != nil {
		home, e := os.UserHomeDir()
		if e != nil {
			e = fmt.Errorf("failed to retrieve user home directory: %w", e)

			return nil, field.InternalError(nil, e)
		}

		target := filepath.Join(home, ".kube", "config")
		content, e := os.ReadFile(target)
		if e != nil {
			e = fmt.Errorf("unable to read .kube/config file from %s: %w", target, e)

			return nil, field.InternalError(nil, e)
		}

		clientconfiguration, e := clientcmd.NewClientConfigFromBytes(content)
		if e != nil {
			e = fmt.Errorf("unable to generate configuration from .kube/config file at %s: %w", target, e)

			return nil, field.InternalError(nil, e)
		}

		configuration, e = clientconfiguration.ClientConfig()
		if e != nil {
			e = fmt.Errorf("unable to generate rest-configuration from .kube/config file at %s: %w", target, e)

			return nil, field.InternalError(nil, e)
		}
	}

	return configuration, nil
}
