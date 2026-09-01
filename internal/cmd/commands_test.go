package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestParseQuery(t *testing.T) {
	t.Parallel()
	got, err := parseQuery([]string{"page=2", "tag=one", "tag=two"})
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{"page": {"2"}, "tag": {"one", "two"}}
	if got.Encode() != want.Encode() {
		t.Fatalf("query = %q, want %q", got.Encode(), want.Encode())
	}
}

func TestParseQueryRejectsMissingEquals(t *testing.T) {
	t.Parallel()
	if _, err := parseQuery([]string{"page"}); err == nil {
		t.Fatal("parseQuery() error = nil")
	}
}

func TestListRejectsInvalidPageSize(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"clients", "list", "--per-page", "101"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "per-page must be between 1 and 100") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestHelpDoesNotExposeEnvironmentAPIKey(t *testing.T) {
	t.Setenv("BRIGHTLOCAL_API_KEY", "super-secret-key")
	var output bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "super-secret-key") {
		t.Fatal("help output exposes the API key")
	}
}

func TestNewReadOnlyCommandsRegistered(t *testing.T) {
	root := NewRootCmd()
	for _, path := range [][]string{
		{"rank-tracker", "reports", "list"},
		{"rank-tracker", "reports", "get"},
		{"rank-tracker", "reports", "history"},
		{"rank-tracker", "reports", "result"},
		{"search-grid", "reports", "list"},
		{"search-grid", "reports", "get"},
		{"search-grid", "runs", "list"},
		{"search-grid", "runs", "get"},
		{"search-grid", "rankings", "competitors"},
		{"search-grid", "rankings", "competitor"},
		{"reputation", "reports", "list"},
		{"reputation", "reports", "get"},
		{"reputation", "reports", "reviews"},
		{"citation-builder", "list"},
		{"citation-builder", "get"},
		{"reference", "time-options"},
		{"reference", "white-label-profiles"},
	} {
		if _, _, err := root.Find(path); err != nil {
			t.Errorf("command %q is not registered: %v", strings.Join(path, " "), err)
		}
	}
}

func TestNewPathCmdEscapesPathAndForwardsQueries(t *testing.T) {
	originalKey, originalBaseURL, originalHTTPClient := apiKey, baseURL, http.DefaultClient
	t.Cleanup(func() {
		apiKey, baseURL, http.DefaultClient = originalKey, originalBaseURL, originalHTTPClient
	})
	apiKey = "test-key"
	baseURL = "https://api.example.test"

	var gotPath string
	var gotQuery url.Values
	http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		gotPath = req.URL.EscapedPath()
		gotQuery = req.URL.Query()
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	cmd := newPathCmd("get <first> <second>", "test path", "/manage/v1/resources/%s/%s")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"a/b", "x y", "--query", "tag=one", "--query", "tag=two"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := gotPath, "/manage/v1/resources/a%2Fb/x%20y"; got != want {
		t.Fatalf("request path = %q, want %q", got, want)
	}
	if got, want := gotQuery["tag"], []string{"one", "two"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("tag query = %v, want %v", got, want)
	}
	if !strings.Contains(output.String(), `"ok": true`) {
		t.Fatalf("output = %q", output.String())
	}
}

func TestWriteRequiresConfirmBeforeSendingRequest(t *testing.T) {
	originalKey, originalBaseURL, originalHTTPClient := apiKey, baseURL, http.DefaultClient
	t.Cleanup(func() {
		apiKey, baseURL, http.DefaultClient = originalKey, originalBaseURL, originalHTTPClient
	})
	apiKey = "test-key"
	baseURL = "https://api.example.test"
	http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("request must not be sent")
		return nil, nil
	})}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"clients", "create", "--data", `{"name":"Acme"}`})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "without --confirm") {
		t.Fatalf("Execute() error = %v, want missing confirmation error", err)
	}
}

func TestWriteDryRunDoesNotRequireAPIKeyOrSendRequest(t *testing.T) {
	originalKey, originalBaseURL, originalHTTPClient := apiKey, baseURL, http.DefaultClient
	t.Cleanup(func() {
		apiKey, baseURL, http.DefaultClient = originalKey, originalBaseURL, originalHTTPClient
	})
	apiKey = ""
	baseURL = "https://api.example.test"
	http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("request must not be sent")
		return nil, nil
	})}

	var output bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"clients", "update", "42", "--data", `{"name":"Acme"}`, "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); !strings.Contains(got, `"dry_run": true`) || !strings.Contains(got, `"method": "PATCH"`) || !strings.Contains(got, `"path": "/manage/v1/clients/42"`) {
		t.Fatalf("dry-run output = %q", got)
	}
}

func TestWriteSendsConfirmedRequest(t *testing.T) {
	originalKey, originalBaseURL, originalHTTPClient := apiKey, baseURL, http.DefaultClient
	t.Cleanup(func() {
		apiKey, baseURL, http.DefaultClient = originalKey, originalBaseURL, originalHTTPClient
	})
	http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := req.URL.Path, "/manage/v1/clients"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Status:     "201 Created",
			Body:       io.NopCloser(strings.NewReader(`{"message":"Created"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	var output bytes.Buffer
	cmd := NewRootCmd()
	apiKey = "test-key"
	baseURL = "https://api.example.test"
	cmd.SetOut(&output)
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs([]string{"clients", "create", "--data", `{"name":"Acme"}`, "--confirm"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"message": "Created"`) {
		t.Fatalf("output = %q", output.String())
	}
}

func TestWriteCommandsRegistered(t *testing.T) {
	root := NewRootCmd()
	for _, path := range [][]string{
		{"locations", "create"},
		{"locations", "update"},
		{"clients", "create"},
		{"clients", "update"},
	} {
		if _, _, err := root.Find(path); err != nil {
			t.Errorf("command %q is not registered: %v", strings.Join(path, " "), err)
		}
	}
}

func TestValidateJSONObject(t *testing.T) {
	t.Parallel()

	for _, data := range []string{"[]", `"string"`, "null", "{"} {
		if _, err := validateJSONObject(data); err == nil {
			t.Errorf("validateJSONObject(%q) error = nil", data)
		}
	}
}

func TestValidateJSONObjectReadsFile(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/payload.json"
	if err := os.WriteFile(path, []byte(`{"name":"Acme"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	body, err := validateJSONObject("@" + path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(body), `{"name":"Acme"}`; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if _, err := validateJSONObject("@" + path + ".missing"); err == nil {
		t.Fatal("validateJSONObject() error = nil for missing file")
	}
}

func TestLocationUpdateSendsConfirmedRequest(t *testing.T) {
	originalKey, originalBaseURL, originalHTTPClient := apiKey, baseURL, http.DefaultClient
	t.Cleanup(func() {
		apiKey, baseURL, http.DefaultClient = originalKey, originalBaseURL, originalHTTPClient
	})
	http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.Method, http.MethodPatch; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := req.URL.EscapedPath(), "/manage/v1/locations/a%2Fb"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"message":"Updated"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	cmd := NewRootCmd()
	apiKey = "test-key"
	baseURL = "https://api.example.test"
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs([]string{"locations", "update", "a/b", "--data", `{"business_name":"Acme"}`, "--confirm"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
