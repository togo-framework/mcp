package togomcp

import "testing"

// list_routes must list every route make:resource registers, PUT included
// (it was missing although the OpenAPI spec has "Update a Post").
func TestRoutesFor(t *testing.T) {
	manifest := "resources:\n  - name: Post\n    table: posts\n  - name: Comment\n    table: comments\n"
	want := "GET    /api/posts\nPOST   /api/posts\nGET    /api/posts/{id}\nPUT    /api/posts/{id}\nDELETE /api/posts/{id}\n" +
		"GET    /api/comments\nPOST   /api/comments\nGET    /api/comments/{id}\nPUT    /api/comments/{id}\nDELETE /api/comments/{id}\n"
	if got := routesFor(manifest); got != want {
		t.Errorf("routesFor =\n%s\nwant\n%s", got, want)
	}
	if got := routesFor(""); got != "(no resources; routes appear after make:resource)" {
		t.Errorf("routesFor(empty) = %q", got)
	}
}
