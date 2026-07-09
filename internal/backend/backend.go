package backend

import (
	"net/http/httputil"

	"github.com/raifmirza/consistent-hash/internal/node"
)

type Backend struct {
	Node  *node.Node
	Proxy *httputil.ReverseProxy
}
