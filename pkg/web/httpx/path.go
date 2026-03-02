package httpx

import (
	"fmt"
	"strings"
)

func BuildPath(method string, path string) string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", method, path))
}

func JoinPaths(base string, path string) string {
	switch {
	case base == "" && path == "":
		return ""
	case base == "":
		return cleanSlash(path)
	case path == "":
		return cleanSlash(base)
	default:
		return cleanSlash(base) + cleanSlash(path)
	}
}

func cleanSlash(p string) string {
	if p == "" || p == "/" {
		return ""
	}

	if p[0] != '/' {
		p = "/" + p
	}

	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}
