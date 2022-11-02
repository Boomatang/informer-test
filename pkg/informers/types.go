package informers

import "context"

type ResourceInformer interface {
	Run(ctx context.Context) error
	Sync()
}
