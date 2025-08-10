package proxy

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPathPredicate(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *PathPredicate
	}{
		{
			name: "emptyPath",
			args: args{path: ""},
			want: &PathPredicate{
				parts: []string{},
			},
		},
		{
			name: "justSlash",
			args: args{path: "/"},
			want: &PathPredicate{
				parts: []string{},
			},
		},
		{
			name: "simplePath",
			args: args{path: "/hello"},
			want: &PathPredicate{
				parts: []string{"hello"},
			},
		},
		{
			name: "simpleMultiPartPath",
			args: args{path: "/hello/world"},
			want: &PathPredicate{
				parts: []string{"hello", "world"},
			},
		},
		{
			name: "simplePath_withTrailingSlash",
			args: args{path: "/hello/"},
			want: &PathPredicate{
				parts: []string{"hello"},
			},
		},
		{
			name: "variablePart",
			args: args{path: "/{something}"},
			want: &PathPredicate{
				parts: []string{"{}"},
			},
		},
		{
			name: "variablePart_withLatterPart",
			args: args{path: "/{something}/world"},
			want: &PathPredicate{
				parts: []string{"{}", "world"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewPathPredicate(tt.args.path), "NewPathPredicate(%v)", tt.args.path)
		})
	}
}

func TestPathPredicate_match(t *testing.T) {
	type fields struct {
		parts []string
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "everythingMatches",
			fields: fields{
				parts: []string{},
			},
			args: args{r: newGetRequest("/anything")},
			want: true,
		},
		{
			name: "singlePart_exactMatch",
			fields: fields{
				parts: []string{"hello"},
			},
			args: args{r: newGetRequest("/hello")},
			want: true,
		},
		{
			name: "singlePart_partMismatch",
			fields: fields{
				parts: []string{"hello"},
			},
			args: args{r: newGetRequest("/world")},
			want: false,
		},
		{
			name: "singlePart_prefixMatch",
			fields: fields{
				parts: []string{"hello"},
			},
			args: args{r: newGetRequest("/hello/world")},
			want: true,
		},
		{
			name: "singlePart_prefixMismatch",
			fields: fields{
				parts: []string{"hello"},
			},
			args: args{r: newGetRequest("/any/world")},
			want: false,
		},
		{
			name: "singlePart_variableMatch",
			fields: fields{
				parts: []string{"{}"},
			},
			args: args{r: newGetRequest("/anything")},
			want: true,
		},
		{
			name: "singlePart_variableMismatch",
			fields: fields{
				parts: []string{"{}"},
			},
			args: args{r: newGetRequest("/")},
			want: false,
		},
		{
			name: "multiPart_exactMatch",
			fields: fields{
				parts: []string{"hello", "world"},
			},
			args: args{r: newGetRequest("/hello/world")},
			want: true,
		},
		{
			name: "multiPart_withVariable",
			fields: fields{
				parts: []string{"hello", "{}"},
			},
			args: args{r: newGetRequest("/hello/anything")},
			want: true,
		},
		{
			name: "multiPart_withVariableMismatch",
			fields: fields{
				parts: []string{"hello", "{}", "world"},
			},
			args: args{r: newGetRequest("/hello/something")},
			want: false,
		},
		{
			name: "multiPart_multipleVariables",
			fields: fields{
				parts: []string{"{}", "test", "{}"},
			},
			args: args{r: newGetRequest("/any/test/value")},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PathPredicate{
				parts: tt.fields.parts,
			}
			assert.Equalf(t, tt.want, p.match(tt.args.r), "match(%v)", tt.args.r)
		})
	}
}

func TestNewMethodPredicate(t *testing.T) {
	type args struct {
		method string
	}
	type want struct {
		predicate *MethodPredicate
		err       error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "getMethodPredicate",
			args: args{method: "GET"},
			want: want{&MethodPredicate{Method: "GET"}, nil},
		},
		{
			name: "postMethodPredicate",
			args: args{method: "POST"},
			want: want{&MethodPredicate{Method: "POST"}, nil},
		},
		{
			name: "invalidMethodPredicate",
			args: args{method: "INVALID"},
			want: want{nil, fmt.Errorf("invalid method: INVALID")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMethodPredicate(tt.args.method)
			assert.Equalf(t, tt.want, want{got, err}, "NewMethodPredicate(%v)", tt.args.method)
		})
	}
}

