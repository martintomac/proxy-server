package proxy

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticHandler_ServeHTTP(t *testing.T) {
	handler := &StaticHandler{message: "Hello there!"}
	responseRecorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/static", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code, "Should be status code 200")
	assert.Equal(t, "Hello there!", responseRecorder.Body.String(), "Should be expected body response")
}

func TestNotFoundHandler_ServeHTTP(t *testing.T) {
	handler := &NotFoundHandler{}
	responseRecorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/example", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "Test-Agent")

	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusNotFound, responseRecorder.Code, "Should be status code 404")
}

func TestDebugHandler_ServeHTTP(t *testing.T) {
	handler := &DebugHandler{}
	responseRecorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/example", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "Test-Agent")

	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code, "Should be status code 200")

	contentType := responseRecorder.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType, "Should be json Content-Type")

	expectedBody := `{"method":"GET","url":"/example","headers":{"User-Agent":["Test-Agent"]},"body":""}` + "\n"
	assert.Equal(t, expectedBody, responseRecorder.Body.String(), "Should be expected json body response")
}

func TestEchoHandler_ServeHTTP(t *testing.T) {
	handler := &EchoHandler{}
	responseRecorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "https://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "Test-Agent")

	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code, "Should be status code 200")

	expectedBody := "" +
		"GET / HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"User-Agent: Test-Agent\r\n" +
		"\r\n"
	assert.Equal(t, expectedBody, responseRecorder.Body.String(), "Should be expected body response")

}

func TestChaosHandler_ServeHTTP(t *testing.T) {
	handler := &ChaosHandler{
		Handler:       &StaticHandler{message: "Hello there!"},
		FailureChance: 0.2,
		rand:          rand.New(rand.NewSource(0)),
	}

	req, err := http.NewRequest("GET", "https://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// based on seed these should return OK
	for i := 0; i < 3; i++ {
		responseRecorder := httptest.NewRecorder()

		handler.ServeHTTP(responseRecorder, req)

		equal := assert.Equal(t, http.StatusOK, responseRecorder.Code, "Should be status code 200")
		if !equal {
			return
		}
	}

	// based on seed this should fail
	responseRecorder := httptest.NewRecorder()

	handler.ServeHTTP(responseRecorder, req)

	equal := assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code, "Should be status code 500")
	if !equal {
		return
	}
}
