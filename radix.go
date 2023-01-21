package dune

import (
	"fmt"
	"net/http"
	"sort"
	"unicode"
)

type Handler = func(http.ResponseWriter, *http.Request, *Params)

type node struct {
	prefix   string
	template string
	children []*node
	param    *node
	wildcard *node
	handler  Handler
	index    struct {
		minChar uint8
		maxChar uint8

		// Map all the characters between minChar and maxChar. Therefore, length = maxChar - minChar.
		// The values point to the relevant children node position for each character.
		//
		// Value == 0 indicates, there's no matching child node for the character.
		// Value > 0 points to the index of the matching child node + 1.
		//
		//	e.g.:
		//	minChar = 97 (a)
		//	maxChar = 100 (d)
		//
		//	children[0] =  97 (a) node
		//	children[2] =  99 (c) node
		//	children[3] = 100 (d) node
		//
		// 	indices[0] = 1
		//	indices[1] = 0
		//	indices[2] = 2
		//	indices[3] = 3
		indices []uint8

		// Index the character lengths of child node prefixes following the exact order of indices.
		//
		// 	e.g.:
		//	minChar = 97 (a)
		//	maxChar = 100 (d)
		//
		//	children[0] =  97 (a) node, prefix = 'apple'
		//	children[2] =  99 (c) node, prefix = 'castle'
		//	children[3] = 100 (d) node, prefix = 'dang'
		//
		// 	size[0] = 5
		// 	size[1] = 0
		//	size[2] = 6
		//	size[3] = 4
		size []int
	}
}

func newRootNode() *node {
	return &node{
		template: "/",
	}
}

func (n *node) insert(path string, handler Handler) (varsCount int) {
	varsCount = scanPath(path)

	if path == "" {
		// Root node.
		n.template = "/"
		n.handler = handler
		return
	}

	newNode := n.addNode(path)
	newNode.template = path
	newNode.handler = handler
	return
}

func (n *node) addNode(path string) *node {
	if path[0] == '/' {
		path = path[1:]
	}

	root := n
	r := newRouteScanner(path)

	for seg := r.next(); seg != ""; seg = r.next() {
		switch seg[0] {
		case ':':
			if root.param != nil {
				if root.param.prefix != seg {
					panic(fmt.Sprintf("param node is already registered with the name %s", root.param.prefix))
				}
				root = root.param
				continue
			}

			root.param = &node{prefix: seg}
			root = root.param
		case '*':
			if root.wildcard != nil {
				panic("wildcard route already registered")
			}

			root.wildcard = &node{prefix: seg}
			root = root.wildcard
		default:
		DFS:
			if seg == "" {
				continue
			}

			candidate, candidateIdx := root.findCandidateByChar(seg[0])
			if candidate == nil {
				child := &node{prefix: seg}
				root.children = append(root.children, child)
				root.reindex()
				root = child
				continue
			}

			longest := longestPrefix(seg, candidate.prefix)

			// Traversal.
			// pfx: /posts
			// seg: /posts|/upsert
			if longest == len(candidate.prefix) {
				root = candidate
				seg = seg[longest:]
				goto DFS
			}

			// Expansion.
			// pfx: categories|/skus
			// seg: categories|
			if longest == len(seg) {
				// Shift down the candidate node and allocate its prior state to the segment.
				branchNode := &node{prefix: candidate.prefix[:longest], children: make([]*node, 1)}

				candidate.prefix = candidate.prefix[longest:]

				branchNode.children[0] = candidate
				branchNode.reindex()

				root.children[candidateIdx] = branchNode
				root.reindex()

				root = branchNode
				continue
			}

			// Collision.
			// pfx: cat|egories
			// seg: cat|woman

			// Split the node into 2 at the point of collision.
			newNode := &node{prefix: seg[longest:]}

			branchNode := &node{prefix: candidate.prefix[:longest], children: make([]*node, 2)}
			branchNode.children[0] = candidate
			branchNode.children[1] = newNode

			candidate.prefix = candidate.prefix[longest:]

			branchNode.reindex()

			root.children[candidateIdx] = branchNode
			root.reindex()

			root = newNode
			continue
		}
	}

	return root
}

