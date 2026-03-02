package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildPath tests the BuildPath function
func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "GET request",
			method:   "GET",
			path:     "/api/users",
			expected: "GET /api/users",
		},
		{
			name:     "POST request",
			method:   "POST",
			path:     "/api/users",
			expected: "POST /api/users",
		},
		{
			name:     "DELETE request with ID",
			method:   "DELETE",
			path:     "/api/users/123",
			expected: "DELETE /api/users/123",
		},
		{
			name:     "empty method",
			method:   "",
			path:     "/api/users",
			expected: "/api/users",
		},
		{
			name:     "empty path",
			method:   "GET",
			path:     "",
			expected: "GET",
		},
		{
			name:     "both empty",
			method:   "",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPath(tt.method, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestJoinPaths tests the JoinPaths function
func TestJoinPaths(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		path     string
		expected string
	}{
		{
			name:     "both empty",
			base:     "",
			path:     "",
			expected: "",
		},
		{
			name:     "empty base",
			base:     "",
			path:     "/api/users",
			expected: "/api/users",
		},
		{
			name:     "empty path",
			base:     "/api",
			path:     "",
			expected: "/api",
		},
		{
			name:     "both with leading slashes",
			base:     "/api",
			path:     "/users",
			expected: "/api/users",
		},
		{
			name:     "base without leading slash",
			base:     "api",
			path:     "/users",
			expected: "/api/users",
		},
		{
			name:     "path without leading slash",
			base:     "/api",
			path:     "users",
			expected: "/api/users",
		},
		{
			name:     "neither with leading slash",
			base:     "api",
			path:     "users",
			expected: "/api/users",
		},
		{
			name:     "base with trailing slash",
			base:     "/api/",
			path:     "/users",
			expected: "/api/users",
		},
		{
			name:     "path with trailing slash",
			base:     "/api",
			path:     "/users/",
			expected: "/api/users",
		},
		{
			name:     "both with trailing slashes",
			base:     "/api/",
			path:     "/users/",
			expected: "/api/users",
		},
		{
			name:     "complex path join",
			base:     "v1/api",
			path:     "users/123",
			expected: "/v1/api/users/123",
		},
		{
			name:     "base is single slash",
			base:     "/",
			path:     "/users",
			expected: "/users",
		},
		{
			name:     "path is single slash",
			base:     "/api",
			path:     "/",
			expected: "/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinPaths(tt.base, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCleanSlash tests the cleanSlash function
func TestCleanSlash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single slash",
			input:    "/",
			expected: "",
		},
		{
			name:     "path with leading slash",
			input:    "/api/users",
			expected: "/api/users",
		},
		{
			name:     "path without leading slash",
			input:    "api/users",
			expected: "/api/users",
		},
		{
			name:     "path with trailing slash",
			input:    "/api/users/",
			expected: "/api/users",
		},
		{
			name:     "path with both leading and trailing slash",
			input:    "/api/users/",
			expected: "/api/users",
		},
		{
			name:     "path without leading slash but with trailing slash",
			input:    "api/users/",
			expected: "/api/users",
		},
		{
			name:     "multiple trailing slashes",
			input:    "/api/users//",
			expected: "/api/users/",
		},
		{
			name:     "single character path",
			input:    "a",
			expected: "/a",
		},
		{
			name:     "single character with leading slash",
			input:    "/a",
			expected: "/a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSlash(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
