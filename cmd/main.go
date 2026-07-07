package main

import (
	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

func main() {
	ring := hash.NewHashRing(5)

	ring.AddNode(&node.Node{
		ID: "Server-1",
	})

	ring.AddNode(&node.Node{
		ID: "Server-2",
	})

	ring.AddNode(&node.Node{
		ID: "Server-3",
	})

}