// findCandidateByCharAndSize search for a children by matching the first char and length.
// If no match is found, it looks up indexer#trailingSlash to see if there's a possible match who has a trailing slash.
// If found, returns the found children with trailing slash and true for the 2nd return value.
// Otherwise, return nil, false.
//
// When ts (2nd return value) is false, there's a guarantee that len(s) >= len(child prefix).
// When ts is true, len(s) = len(child prefix) - 1.
func (n *node) findCandidateByCharAndSize(c uint8, size int) *node {
	if n.index.minChar <= c && c <= n.index.maxChar {
		offset := c - n.index.minChar
		index := n.index.indices[offset]
		if index == 0 {
			return nil
		}

		childSize := n.index.size[offset]
		if size >= childSize {
			return n.children[index-1] // Decrease by 1 to get the exact child node index.
		}
	}

	return nil
}

func (n *node) findCandidateByChar(c uint8) (*node, uint8) {
	if n.index.minChar <= c && c <= n.index.maxChar {
		offset := c - n.index.minChar
		childIndex := n.index.indices[offset]
		if childIndex == 0 {
			return nil, 0
		}

		return n.children[childIndex-1], childIndex - 1 // Decrease by 1 to get the exact child node index.
	}

	return nil, 0
}

func (n *node) reindex() {
	if len(n.children) == 0 {
		return
	}

	sort.Slice(n.children, func(i, j int) bool {
		return n.children[i].prefix[0] < n.children[j].prefix[0]
	})

	n.index.minChar = n.children[0].prefix[0]
	n.index.maxChar = n.children[len(n.children)-1].prefix[0]
	rng := n.index.maxChar - n.index.minChar + 1

	if len(n.index.indices) != int(rng) {
		n.index.indices = make([]uint8, rng)
	}

	if len(n.index.size) != int(rng) {
		n.index.size = make([]int, rng)
	}

	for i, child := range n.children {
		idx := child.prefix[0] - n.index.minChar
		n.index.indices[idx] = uint8(i) + 1
		n.index.size[idx] = len(child.prefix)
	}
}

func (n *node) search(path string, paramInjector func() *Params) (*node, *Params) {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if path == "" {
		return n, nil
	}

	fn, ps := n._search(path, nil, paramInjector)
	return fn, ps
}

