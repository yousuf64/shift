package shift

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type node struct {
	prefix    string
	template  string
	children  []*node
	param     *node
	wildcard  *node
	handler   HandlerFunc
	paramKeys *[]string // Nil paramKeys denote the route is static.
	index     struct {
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

func (n *node) insert(path string, handler HandlerFunc) (varsCount int) {
	varsCount = scanPath(path)

	if path == "" {
		// Root node.
		n.template = "/"
		n.handler = handler
		return
	}

	newNode, paramKeys := n.addNode(path)
	if newNode.handler != nil {
		panic(fmt.Sprintf("%s conflicts with already registered route %s", path, newNode.template))
	}

	newNode.template = path
	newNode.handler = handler
	if len(paramKeys) > 0 {
		rs := reverseSlice(paramKeys)
		newNode.paramKeys = &rs
	}
	return
}

func reverseSlice(s []string) (rs []string) {
	if len(s) > 1 {
		for i := 0; i < len(s)/2; i++ {
			(s)[i], (s)[len(s)-1-i] = (s)[len(s)-1-i], (s)[i]
		}
	}
	return s
}

func (n *node) addNode(path string) (root *node, paramKeys []string) {
	if path[0] == '/' {
		path = path[1:]
	}

	root = n
	r := newRouteScanner(path)

	for seg := r.next(); seg != ""; seg = r.next() {
		switch seg[0] {
		case ':':
			paramKeys = append(paramKeys, seg[1:])
			if root.param != nil {
				root = root.param
				continue
			}

			root.param = &node{prefix: ":"}
			root = root.param
		case '*':
			paramKeys = append(paramKeys, seg[1:])
			if root.wildcard != nil {
				root = root.wildcard
				break
			}

			root.wildcard = &node{prefix: "*"}
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

	return root, paramKeys
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

	// Sort children by prefix's first char.
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

	return n.searchRecursion(path, nil, paramInjector)
}

// searchRecursion recursively traverses the radix tree looking for a matching node.
// Returns the matched node if found.
// Returns Params only when matched node is a param node. Returns <nil> otherwise.
func (n *node) searchRecursion(path string, params *Params, paramInjector func() *Params) (*node, *Params) {
	// Search a matching node inside node's children.
	// Char could be indexed?
	if c := path[0]; n.index.minChar <= c && c <= n.index.maxChar {
		// Yes, char could be indexed...

		// Is char really indexed?
		if idx := n.index.indices[c-n.index.minChar]; idx != 0 {
			// Char is indexed!!!

			if child := n.children[idx-1]; child != nil {
				if path == child.prefix {
					// Perfect match.
					// path: /foobar
					// pref: /foobar

					// Dead end #1
					if child.handler != nil {
						if child.paramKeys != nil {
							params = paramInjector()
							params.setKeys(child.paramKeys)
						}
						return child, params
					}

					// But a handler is not registered :(
					//
					// So, lets fallback to wildcard node...
					// 	No need to perform <nil> check for handler and paramKeys here
					// 	since a wildcard node must always have a handler and paramKeys.
					//
					// Dead end #2
					if child.wildcard != nil {
						params = paramInjector()
						params.setKeys(child.wildcard.paramKeys)
						params.appendValue(path[len(child.prefix):])
						return child.wildcard, params
					}

					// No match :/
					return nil, params
				} else if strings.HasPrefix(path, child.prefix) {
					// path: /foobar
					// pref: /foo

					// Explore child...
					var innerChild *node
					innerChild, params = child.searchRecursion(path[len(child.prefix):], params, paramInjector)
					if innerChild != nil && innerChild.handler != nil {
						return innerChild, params
					}
				}
			}
		}
	}

	// Couldn't find a matching node within children nodes.
	// So lets fallback to param node.
	if n.param != nil {
		// Check if more sections are left to match in the path.

		if idx := strings.IndexByte(path, '/'); idx == -1 {
			// No more sections to match.
			// Dead end #3
			if n.param.handler != nil {
				params = paramInjector()
				params.setKeys(n.param.paramKeys) // Param node would always have paramKeys.
				params.appendValue(path)
				return n.param, params
			}
		} else {
			// Traverse the param node until all the path sections are matched.
			var innerChild *node
			innerChild, params = n.param.searchRecursion(path[idx:], params, paramInjector)
			if innerChild != nil && innerChild.handler != nil {
				params.appendValue(path[:idx])
				return innerChild, params
			}
		}
	}

	// No luck with param node :/
	// Lets fallback to wildcard node.
	//	No need to perform <nil> check for handler and paramKeys here
	// 	since a wildcard node must always have a handler and paramKeys.
	//
	// Dead end #4
	if n.wildcard != nil {
		params = paramInjector()
		params.setKeys(n.wildcard.paramKeys)
		params.appendValue(path)
		return n.wildcard, params
	}

	// No match :(((
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

func (n *node) caseInsensitiveSearch(path string, paramInjector func() *Params) (*node, *Params, string) {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if path == "" {
		return n, nil, ""
	}

	var buf reverseBuffer = newReverseBuffer128() // No heap allocation.
	if lng := len(path) + 1; lng > 128 {          // Account an additional space for the leading slash.
		buf = newSizedReverseBuffer(lng) // For long paths, allocate a sized buffer on heap.
	}

	fn, ps := n.caseInsensitiveSearchRecursion(path, nil, paramInjector, buf)
	if fn != nil && fn.handler != nil {
		buf.WriteString("/") // Write leading slash.
	}
	return fn, ps, buf.String()
}

// TODO: optimize search...
func (n *node) caseInsensitiveSearchRecursion(path string, params *Params, paramInjector func() *Params, buf reverseBuffer) (*node, *Params) {
	var swappedChild bool

	// Look for a child node whose first char equals searching path's first char and prefix length
	// is less than or equal searching path's length.
	child := n.findCandidateByCharAndSize(path[0], len(path))

TraverseChild:
	if child != nil {
		// Find the longest common prefix between child's prefix and searching path.
		// If child's prefix is fully matched, continue...
		// Otherwise, fallback...
		if longest := longestPrefixCaseInsensitive(child.prefix, path); longest == len(child.prefix) {
			if longest == len(path) {
				// Perfect match. And no further segments are left to cover in the searching path.
				// path: /foobar
				// pref: /foobar

				// Dead end #1
				if child.handler != nil {
					if child.paramKeys != nil {
						params = paramInjector()
						params.setKeys(child.paramKeys)
					}
					buf.WriteString(child.prefix)
					return child, params
				}

				// Though there's a matching node, it doesn't have a handler.
				// Try to elect matched node's wildcard node.
				//// No need to perform nil check for handler and paramKeys here
				//// since a wildcard node must always have a handler and paramKeys.
				//
				// Dead end #2
				if child.wildcard != nil {
					params = paramInjector()
					params.setKeys(child.wildcard.paramKeys)
					params.appendValue(path[longest:])
					buf.WriteString(path[longest:])
					return child.wildcard, params
				}

				return nil, params
			} else {
				// There are more segments to cover in the searching path.

				// Traverse the child node recursively until a match is found.
				var dfsChild *node
				if dfsChild, params = child.caseInsensitiveSearchRecursion(path[len(child.prefix):], params, paramInjector, buf); dfsChild != nil && dfsChild.handler != nil {
					// Found a matching node with a registered handler.
					buf.WriteString(child.prefix)
					return dfsChild, params
				}
			}
		}
	}

	// Didn't find a matching node.

	// We could try swapping if we haven't swapped already...
	if !swappedChild {
		if sc, swapped := swapCase(path[0]); swapped {
			child = n.findCandidateByCharAndSize(sc, len(path))
			swappedChild = true
			goto TraverseChild
		}
	}

	// Fallback to param node.
	if n.param != nil {
		// Check if more segments are left to cover in the searching path.
		if idx := strings.IndexByte(path, '/'); idx == -1 {

			// No more segments in the path.
			// Dead end #3
			if n.param.handler != nil {
				params = paramInjector()
				params.setKeys(n.param.paramKeys) // Param node would always have paramKeys.
				params.appendValue(path)
				buf.WriteString(path)
				return n.param, params
			}

			// The param node might have children who have handlers but no need to explore them
			// since the searching path has no more segments left to cover.
			// Thus, fallback to the wildcard node.

		} else {

			// Traverse the param node until all the segments are exhausted.
			if child, params = n.param.caseInsensitiveSearchRecursion(path[idx:], params, paramInjector, buf); child != nil && child.handler != nil {
				params.appendValue(path[:idx])
				buf.WriteString(path[:idx])
				return child, params
			}
		}
	}

	// Fallback to wildcard node.
	//
	// This also facilitates to fall back to the nearest wildcard node in the recursion stack when no match is found.
	//// No need to perform nil check for handler and paramKeys here
	//// since a wildcard node must always have a handler and paramKeys.
	//
	// Dead end #4
	if n.wildcard != nil {
		params = paramInjector()
		params.setKeys(n.wildcard.paramKeys)
		params.appendValue(path)
		buf.WriteString(path)
		return n.wildcard, params
	}

	return nil, params
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
	for ; i < max; i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	return i
}

func longestPrefixCaseInsensitive(s1, s2 string) int {
	max := len(s1)
	if len(s2) < max {
		max = len(s2)
	}

	i := 0
	for ; i < max; i++ {
		if s1[i] != s2[i] {
			if sc, swapped := swapCase(s2[i]); swapped && s1[i] == sc {
				continue
			}
			return i
		}
	}
	return i
}

func swapCase(r uint8) (uint8, bool) {
	if r < 'A' || r > 'z' || r > 'Z' && r < 'a' {
		return r, false
	}

	isLower := r >= 'a' && r <= 'z'

	if isLower {
		r -= 'a' - 'A'
	} else {
		r += 'a' - 'A'
	}

	return r, true
}
