package main

import (
	"fmt"

	"github.com/raifmirza/consistent-hash/internal/hash"
	"github.com/raifmirza/consistent-hash/internal/node"
)

func main() {
	ring := hash.NewHashRing(10)

	s1 := &node.Node{ID: "Server-1"}
	s2 := &node.Node{ID: "Server-2"}
	s3 := &node.Node{ID: "Server-3"}

	ring.AddNode(s1)
	ring.AddNode(s2)
	ring.AddNode(s3)

	keys := []string{
		"user1",
		"user2",
		"user3",
		"user4",
		"user5",
	}

	fmt.Println("Before removal")

	for _, key := range keys {
		server, _ := ring.GetNode(key)
		fmt.Printf("%s -> %s\n", key, server.ID)
	}

	ring.RemoveNode(s2)

	fmt.Println("\nAfter removal")

	for _, key := range keys {
		server, _ := ring.GetNode(key)
		fmt.Printf("%s -> %s\n", key, server.ID)
	}

}
