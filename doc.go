/*
	Package btrie provides a binary trie implementation (radix or prefix tree).
	The trees are partially compressed, i.e. strings of single-child nodes are
	compressed from the end of the tree, producing a tree only as deep as the
	longest common prefix. However, edges in the middle do not compress so the
	space usage is suboptimal when keys have long common prefixes.
	
	Keys are made up of an arbitrary length of bytes values are arbitrary
	interfaces

	Naturally, the advantage of using a binary trie instead of a hashtable
	(natively implemented as golang maps) is the ability to obtain all entries
	in a sorted fashion or to efficiently explore the "neighbourhood" of a given
	entry by requesting the subtree or querying on a range of values.

	Performance is optimized, among other things, by using iterative rather
	than recursive algorithms.

	This implementation is not thread-safe (goroutine-safe). Calling code should
	synchronize if necessary.
*/
package btrie
