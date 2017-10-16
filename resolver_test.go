package main

import "testing"

func TestResolverParser(t *testing.T) {

	cases := []struct {
		val      string
		resolver string
		path     string
	}{
		{
			// protocol-ish path givs us a resolver
			"foo://baz",
			"foo",
			"baz",
		},
		{
			// non-protocol-ish paths give us blank resolver
			"foo/baz",
			"",
			"foo/baz",
		},
	}

	for _, tc := range cases {
		t.Run(tc.val, func(t *testing.T) {
			r, p := ParseResolverPath(tc.val)
			if tc.resolver != r {
				t.Errorf("expected resolver %s got %s", tc.resolver, r)
			}
			if tc.path != p {
				t.Errorf("expected path %s got %s", tc.path, p)
			}
		})
	}
}
