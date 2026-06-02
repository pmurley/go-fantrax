package auth_client

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func makeResp(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestReadBody_PageError_ReturnsError(t *testing.T) {
	resp := makeResp(`{"pageError":{"code":"STALE_CLIENT","title":"App Update Required","text":"Your browser is using an outdated cached version."}}`)
	_, err := readBody(resp)
	if err == nil {
		t.Fatal("expected error for STALE_CLIENT response, got nil")
	}
	if !strings.Contains(err.Error(), "STALE_CLIENT") {
		t.Errorf("error should mention STALE_CLIENT, got: %v", err)
	}
}

func TestReadBody_ValidResponse_ReturnsBody(t *testing.T) {
	payload := `{"data":{},"roles":[],"responses":[{"data":{"userInfo":{}}}]}`
	resp := makeResp(payload)
	body, err := readBody(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != payload {
		t.Errorf("body mismatch: got %q, want %q", body, payload)
	}
}

func TestReadBody_EmptyPageError_DoesNotError(t *testing.T) {
	// pageError present but code is empty — not a real error, don't reject it.
	resp := makeResp(`{"pageError":{"code":"","text":""},"responses":[]}`)
	_, err := readBody(resp)
	if err != nil {
		t.Errorf("empty pageError code should not error, got: %v", err)
	}
}

func TestReadBody_NoPageError_DoesNotError(t *testing.T) {
	resp := makeResp(`{"responses":[]}`)
	_, err := readBody(resp)
	if err != nil {
		t.Errorf("response without pageError should not error, got: %v", err)
	}
}

func TestReadBody_UnparsableJSON_ReturnsBodyNotError(t *testing.T) {
	// Some /fxa/ endpoints return non-JSON; readBody should not error on those.
	resp := makeResp(`not json at all`)
	body, err := readBody(resp)
	if err != nil {
		t.Errorf("unparseable body should pass through, got err: %v", err)
	}
	if string(body) != "not json at all" {
		t.Errorf("body should be returned verbatim, got %q", body)
	}
}
