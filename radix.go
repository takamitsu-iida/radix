package radix

import (
	"sort"
)

// Tree definition
type Tree struct {
	root *node
	size int
}

// Constructor
// New() returns empty Tree instance
func New() *Tree {
	return &Tree{
		root: &node{},
		size: 0,
	}
}

// Len() returns number of key-value-pairs stored in the Tree
func (t *Tree) Len() int {
	return t.size
}

// Load() receive a mpa and store its contents in the Tree
func (t *Tree) Load(m map[string]interface{}) {
	for k, v := range m {
		t.Insert(k, v)
	}
}

// ToMap() is reverse of Load()
// Convert key-value-pair to map and return
func (t *Tree) ToMap() map[string]interface{} {
	m := make(map[string]interface{}, t.size)
	t.Walk(func(k string, v interface{}) bool {
		m[k] = v
		return false
	})
	return m
}

// node definition
type node struct {
	leaf     *Leaf  // reference to a leaf node or nil
	prefixes []rune // Unique part excluding the intersection until this node
	edges    []edge // slice of edge, always kept sorted
}

// Leaf definition, Leaf stores a key-value-pair
type Leaf struct {
	key   string
	value interface{}
}

// edge definition, edge have single-letter labels that identify branches
type edge struct {
	label rune  // single letter
	node  *node // child node
}

// returns true if it has a reference to a leaf node
func (n *node) isLeaf() bool {
	return n.leaf != nil
}

// add the specified edge
func (n *node) addEdge(e edge) {
	n.edges = append(n.edges, e)
	// keep sorted
	sort.Slice(n.edges, func(i, j int) bool { return n.edges[i].label < n.edges[j].label })
}

// return number of edges
func (n *node) edgeLen() int {
	return len(n.edges)
}

// returns the child node beyond the edge of the specified label letter
func (n *node) getChild(label rune) *node {
	// find the same label in the edge slice
	length := len(n.edges)
	index := sort.Search(length, func(i int) bool {
		return n.edges[i].label >= label
	})

	// if found, return the child node at the end of the edge
	if index < length && n.edges[index].label == label {
		return n.edges[index].node
	}

	return nil
}

// if there is an edge corresponding to the specified label letter,
// change the reference to the child node and return true, otherwise return false
func (n *node) updateEdge(label rune, node *node) bool {
	// find the same label in the edge slice
	length := len(n.edges)
	index := sort.Search(length, func(i int) bool {
		return n.edges[i].label >= label
	})

	// if found, replace the child node of the edge with the specified node
	if index < length && n.edges[index].label == label {
		n.edges[index].node = node
		return true
	}

	return false
}

// delete the edge with the specified label
func (n *node) deleteEdge(label rune) bool {
	length := len(n.edges)
	index := sort.Search(length, func(i int) bool {
		return n.edges[i].label >= label
	})

	// if found, delete n.edges[index]
	if index < length && n.edges[index].label == label {
		n.edges = n.edges[:index+copy(n.edges[index:], n.edges[index+1:])]

		/*
			copy(n.edges[index:], n.edges[index+1:])
			n.edges[len(n.edges)-1] = edge{}   // replace last element as empty edge
			n.edges = n.edges[:len(n.edges)-1] // remove last element
		*/

		return true
	}

	return false
}

// return the smallest int
func minIntOf(vars ...int) int {
	min := vars[0]
	for _, i := range vars {
		if min > i {
			min = i
		}
	}
	return min
}

// returns the number of characters common to the given key1 and key2
func commonLength(key1, key2 []rune) int {
	max := minIntOf(len(key1), len(key2))

	// check one character at a time and if a character that is not common is found,
	// exit the loop and return the current length
	var i int
	for i = 0; i < max; i++ {
		if key1[i] != key2[i] {
			break
		}
	}
	return i
}

func startsWith(runes, prefixes []rune) bool {
	if len(prefixes) > len(runes) {
		return false
	}
	return len(prefixes) == commonLength(runes, prefixes)
}

//
// ツリーをルートノードからたどってノードにたどり着くたびにプレフィクス文字列を探索キーから削っていく。
// この繰り返しをループで実装するか、再帰関数で作るか、迷うところ。
// ここではループで実装
//