func (n *node) _search(path string, params *Params, paramInjector func() *Params) (*node, *Params) {

	// Look for a child node whose first char equals searching path's first char and prefix length
	// is less than or equal searching path's length.
	if child := n.findCandidateByCharAndSize(path[0], len(path)); child != nil {

		// Find the longest common prefix between child's prefix and searching path.
		// If child's prefix is fully matched, continue...
		// Otherwise, fallback...
		if longest := longestPrefix(child.prefix, path); longest == len(child.prefix) {

			// Perfect match. And no further segments are left to cover in the searching path.
			if longest == len(path) {
				if child.handler != nil {
					return child, params
				}

				// Though there's a matching node, it doesn't have a handler.
				// Try to elect matched node's wildcard node.
				// No need to nil check wildcard node's handler since wildcard nodes would always have a handler.
				if child.wildcard != nil {
					if params == nil {
						params = paramInjector()
					}
					params.set(child.wildcard.prefix[1:], path[longest:])
					return child.wildcard, params
				}

				return nil, params
			} else {
				// There are more segments to cover in the searching path.

				// Traverse the child node recursively until a match is found.
				var dfsChild *node
				if dfsChild, params = child._search(path[len(child.prefix):], params, paramInjector); dfsChild != nil && dfsChild.handler != nil {
					// Found a matching node with a registered handler.
					return dfsChild, params
				}

				// Didn't find a matching node by traversing the child node.

				// Hence, traverse child's param node.
				// TODO: This may not be necessary.
				//
				// eg:
				// node 1: /search/go		// This might be the right match, but NO!
				// node 2: /search/:var		// So let's fallback to the param node
				//
				// search: /search/gone

				//if child.param != nil {
				//	remPath := path[longest:]
				//
				//	r := routeScanner{path: remPath}
				//
				//	if params == nil {
				//		params = paramInjector()
				//	}
				//
				//	if idx := r.indexOf('/'); idx == -1 {
				//		// No more segments in the path.
				//		if child.param.handler != nil {
				//			params.set(child.param.prefix[1:], path)
				//			return child.param, params
				//		}
				//
				//		// The param node might have children who have handlers but no need to explore them since the searching path has no more segments.
				//		// Due to the matched param node having no handler, fallback to the wildcard node.
				//		goto Param
				//	} else {
				//		// Traverse the param node until all the segments are exhausted.
				//		params.set(child.param.prefix[1:], remPath[:idx])
				//		if child, params := child.param._search(remPath[idx:], params, paramInjector); child != nil && child.handler != nil {
				//			return child, params
				//		}
				//	}
				//
				//	goto Param
				//}
			}
		}
	}

	// Didn't find a matching node.

	// Fallback to param node.

	if n.param != nil {
		r := routeScanner{path: path}

		if params == nil {
			params = paramInjector()
		}

		// Check if more segments are left to cover in the searching path.
		if idx := r.indexOf('/'); idx == -1 {

			// No more segments in the path.
			if n.param.handler != nil {
				params.set(n.param.prefix[1:], path)
				return n.param, params
			}

			// The param node might have children who have handlers but no need to explore them
			// since the searching path has no more segments left to cover.
			// Thus, fallback to the wildcard node.

		} else {

			// Traverse the param node until all the segments are exhausted.
			if child, params := n.param._search(path[idx:], params, paramInjector); child != nil && child.handler != nil {
				params.set(n.param.prefix[1:], path[:idx])
				return child, params
			}
		}
	}

	// Fallback to wildcard node.
	//
	// This also facilitates to fall back to the nearest wildcard node in the recursion stack when no match is found.
	//
	// No need to nil check wildcard node's handler since wildcard nodes must have a handler.
	if n.wildcard != nil {
		if params == nil {
			params = paramInjector()
		}

		params.set(n.wildcard.prefix[1:], path)
		return n.wildcard, params
	}

	return nil, params
}

func scanPath(path string) (varsCount int) {
	if path == "" || path[0] != '/' {
		panic("path must have a leading slash")
	}

	inParams := false
	inWC := false
	for i, c := range []byte(path) {
		if unicode.IsSpace(rune(c)) {
			panic("path shouldn't contain any whitespace")
		}

		if inWC {
			switch c {
			case '/', ':':
				panic("another segment shouldn't follow a wildcard segment")
			case '*':
				panic("only one wildcard segment is allowed")
			}
		}

		if inParams {
			switch c {
			case '/':
				if path[i-1] == ':' {
					panic("param must have a name")
				}
				inParams = false
				continue
			case ':':
				panic("only one param segment is allowed within the same scope")
			case '*':
				panic("wildcard segment shouldn't follow the param segment within the same scope")
			}
		}

		if c == '*' {
			inWC = true
			varsCount++
			continue
		}

		if c == ':' {
			inParams = true
			varsCount++
			continue
		}
	}

	if inParams && path[len(path)-1] == ':' {
		panic("param must have a name")
	}

	if inWC && path[len(path)-1] == '*' {
		panic("wildcard must have a name")
	}

	return
}

func findParamsCount(path string) (c int) {
	for _, b := range []byte(path) {
		if b == ':' || b == '*' {
			c++
		}
	}
	return c
}

func longestPrefix(s1, s2 string) int {
	max := len(s1)
	if len(s2) < max {
		max = len(s2)
	}

	i := 0
	for i < max {
		if s1[i] != s2[i] {
			return i
		}
		i++
	}
	return i
}
