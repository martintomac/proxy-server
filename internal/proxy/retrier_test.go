package proxy

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestRetrierHandler_ServeHTTP(t *testing.T) {
	t.Run("should retry once on 500 and succeed with 200", func(t *testing.T) {
		h := &RetrierHandler{
			Handler:     &MockHandler{statusCodes: []int{http.StatusInternalServerError, http.StatusOK}},
			RetryPolicy: &RetryOnNon2xxRetryPolicy{},
			Retries:     1,
		}

		brw := NewBufferedResponseWriter()

		request, err := http.NewRequest("GET", "http://localhost:8080", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.ServeHTTP(brw, request)

		assert.Equal(t, 2, h.Handler.(*MockHandler).invocations, "Expected handler to be invoked twice")
		assert.Equal(t, http.StatusOK, brw.statusCode, "Expected status code to be 200")
	})
	t.Run("should retry once on 500 and still fail", func(t *testing.T) {
		h := &RetrierHandler{
			Handler:     &MockHandler{statusCodes: []int{http.StatusInternalServerError}},
			RetryPolicy: &RetryOnNon2xxRetryPolicy{},
			Retries:     1,
		}

		brw := NewBufferedResponseWriter()

		request, err := http.NewRequest("GET", "http://localhost:8080", nil)
		if err != nil {
			t.Fatal(err)
		}

		h.ServeHTTP(brw, request)

		assert.Equal(t, 2, h.Handler.(*MockHandler).invocations, "Expected handler to be invoked twice")
		assert.Equal(t, http.StatusInternalServerError, brw.statusCode, "Expected status code to be 500")
	})
}

type MockHandler struct {
	statusCodes []int
	invocations int
}

func (h *MockHandler) ServeHTTP(
	w http.ResponseWriter,
	_ *http.Request,
) {
	defer func() {
		h.invocations++
	}()

	numOfStatusCodes := len(h.statusCodes)

	var statusCode int
	switch {
	case numOfStatusCodes == 0:
		statusCode = http.StatusOK
	case h.invocations < numOfStatusCodes:
		statusCode = h.statusCodes[h.invocations]
	default:
		statusCode = h.statusCodes[numOfStatusCodes-1]
	}

	w.WriteHeader(statusCode)
}
