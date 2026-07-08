package health

import (
	"context"

	"github.com/raifmirza/consistent-hash/internal/node"
)

type Prober interface {
	Probe(ctx context.Context, node *node.Node) error
}