func TestHeaderPredicate_match(t *testing.T) {
	type fields struct {
		Name  string
		Value string
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "matchingHeaderValue",
			fields: fields{
				Name:  "X-Test-Header",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/")
				req.Header.Add("X-Test-Header", "value")
				return req
			}()},
			want: true,
		},
		{
			name: "nonMatchingHeaderValue",
			fields: fields{
				Name:  "X-Test-Header",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/")
				req.Header.Add("X-Test-Header", "wrongValue")
				return req
			}()},
			want: false,
		},
		{
			name: "headerNameNotPresent",
			fields: fields{
				Name:  "X-Test-Header",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/")
				return req
			}()},
			want: false,
		},
		{
			name: "matchingAmongMultipleHeaders",
			fields: fields{
				Name:  "X-Test-Header",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/")
				req.Header.Add("X-Test-Header", "value")
				req.Header.Add("X-Test-Header", "anotherValue")
				return req
			}()},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HeaderPredicate{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			assert.Equalf(t, tt.want, p.match(tt.args.r), "match(%v)", tt.args.r)
		})
	}
}

func TestNewHeaderPredicate(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	type want struct {
		predicate *HeaderPredicate
		err       error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "validHeaderPredicate",
			args: args{
				name:  "X-Test-Header",
				value: "value",
			},
			want: want{
				predicate: &HeaderPredicate{Name: "X-Test-Header", Value: "value"},
				err:       nil,
			},
		},
		{
			name: "emptyHeaderName",
			args: args{
				name:  "",
				value: "value",
			},
			want: want{
				predicate: nil,
				err:       fmt.Errorf("header name is empty"),
			},
		},
		{
			name: "emptyHeaderValue",
			args: args{
				name:  "X-Test-Header",
				value: "",
			},
			want: want{
				predicate: &HeaderPredicate{Name: "X-Test-Header", Value: ""},
				err:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHeaderPredicate(tt.args.name, tt.args.value)
			assert.Equalf(t, tt.want, want{got, err}, "NewHeaderPredicate(%v, %v)", tt.args.name, tt.args.value)
		})
	}
}

func TestQueryPredicate_match(t *testing.T) {
	type fields struct {
		Name  string
		Value string
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "exactMatch",
			fields: fields{
				Name:  "key",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/?key=value")
				return req
			}()},
			want: true,
		},
		{
			name: "mismatchValue",
			fields: fields{
				Name:  "key",
				Value: "expectedValue",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/?key=wrongValue")
				return req
			}()},
			want: false,
		},
		{
			name: "missingQueryParam",
			fields: fields{
				Name:  "key",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/?anotherKey=value")
				return req
			}()},
			want: false,
		},
		{
			name: "emptyQueryName",
			fields: fields{
				Name:  "",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/?key=value")
				return req
			}()},
			want: false,
		},
		{
			name: "emptyQueryValue",
			fields: fields{
				Name:  "key",
				Value: "",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/?key=")
				return req
			}()},
			want: true,
		},
		{
			name: "multipleQueryParams",
			fields: fields{
				Name:  "key",
				Value: "value",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/?key=value&anotherKey=anotherValue")
				return req
			}()},
			want: true,
		},
		{
			name: "multipleValuesForSameKey",
			fields: fields{
				Name:  "key",
				Value: "value2",
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/?key=value1&key=value2")
				return req
			}()},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &QueryPredicate{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			assert.Equalf(t, tt.want, p.match(tt.args.r), "match(%v)", tt.args.r)
		})
	}
}

func TestNewQueryPredicate(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	type want struct {
		predicate *QueryPredicate
		err       error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "validQueryPredicate",
			args: args{
				name:  "key",
				value: "value",
			},
			want: want{
				predicate: &QueryPredicate{Name: "key", Value: "value"},
				err:       nil,
			},
		},
		{
			name: "emptyQueryName",
			args: args{
				name:  "",
				value: "value",
			},
			want: want{
				predicate: nil,
				err:       fmt.Errorf("query name is empty"),
			},
		},
		{
			name: "emptyQueryValue",
			args: args{
				name:  "key",
				value: "",
			},
			want: want{
				predicate: &QueryPredicate{Name: "key", Value: ""},
				err:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewQueryPredicate(tt.args.name, tt.args.value)
			assert.Equalf(t, tt.want, want{got, err}, "NewQueryPredicate(%v, %v)", tt.args.name, tt.args.value)
		})
	}
}

