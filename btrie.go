// Copyright 2018 Basheer Dargham

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package btrie

import "bytes"

// Entry represents a key-value pair stored in the BTrie
type Entry struct {
	key   []byte
	value interface{}
}

func (self Entry) Key() []byte {
	return self.key
}

func (self Entry) Value() interface{} {
	return self.value
}

type node struct {
	next [2]*node
	leaf *Entry
}

// BTrie represents the entire in-memory trie structure
type BTrie struct {
	root node
	size int
}

// New is a constructor to create a new BTrie
func New() *BTrie {
	return &BTrie{}
}

func (self *BTrie) Size() int {
	return self.size
}

func (self *BTrie) drillDown(key []byte, stack *stack, dir *Direction) (closest *node, isMatch bool) {
	current := &self.root
outer:
	for i := 0; i < len(key); i++ {
		for j := 7; j >= 0; j-- {
			bit := (key[i] >> byte(j)) & 1
			if current.next[bit] != nil {
				if stack != nil {
					switch {
					case dir == nil:
						stack.push(current)
					case *dir == Forward && bit == 0:
						stack.push(current.next[1])
					case *dir == Backwards && bit == 1:
						stack.push(current.next[0])
					}
				}
				current = current.next[bit]
			} else {
				break outer
			}
		}
	}
	return current, current.leaf != nil && bytes.Equal(current.leaf.key, key)
}

// SPut is a convenience method for using string keys (and converting them to byte slices)
func (self *BTrie) SPut(key string, value interface{}) (oldvalue interface{}) {
	return self.Put([]byte(key), value)
}

// SGet is a convenience method for using string keys (and converting them to byte slices)
func (self *BTrie) SGet(key string) (value interface{}) {
	return self.Get([]byte(key))
}

// SRemove is a convenience method for using string keys (and converting them to byte slices)
func (self *BTrie) SRemove(key string) (oldvalue interface{}) {
	return self.Remove([]byte(key))
}

// Put writes values to the trie. If value is nil, call is equivalent to Remove,
// so no nil values can be stored
func (self *BTrie) Put(key []byte, value interface{}) (oldvalue interface{}) {

	if value == nil {
		return self.Remove(key)
	}

	current := &self.root

outer:
	for i := 0; i < len(key); i++ {
		for j := 7; j >= 0; j-- {

			bit := (key[i] >> byte(j)) & 1

			switch {
			case current.next[bit] != nil:
				current = current.next[bit]
			case current.next[^bit&1] != nil:
				current.next[bit] = &node{}
				current = current.next[bit]
			case current.leaf == nil:
				break outer
			case j == 7 && len(current.leaf.key) == i:
				// current.leaf.key can't be moved down any further
				current.next[bit] = &node{}
				current = current.next[bit]
				break outer
			default:
				cbit := (current.leaf.key[i] >> byte(j)) & 1
				if cbit != bit {
					current.next[cbit] = &node{leaf: current.leaf}
					current.leaf = nil
					current.next[bit] = &node{}
					current = current.next[bit]
					break outer
				} else {
					current.next[bit] = &node{leaf: current.leaf}
					current.leaf = nil
					current = current.next[bit]
				}
			}
		}
	}
	if current.leaf != nil {
		if len(current.leaf.key) == len(key) {
			oldvalue = current.leaf.value
		} else {
			// New key has reached it's full length but needs to push the old node one step lower
			cbit := (current.leaf.key[len(key)] >> 7) & 1
			current.next[cbit] = &node{leaf: current.leaf}
		}
	}
	current.leaf = &Entry{make([]byte, len(key)), value}
	copy(current.leaf.key, key)
	self.size++
	return
}

// Get returns the previously stored value, or nil if not present
func (self *BTrie) Get(key []byte) (value interface{}) {
	if res, isMatch := self.drillDown(key, nil, nil); isMatch {
		return res.leaf.value
	}
	return nil
}

// Remove removes values from the trie. If key is not found, nothing
// is done and nil is returned as an oldvalue
func (self *BTrie) Remove(key []byte) (oldvalue interface{}) {
	parents := newStack()
	if node, isMatch := self.drillDown(key, parents, nil); isMatch {
		oldvalue = node.leaf.value
		node.leaf = nil
		parents.push(node)
		for parents.len() > 1 { // never delete the root node
			n := parents.pop()
			if n.leaf == nil && n.next[0] == nil && n.next[1] == nil {
				p := parents.peep()
				for i := 0; i < 2; i++ {
					if p.next[i] == n {
						p.next[i] = nil
					}
				}
			} else {
				break
			}
		}
		self.size--
	}
	return
}

// Direction specifies the direction in which the trie is to be traversed
type Direction int

const (
	Forward   Direction = iota // Traverse trie from lowest to highest key values
	Backwards                  // Traverse trie from highest to lowest key values
)

// TraversalOpts specifies how the trie is to be traversed. Defaults are provided
// in the struct for easier calls
type TraversalOpts struct {
	Dir Direction

	// From indicates the key at which to start traversal
	// (whether or not that key is in the trie)
	From []byte

	// SubtreeOnly indicates that only entries who's keys are a prefix
	// of the From key are to be visited.
	SubtreeOnly bool

	// Limit sets the max number of values to return. If 0, return values with no limit
	Limit uint64 
}

const MaxUint64 = ^uint64(0)

// TraverseFully does a complete DFS (Depth-First Search) traversal of the entire trie
func (self *BTrie) TraverseFully() *Cursor {
	return self.Traverse(TraversalOpts{})
}

// Traverse does a DFS (Depth-First Search) traversal in which certain options can be specified
func (self *BTrie) Traverse(opts TraversalOpts) *Cursor {
	if opts.Limit == 0 {
		opts.Limit = MaxUint64
	}
	result := Cursor{nil, *newStack(), opts}
	startNode := &self.root
	if opts.From != nil {
		stack := &result.stack
		if opts.SubtreeOnly {
			stack = nil
		}
		startNode, _ = self.drillDown(opts.From, stack, &opts.Dir)
	}
	result.stack.push(startNode)
	return &result
}

// Cursor defines the result of a trie traversal and allows the caller
// to iterate on the results.
//
// As cursor's point the data within the trie, behaviour is
// undetermined if tree content changes while traversal is in
// progress. All synchronization needs to be performed outside
// the library.
type Cursor struct {
	curr  *node
	stack stack
	opts TraversalOpts
}

// HasNext should always be called before Next.
func (self *Cursor) HasNext() bool {
	if self.curr != nil {
		return true
	}
	for {
		if self.stack.len() == 0 || self.opts.Limit == 0 {
			return false
		}
		self.curr = self.stack.pop()
		self.stack.push(self.curr.next[1&^int(self.opts.Dir)])
		self.stack.push(self.curr.next[int(self.opts.Dir)])
		if self.curr.leaf != nil {
			self.opts.Limit--
			return true
		}
	}
}

func (self *Cursor) Next() *Entry {
	result := self.curr.leaf
	self.curr = nil
	return result
}

type stack struct {
	s []*node
}

func newStack() *stack {
	return &stack{make([]*node, 0, 16)}
}

func (self *stack) len() int {
	return len(self.s)
}

func (self *stack) push(val *node) {
	if val != nil {
		self.s = append(self.s, val)
	}
}

func (self *stack) pop() *node {
	l := len(self.s) - 1
	res := self.s[l]
	self.s = self.s[:l]
	return res
}

func (self *stack) peep() *node {
	return self.s[len(self.s)-1]
}
