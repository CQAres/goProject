package trie

import (
	"testing"
)

// ============ Node Tests ============

func TestNodeInsertSingle(t *testing.T) {
	root := &Node{}
	root.Insert("/user", []string{"user"}, 0)

	if len(root.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children))
	}
	if root.Children[0].Part != "user" {
		t.Errorf("expected part 'user', got '%s'", root.Children[0].Part)
	}
}

func TestNodeInsertMultiple(t *testing.T) {
	root := &Node{}
	root.Insert("/user/profile", []string{"user", "profile"}, 0)

	// Find user node
	userNode := root.Children[0]
	if userNode.Part != "user" {
		t.Errorf("expected 'user', got '%s'", userNode.Part)
	}

	// Find profile node
	profileNode := userNode.Children[0]
	if profileNode.Part != "profile" {
		t.Errorf("expected 'profile', got '%s'", profileNode.Part)
	}
	if profileNode.Pattern != "/user/profile" {
		t.Errorf("expected pattern '/user/profile', got '%s'", profileNode.Pattern)
	}
}

func TestNodeInsertParam(t *testing.T) {
	root := &Node{}
	root.Insert("/user/:id", []string{"user", ":id"}, 0)

	userNode := root.Children[0]
	if userNode.Part != "user" {
		t.Errorf("expected 'user', got '%s'", userNode.Part)
	}

	idNode := userNode.Children[0]
	if idNode.Part != ":id" {
		t.Errorf("expected ':id', got '%s'", idNode.Part)
	}
	if !idNode.IsWild {
		t.Errorf("expected IsWild=true for :id")
	}
}

func TestNodeInsertWildcard(t *testing.T) {
	root := &Node{}
	root.Insert("/static/*filepath", []string{"static", "*filepath"}, 0)

	staticNode := root.Children[0]
	if staticNode.Part != "static" {
		t.Errorf("expected 'static', got '%s'", staticNode.Part)
	}

	filepathNode := staticNode.Children[0]
	if filepathNode.Part != "*filepath" {
		t.Errorf("expected '*filepath', got '%s'", filepathNode.Part)
	}
	if !filepathNode.IsWild {
		t.Errorf("expected IsWild=true for *filepath")
	}
}

// ============ Search Tests ============

func TestNodeSearchExact(t *testing.T) {
	root := &Node{}
	root.Insert("/user/profile", []string{"user", "profile"}, 0)

	result := root.Search([]string{"user", "profile"}, 0)
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Pattern != "/user/profile" {
		t.Errorf("expected pattern '/user/profile', got '%s'", result.Pattern)
	}
}

func TestNodeSearchParam(t *testing.T) {
	root := &Node{}
	root.Insert("/user/:id", []string{"user", ":id"}, 0)

	result := root.Search([]string{"user", "123"}, 0)
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Pattern != "/user/:id" {
		t.Errorf("expected pattern '/user/:id', got '%s'", result.Pattern)
	}
}

func TestNodeSearchWildcard(t *testing.T) {
	root := &Node{}
	root.Insert("/static/*filepath", []string{"static", "*filepath"}, 0)

	result := root.Search([]string{"static", "css", "style.css"}, 0)
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Pattern != "/static/*filepath" {
		t.Errorf("expected pattern '/static/*filepath', got '%s'", result.Pattern)
	}
}

func TestNodeSearchNotFound(t *testing.T) {
	root := &Node{}
	root.Insert("/user/profile", []string{"user", "profile"}, 0)

	result := root.Search([]string{"post", "123"}, 0)
	if result != nil {
		t.Errorf("expected nil for non-matching path, got %s", result.Pattern)
	}
}

func TestNodeSearchPartial(t *testing.T) {
	root := &Node{}
	root.Insert("/user/profile", []string{"user", "profile"}, 0)

	// Search with only partial path
	result := root.Search([]string{"user"}, 0)
	if result != nil {
		t.Errorf("expected nil for partial path, got %s", result.Pattern)
	}
}

// ============ MatchChild Tests ============