// Add a new key-value pair to the tree.
// returns true if newly inserted.
// returns false if update existing key-value pair.
func (t *Tree) Insert(k string, v interface{}) (inserted bool) {
	// Make a rune slice of k and use it as a search key
	searches := []rune(k)

	var parent *node
	n := t.root
	for {

		// The search key length is 0, which means that the existing node has that key.
		if len(searches) == 0 {
			if n.isLeaf() {
				n.leaf.key = k
				n.leaf.value = v
				return false // false means overwrite existing node
			}

			// create a new leaf
			n.leaf = &Leaf{
				key:   k,
				value: v,
			}
			t.size++    // increase the size of the tree by +1
			return true // true means newly inserted
		}

		// shift the parent node to n
		parent = n

		// shift n to the child node and proceed tree search
		n = n.getChild(searches[0])

		// if child node n does not exist, create an edge, spawn a new branch and exit
		if n == nil {
			e := edge{
				label: searches[0],
				node: &node{
					leaf: &Leaf{
						key:   k,
						value: v,
					},
					prefixes: searches,
				},
			}
			parent.addEdge(e)
			t.size++    // increase the size of the tree by +1
			return true // true means newly inserted
		}

		//
		// If child node n exists, still in the process of searching.
		//

		// find out the length of the common part between searches and the prefixes of the node
		commonLen := commonLength(searches, n.prefixes)

		// If length of the common part matches the length of n.prefixes,
		// the search key may be longer than n.prefixes,
		// so take out the unique part and continue the search.
		if commonLen == len(n.prefixes) {
			searches = searches[commonLen:] // unique part
			continue
		}

		// If length of the common part is shorter than search key,
		// split n into n1 and n2 and branch out from n1
		//   BEFORE: parent -(edge)- n
		//   AFTER : parent -(edge)- n1 -+-(edge)--- n2
		//                               +-(edge)--- newNode
		//

		n1 := &node{}                        // create new node n1
		n1.prefixes = n.prefixes[:commonLen] // n1 has common part of the prefixes

		n2 := n                              // n2 should take over n
		n2.prefixes = n.prefixes[commonLen:] // unique part of the prefixes

		parent.updateEdge(searches[0], n1) // change the parent node's edge from n to n1

		n1.addEdge(edge{label: n2.prefixes[0], node: n2}) // add edge to n2

		// create new leaf and size +1
		leaf := &Leaf{
			key:   k,
			value: v,
		}
		t.size++

		// the unique part after the common part becomes the prefixes
		prefixes := searches[commonLen:]

		// if the search key length is 0, then n1 has the key
		if len(prefixes) == 0 {
			n1.leaf = leaf
		} else {
			// add new edge to n1 and hang a new node with leaf
			n1.addEdge(edge{
				label: prefixes[0],
				node: &node{
					leaf:     leaf,
					prefixes: prefixes,
				},
			})
		}

		return true
	}
}

// Delete key-value pair and returns its value and true.
// If key not found, returns nil and false.
func (t *Tree) Delete(key string) (value interface{}, deleted bool) {
	// default (when not deleted) return value
	value = nil
	deleted = false

	// Search logic is similar to Insert ()

	searches := []rune(key)
	var parent *node
	var parentLabel rune
	n := t.root
	for {
		if len(searches) == 0 {
			if n.isLeaf() {
				leaf := n.leaf
				n.leaf = nil
				t.size--

				// If n has no edge, delete the edge from parent to n
				if parent != nil && len(n.edges) == 0 {
					parent.deleteEdge(parentLabel)
				}

				// If n has only one edge, mearge n and child
				if n != t.root && len(n.edges) == 1 {
					n.mergeChild()
				}

				// If parent has only one edge and parent has no leaf, merge parent and n
				if parent != nil && parent != t.root && len(parent.edges) == 1 && !parent.isLeaf() {
					parent.mergeChild()
				}
				return leaf.value, true
			}
			break
		}

		// shift the parent node to n
		parent = n
		parentLabel = searches[0]

		// Find child node
		n = n.getChild(searches[0])
		if n == nil {
			break // key not found, means no target node
		}

		if startsWith(searches, n.prefixes) {
			searches = searches[len(n.prefixes):] // shift the search key to a unique part and continue searching
		} else {
			break // exit the loop because there is no key to look for
		}
	}

	return value, deleted
}

