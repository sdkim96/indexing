package input

import "context"

type Provider interface {
	Provide(ctx context.Context, sourceID string) (Input, error)
}
