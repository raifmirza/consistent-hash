package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/raifmirza/consistent-hash/internal/node"
)

type HttpProber struct {
	client *http.Client
}

func NewHTTPProber(timeout time.Duration) *HttpProber {
	return &HttpProber{client: &http.Client{
		Timeout: timeout,
	},
	}
}

func (p *HttpProber) Probe(ctx context.Context, n *node.Node) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, n.Address+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
