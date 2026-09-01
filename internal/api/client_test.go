package api

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientGet(t *testing.T) {
	t.Parallel()

	client, err := New("test-key", "https://api.example.test", &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/manage/v1/clients" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("x-api-key = %q", got)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("page = %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"items":[]}`)),
			Header:     make(http.Header),
		}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	body, err := client.Get(context.Background(), "/manage/v1/clients", url.Values{"page": {"2"}})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(body), `{"items":[]}`; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestClientRequiresAPIKey(t *testing.T) {
	t.Parallel()
	if _, err := New("", DefaultBaseURL, nil); err == nil {
		t.Fatal("New() error = nil")
	}
}

func TestClientRejectsPathTraversalBeforeRequest(t *testing.T) {
	t.Parallel()

	client, err := New("test-key", "https://api.example.test", &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("request must not be sent")
		return nil, nil
	})})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"/manage/v1/../../seo-tools/api",
		"/manage/v1/%252e%252e/seo-tools/api",
	} {
		_, err = client.Get(context.Background(), path, nil)
		if err == nil || !strings.Contains(err.Error(), "traversal") {
			t.Fatalf("Get(%q) error = %v, want traversal error", path, err)
		}
	}
}