func (n *node) mergeChild() {
	if len(n.edges) != 1 {
		return
	}
	e := n.edges[0]
	child := e.node
	n.prefixes = append(n.prefixes, child.prefixes...)
	n.leaf = child.leaf
	n.edges = child.edges
}

// If there is a key-value pair corresponding to given key, it will be returned,
// otherwise nil and false will be returned.
func (t *Tree) Get(key string) (interface{}, bool) {
	searches := []rune(key)
	n := t.root
	for {
		if len(searches) == 0 {
			if n.isLeaf() {
				return n.leaf.value, true
			}
			break
		}

		n = n.getChild(searches[0])
		if n == nil {
			// no child means key not found
			break
		}

		if startsWith(searches, n.prefixes) {
			searches = searches[len(n.prefixes):]
		} else {
			break
		}
	}
	return nil, false
}

// Returns the closest key-value pair in a longest match rule
func (t *Tree) LongestMatch(key string) (string, interface{}, bool) {
	searches := []rune(key)
	var last *Leaf
	n := t.root
	for {
		if n.isLeaf() {
			last = n.leaf
		}

		if len(searches) == 0 {
			break
		}

		n = n.getChild(searches[0])
		if n == nil {
			break
		}

		if startsWith(searches, n.prefixes) {
			searches = searches[len(n.prefixes):]
		} else {
			break
		}
	}

	// this is different from Get()
	// return the last found
	if last != nil {
		return last.key, last.value, true
	}

	return "", nil, false
}

// Find all key-values starting with a given key
func (t *Tree) Collect(key string) []Leaf {
	leafs := []Leaf{}

	searches := []rune(key)
	var found *node
	n := t.root
	for {
		if len(searches) == 0 {
			found = n
			break
		}

		n = n.getChild(searches[0])
		if n == nil {
			// no child means key not found in this tree
			return leafs
		}

		if startsWith(searches, n.prefixes) {
			searches = searches[len(n.prefixes):]
			continue // searching
		}

		if startsWith(n.prefixes, searches) {
			found = n
		}
		break
	}

	if found == nil {
		return leafs
	}

	// starting from the found, collect all key-value pairs
	walk(found, func(k string, v interface{}) bool {
		leafs = append(leafs, Leaf{key: k, value: v})
		return false
	})

	return leafs
}

func (t *Tree) CollectKeys(key string) []string {
	leafs := t.Collect(key)
	if leafs == nil {
		return []string{}
	}

	keys := []string{}
	for _, leaf := range leafs {
		keys = append(keys, leaf.key)
	}
	return keys
}

// The edges are sorted, so if you follow the younger edge, you will reach the top value.
func (t *Tree) Top() (string, interface{}, bool) {
	n := t.root
	for {
		if n.isLeaf() {
			return n.leaf.key, n.leaf.value, true
		}
		if len(n.edges) > 0 {
			n = n.edges[0].node
		} else {
			break
		}
	}
	return "", nil, false
}

// The edges are sorted, so if you follow the older edge, you will reach the last value.
func (t *Tree) Bottom() (string, interface{}, bool) {
	n := t.root
	for {
		if num := len(n.edges); num > 0 {
			n = n.edges[num-1].node
			continue
		}
		if n.isLeaf() {
			return n.leaf.key, n.leaf.value, true
		}
		break
	}
	return "", nil, false
}

// Callback function to pass when exploring the tree.
// If true is returned, the tree search will stop at that point.
type WalkCallback func(s string, v interface{}) bool

// Follow the tree from the root node and execute the callback function when you find the leaf
func (t *Tree) Walk(fn WalkCallback) {
	walk(t.root, fn)
}

// Call the callback function when the leaf is reached, and end the search when it returns true
func walk(n *node, fn WalkCallback) bool {
	if n.leaf != nil {
		if fn(n.leaf.key, n.leaf.value) {
			return true
		}
	}

	// move to the child node beyond the edge of n and continue searching
	for _, e := range n.edges {
		if walk(e.node, fn) {
			return true
		}
	}

	return false
}
