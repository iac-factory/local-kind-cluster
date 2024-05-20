package installation

import (
	"context"

	"installer.operators.ethr.gg/installer/api/v1alpha1"
)

type Installer interface {
	Helm(ctx context.Context, object *v1alpha1.Helm) error
	Manifest(ctx context.Context, object *v1alpha1.Manifest) error
	Kustomize(ctx context.Context, object *v1alpha1.Kustomize) error
}

type installer struct{}

func (i *installer) Manifest(ctx context.Context, object *v1alpha1.Manifest) error {
	return manifest(ctx, object)
}

func (i *installer) Kustomize(ctx context.Context, object *v1alpha1.Kustomize) error {
	return kustomize(ctx, object)
}

func (i *installer) Helm(ctx context.Context, object *v1alpha1.Helm) error {
	return helm(ctx, object)
}

func New() Installer {
	return &installer{}
}
