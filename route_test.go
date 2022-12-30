package dune

import "testing"

type item struct {
	path     string
	segments []string
}

type testTable = []item

func TestRouteScanner(t *testing.T) {
	table := testTable{
		{path: "/blog/posts", segments: []string{"/blog/posts"}},
		{path: "/users/:id/action", segments: []string{"/users/", ":id", "/action"}},
		{path: "/assets/*dir", segments: []string{"/assets/", "*dir"}},
		{path: "/heroes/:name/:power", segments: []string{"/heroes/", ":name", "/", ":power"}},
	}

	for _, item := range table {
		r := newRouteScanner(item.path)
		for seg, i := r.next(), 0; seg != ""; seg, i = r.next(), i+1 {
			if seg != item.segments[i] {
				t.Errorf("expected: %s, got: %s", item.segments[i], seg)
			}
		}
	}
}
