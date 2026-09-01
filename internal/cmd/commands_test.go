package cmd

import (
	"bytes"
	"net/url"
	"strings"
	"testing"
)

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
