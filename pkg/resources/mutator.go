package resources

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Mutator[K client.Object] interface {
	Empty() K
	Mutate(res K) error
	String() string
	MetadataMutator() MetadataMutator
}

func GetResource[K client.Object](ctx context.Context, clt client.Client, m Mutator[K], opts ...client.GetOption) (K, error) {
	res := m.Empty()
	if err := clt.Get(ctx, client.ObjectKeyFromObject(res), res, opts...); err != nil {
		return res, fmt.Errorf("failed to get %s: %w", m.String(), err)
	}
	return res, nil
}

func CreateOrUpdateResource[K client.Object](ctx context.Context, clt client.Client, m Mutator[K]) error {
	res := m.Empty()
	_, err := controllerutil.CreateOrUpdate(ctx, clt, res, func() error {
		return m.Mutate(res)
	})
	if err != nil {
		return fmt.Errorf("failed to create or update %s: %w", m.String(), err)
	}
	return nil
}

func DeleteResource[K client.Object](ctx context.Context, clt client.Client, m Mutator[K], opts ...client.DeleteOption) error {
	res := m.Empty()
	if err := clt.Delete(ctx, res, opts...); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("failed to delete %s: %w", m.String(), err)
	}
	return nil
}
