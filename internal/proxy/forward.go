package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ForwardHandler struct {
	URL    url.URL
	Client *http.Client
}

func NewForwardHandler(targetURL string) (*ForwardHandler, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	forwardHandler := ForwardHandler{
		URL: *u,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	return &forwardHandler, nil
}

func (h *ForwardHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	newReq, err := http.NewRequestWithContext(ctx, r.Method, h.URL.String(), bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			newReq.Header.Add(name, value)
		}
	}

	resp, err := h.Client.Do(newReq)
	if err != nil {
		log.Printf("Error forwarding request: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response: %v", err)
	}
}
