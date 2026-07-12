//go:build minipsql_stub

package engine

import "context"

func New(context.Context, Options) (Cluster, error) {
	return nil, ErrUnavailable
}
