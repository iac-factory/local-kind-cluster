package installation

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"installer.operators.ethr.gg/installer/api/v1alpha1"
)

func helm(ctx context.Context, object *v1alpha1.Helm) error {
	return field.InternalError(nil, errors.New("implementation not available"))
}