func TestNodeMatchChild(t *testing.T) {
	root := &Node{Part: "root"}
	child1 := &Node{Part: "user"}
	child2 := &Node{Part: ":id", IsWild: true}
	root.Children = []*Node{child1, child2}

	result := root.MatchChild("user")
	if result != child1 {
		t.Errorf("expected to match 'user' node")
	}

	result = root.MatchChild("123")
	if result != child2 {
		t.Errorf("expected to match ':id' wildcard node")
	}

	// Wildcard nodes match any part, so "nonexistent" matches :id
	result = root.MatchChild("nonexistent")
	if result != child2 {
		t.Errorf("expected to match ':id' wildcard node for 'nonexistent'")
	}

	// Test with no wildcard: only exact match works
	root2 := &Node{Part: "root"}
	child3 := &Node{Part: "post"}
	root2.Children = []*Node{child3}

	result = root2.MatchChild("nonexistent")
	if result != nil {
		t.Errorf("expected nil for non-existent child without wildcard")
	}
}

func TestNodeMatchChildren(t *testing.T) {
	root := &Node{Part: "root"}
	child1 := &Node{Part: ":id", IsWild: true}
	child2 := &Node{Part: "*path", IsWild: true}
	root.Children = []*Node{child1, child2}

	results := root.MatchChildren("anything")
	if len(results) != 2 {
		t.Errorf("expected 2 matches, got %d", len(results))
	}
}

// ============ Complex Routes Tests ============

func TestMultipleRoutes(t *testing.T) {
	root := &Node{}

	routes := []string{
		"/",
		"/user",
		"/user/:id",
		"/user/:id/profile",
		"/user/:id/post/:pid",
		"/static/*filepath",
		"/api/v1/users",
		"/api/v1/posts/:pid",
	}

	for _, route := range routes {
		parts := parsePattern(route)
		if parts == nil {
			parts = []string{}
		}
		root.Insert(route, parts, 0)
	}

	tests := []struct {
		path     string
		expected string
		found    bool
	}{
		{"/", "/", true},
		{"/user", "/user", true},
		{"/user/123", "/user/:id", true},
		{"/user/123/profile", "/user/:id/profile", true},
		{"/user/123/post/456", "/user/:id/post/:pid", true},
		{"/static/css/style.css", "/static/*filepath", true},
		{"/api/v1/users", "/api/v1/users", true},
		{"/api/v1/posts/789", "/api/v1/posts/:pid", true},
		{"/unknown", "", false},
	}

	for _, tt := range tests {
		parts := parsePattern(tt.path)
		if parts == nil {
			parts = []string{}
		}
		result := root.Search(parts, 0)

		if tt.found {
			if result == nil {
				t.Errorf("path %s: expected to find %s, got nil", tt.path, tt.expected)
			} else if result.Pattern != tt.expected {
				t.Errorf("path %s: expected %s, got %s", tt.path, tt.expected, result.Pattern)
			}
		} else {
			if result != nil {
				t.Errorf("path %s: expected not found, got %s", tt.path, result.Pattern)
			}
		}
	}
}

// ============ Helper Function Tests ============

func parsePattern(pattern string) []string {
	// Simple implementation matching the router's parsePattern
	if pattern == "/" {
		return nil
	}

	parts := make([]string, 0)
	start := 0
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '/' {
			if i > start {
				parts = append(parts, pattern[start:i])
			}
			start = i + 1
		}
	}
	if start < len(pattern) {
		parts = append(parts, pattern[start:])
	}
	return parts
}

// ============ Edge Cases ============

func TestEmptyPattern(t *testing.T) {
	root := &Node{}
	root.Insert("/", []string{}, 0)

	result := root.Search([]string{}, 0)
	if result == nil {
		t.Error("expected result for root '/', got nil")
	}
}

func TestConsecutiveSlashes(t *testing.T) {
	root := &Node{}

	// The actual parsePattern would handle this differently
	// This tests the trie structure behavior
	root.Children = []*Node{
		{Part: "user"},
	}

	result := root.MatchChild("user")
	if result == nil {
		t.Error("expected to match 'user'")
	}
}

func TestWildcardAtEnd(t *testing.T) {
	root := &Node{}
	root.Insert("/a/*", []string{"a", "*"}, 0)

	// Search with extra parts
	result := root.Search([]string{"a", "b", "c", "d"}, 0)
	if result == nil {
		t.Error("expected to match wildcard route")
	}
}
