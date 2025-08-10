package proxy

import (
	"fmt"
	"net/http"
	"strings"
)

type MatchingPredicate interface {
	match(r *http.Request) bool
}

type RequestPredicate struct {
	Method *MethodPredicate
	Path   *PathPredicate
	Header *HeaderPredicate
	Query  *QueryPredicate
}

func (p *RequestPredicate) match(r *http.Request) bool {
	hasNonMatchingPredicate :=
		p.Method != nil && !p.Method.match(r) ||
			p.Path != nil && !p.Path.match(r) ||
			p.Header != nil && !p.Header.match(r) ||
			p.Query != nil && !p.Query.match(r)
	return !hasNonMatchingPredicate
}

type PathPredicate struct {
	parts []string
}

func NewPathPredicate(path string) *PathPredicate {
	return &PathPredicate{
		parts: parsePath(path),
	}
}

func parsePath(path string) []string {
	rawParts := splitToParts(path)

	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if isVariablePart(part) {
			parts = append(parts, "{}")
		} else {
			parts = append(parts, part)
		}
	}

	return parts
}

func splitToParts(path string) []string {
	if path == "" || path == "/" {
		return []string{}
	}
	parts := strings.Split(path, "/")
	if parts[0] == "" {
		parts = parts[1:]
	}
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func isVariablePart(part string) bool {
	return strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}")
}

func (p *PathPredicate) match(r *http.Request) bool {
	if len(p.parts) == 0 {
		return true
	}

	reqParts := splitToParts(r.URL.Path)
	for i, part := range p.parts {
		if i >= len(reqParts) {
			return false
		}
		if part == "{}" {
			continue
		}
		if part != reqParts[i] {
			return false
		}
	}

	return true
}

type MethodPredicate struct {
	Method string
}

func NewMethodPredicate(method string) (*MethodPredicate, error) {
	allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, allowedMethod := range allowedMethods {
		if allowedMethod == method {
			return &MethodPredicate{
				Method: allowedMethod,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid method: %v", method)
}

func (p *MethodPredicate) match(r *http.Request) bool {
	return r.Method == p.Method
}

type HeaderPredicate struct {
	Name  string
	Value string
}

func NewHeaderPredicate(name, value string) (*HeaderPredicate, error) {
	if name == "" {
		return nil, fmt.Errorf("header name is empty")
	}
	return &HeaderPredicate{
		Name:  name,
		Value: value,
	}, nil
}

func (p *HeaderPredicate) match(r *http.Request) bool {
	for _, value := range r.Header.Values(p.Name) {
		if value == p.Value {
			return true
		}
	}
	return false
}

type QueryPredicate struct {
	Name  string
	Value string
}

func NewQueryPredicate(name, value string) (*QueryPredicate, error) {
	if name == "" {
		return nil, fmt.Errorf("query name is empty")
	}
	return &QueryPredicate{
		Name:  name,
		Value: value,
	}, nil
}

func (p *QueryPredicate) match(r *http.Request) bool {
	values := r.URL.Query()[p.Name]
	for _, value := range values {
		if value == p.Value {
			return true
		}
	}
	return false
}