func TestRequestPredicate_match(t *testing.T) {
	type fields struct {
		Method *MethodPredicate
		Path   *PathPredicate
		Header *HeaderPredicate
		Query  *QueryPredicate
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "allNilPredicates",
			fields: fields{
				Method: nil,
				Path:   nil,
				Header: nil,
				Query:  nil,
			},
			args: args{r: newGetRequest("/")},
			want: true,
		},
		{
			name: "methodPredicateMatch",
			fields: fields{
				Method: &MethodPredicate{Method: "GET"},
				Path:   nil,
				Header: nil,
				Query:  nil,
			},
			args: args{r: newGetRequest("/")},
			want: true,
		},
		{
			name: "methodPredicateMismatch",
			fields: fields{
				Method: &MethodPredicate{Method: "POST"},
				Path:   nil,
				Header: nil,
				Query:  nil,
			},
			args: args{r: newGetRequest("/")},
			want: false,
		},
		{
			name: "pathPredicateMatch",
			fields: fields{
				Method: nil,
				Path:   &PathPredicate{parts: []string{"hello"}},
				Header: nil,
				Query:  nil,
			},
			args: args{r: newGetRequest("/hello")},
			want: true,
		},
		{
			name: "pathPredicateMismatch",
			fields: fields{
				Method: nil,
				Path:   &PathPredicate{parts: []string{"hello"}},
				Header: nil,
				Query:  nil,
			},
			args: args{r: newGetRequest("/world")},
			want: false,
		},
		{
			name: "headerPredicateMatch",
			fields: fields{
				Method: nil,
				Path:   nil,
				Header: &HeaderPredicate{Name: "X-Test", Value: "value"},
				Query:  nil,
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/")
				req.Header.Add("X-Test", "value")
				return req
			}()},
			want: true,
		},
		{
			name: "headerPredicateMismatch",
			fields: fields{
				Method: nil,
				Path:   nil,
				Header: &HeaderPredicate{Name: "X-Test", Value: "value"},
				Query:  nil,
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/")
				req.Header.Add("X-Test", "wrongValue")
				return req
			}()},
			want: false,
		},
		{
			name: "queryPredicateMatch",
			fields: fields{
				Method: nil,
				Path:   nil,
				Header: nil,
				Query:  &QueryPredicate{Name: "key", Value: "value"},
			},
			args: args{r: newGetRequest("/?key=value")},
			want: true,
		},
		{
			name: "queryPredicateMismatch",
			fields: fields{
				Method: nil,
				Path:   nil,
				Header: nil,
				Query:  &QueryPredicate{Name: "key", Value: "value"},
			},
			args: args{r: newGetRequest("/?key=wrongValue")},
			want: false,
		},
		{
			name: "methodAndPathMatch",
			fields: fields{
				Method: &MethodPredicate{Method: "GET"},
				Path:   &PathPredicate{parts: []string{"hello"}},
				Header: nil,
				Query:  nil,
			},
			args: args{r: newGetRequest("/hello")},
			want: true,
		},
		{
			name: "methodAndPathMismatch",
			fields: fields{
				Method: &MethodPredicate{Method: "POST"},
				Path:   &PathPredicate{parts: []string{"hello"}},
				Header: nil,
				Query:  nil,
			},
			args: args{r: newGetRequest("/world")},
			want: false,
		},
		{
			name: "allPredicatesMatch",
			fields: fields{
				Method: &MethodPredicate{Method: "GET"},
				Path:   &PathPredicate{parts: []string{"hello", "world"}},
				Header: &HeaderPredicate{Name: "X-Test", Value: "value"},
				Query:  &QueryPredicate{Name: "key", Value: "value"},
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/hello/world?key=value")
				req.Header.Add("X-Test", "value")
				return req
			}()},
			want: true,
		},
		{
			name: "allPredicatesMismatch",
			fields: fields{
				Method: &MethodPredicate{Method: "POST"},
				Path:   &PathPredicate{parts: []string{"hello", "world"}},
				Header: &HeaderPredicate{Name: "X-Test", Value: "value"},
				Query:  &QueryPredicate{Name: "key", Value: "value"},
			},
			args: args{r: func() *http.Request {
				req := newGetRequest("/hello/universe?key=wrongValue")
				req.Header.Add("X-Test", "wrongValue")
				return req
			}()},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &RequestPredicate{
				Method: tt.fields.Method,
				Path:   tt.fields.Path,
				Header: tt.fields.Header,
				Query:  tt.fields.Query,
			}
			assert.Equalf(t, tt.want, p.match(tt.args.r), "match(%v)", tt.args.r)
		})
	}
}

func newGetRequest(path string) *http.Request {
	return newRequest("GET", path)
}

func newRequest(method, path string) *http.Request {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	return req
}
