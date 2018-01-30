# btrie
An in-memory binary Trie implementation (also called radix or prefix trees)

GoDoc Documentation: https://godoc.org/github.com/bdargham/btrie

Test Coverage: [![Coverage Status](https://coveralls.io/repos/github/bdargham/btrie/badge.svg?branch=master)](https://coveralls.io/github/bdargham/btrie?branch=master)

## Simple usage example

```package main

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
```
