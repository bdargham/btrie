package main

import (
	"fmt"
	"github.com/bdargham/btrie"
)

func main() {

	trie := btrie.New()

	trie.SPut("perf", "When things go fast")
	trie.SPut("winter", "... when it's really cold")
	trie.SPut("fall", "Autumn")
	trie.SPut("win", "We don't want to lose... ")
	trie.SPut("fallout", "Mushroom cloud")

	dfsOpts := btrie.TraversalOpts{From: []byte("w")}
	for cursor := trie.Traverse(dfsOpts); cursor.HasNext(); {
		fmt.Println(cursor.Next().Value())
	}
}
