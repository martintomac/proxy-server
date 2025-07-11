package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForwardHandler_ServeHTTP(t *testing.T) {
	t.Run("successful forward request", func(t *testing.T) {
		// Create a mock target server
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify headers were forwarded
			assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"), "Expected header not found")

			// Verify body was forwarded
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, "test body", string(body), "Expected body not forwarded correctly")

			// Return a response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"message": "success"}`))
			if err != nil {
				assert.NoError(t, err, "Failed to write response")
			}
		}))
		defer targetServer.Close()

		handler, err := NewForwardHandler(targetServer.URL)
		require.NoError(t, err, "Failed to create ForwardHandler")

		req := httptest.NewRequest("POST", "/test", strings.NewReader("test body"))
		req.Header.Set("X-Test-Header", "test-value")
		req.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code to match")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Expected Content-Type header to be forwarded")

		expectedBody := `{"message": "success"}`
		assert.Equal(t, expectedBody, w.Body.String(), "Expected body to match")
	})

	t.Run("handles target server error", func(t *testing.T) {
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("Internal Server Error"))
			if err != nil {
				assert.NoError(t, err, "Failed to write response")
			}
		}))
		defer targetServer.Close()

		handler, err := NewForwardHandler(targetServer.URL)
		require.NoError(t, err, "Failed to create ForwardHandler")

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status code to match")
	})

	t.Run("handles unreachable target", func(t *testing.T) {
		handler, err := NewForwardHandler("http://localhost:99999")
		require.NoError(t, err, "Failed to create ForwardHandler")

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code, "Expected status code to match")
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, method, r.Method, "Expected HTTP method to match")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("OK"))
					if err != nil {
						assert.NoError(t, err, "Failed to write response")
					}
				}))
				defer targetServer.Close()

				handler, err := NewForwardHandler(targetServer.URL)
				require.NoError(t, err, "Failed to create ForwardHandler")

				var body io.Reader
				if method == "POST" || method == "PUT" || method == "PATCH" {
					body = strings.NewReader("test data")
				}

				req := httptest.NewRequest(method, "/test", body)
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code, "Expected status code to match")
			})
		}
	})

	t.Run("handles request with large body", func(t *testing.T) {
		largeBody := strings.Repeat("a", 1024*1024) // large body (1MB)

		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, len(largeBody), len(body), "Expected body length to match")
			w.WriteHeader(http.StatusOK)
		}))
		defer targetServer.Close()

		handler, err := NewForwardHandler(targetServer.URL)
		require.NoError(t, err, "Failed to create ForwardHandler")

		req := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code to match")
	})

	t.Run("handles context timeout", func(t *testing.T) {
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(500 * time.Millisecond) // Longer than the 200 ms timeout
			w.WriteHeader(http.StatusOK)
		}))
		defer targetServer.Close()

		targetUrl, err := url.Parse(targetServer.URL)
		if err != nil {
			assert.NoError(t, err, "Failed to parse target URL")
		}
		handler := &ForwardHandler{
			URL:    *targetUrl,
			Client: &http.Client{Timeout: 200 * time.Millisecond},
		}

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code, "Expected status code to match")
	})

	t.Run("forwards multiple headers correctly", func(t *testing.T) {
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"), "Authorization header not forwarded")
			assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"), "Custom header not forwarded")

			w.Header().Set("X-Response-Header", "response-value")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			if err != nil {
				assert.NoError(t, err, "Failed to write response")
			}
		}))
		defer targetServer.Close()

		handler, err := NewForwardHandler(targetServer.URL)
		if err != nil {
			t.Fatalf("Failed to create ForwardHandler: %v", err)
		}

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer token")
		req.Header.Set("X-Custom-Header", "custom-value")

		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code to match")
		assert.Equal(t, "response-value", w.Header().Get("X-Response-Header"), "Response header not forwarded")
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"), "Cache-Control header not forwarded")
	})
}

func TestNewForwardHandler(t *testing.T) {
	t.Run("creates handler with valid URL", func(t *testing.T) {
		handler, err := NewForwardHandler("https://example.com")
		require.NoError(t, err, "Expected no error")

		assert.Equal(t, "https://example.com", handler.URL.String(), "Expected URL to be set correctly")
		assert.NotNil(t, handler.Client, "Expected client to be initialized")
		assert.Equal(t, 30*time.Second, handler.Client.Timeout, "Expected timeout to be 30 seconds")
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		_, err := NewForwardHandler(":/invalid-url")
		assert.Error(t, err, "Expected error for invalid URL")
	})
}
