package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := New("test-api-key", WithBaseURL(server.URL))
	return c, server
}

func TestClient_Get_returns_resource(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Token test-api-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Account{Username: "test", EmailAddress: "test@example.com"})
	})

	var account Account
	err := c.Get(context.Background(), "/accounts/me", &account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account.Username != "test" {
		t.Errorf("expected username 'test', got %q", account.Username)
	}
}

func TestClient_Post_sends_body_and_returns_result(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected JSON content type, got %q", r.Header.Get("Content-Type"))
		}
		var input TagInput
		_ = json.NewDecoder(r.Body).Decode(&input)
		if input.Name != "test-tag" {
			t.Errorf("expected name 'test-tag', got %q", input.Name)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Tag{ID: "tag-123", Name: input.Name, Color: input.Color})
	})

	var tag Tag
	err := c.Post(context.Background(), "/tags", TagInput{Name: "test-tag", Color: "#ff0000"}, &tag)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.ID != "tag-123" {
		t.Errorf("expected ID 'tag-123', got %q", tag.ID)
	}
}

func TestClient_Patch_sends_partial_update(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Tag{ID: "tag-123", Name: "updated", Color: "#00ff00"})
	})

	var tag Tag
	name := "updated"
	err := c.Patch(context.Background(), "/tags/tag-123", TagUpdateInput{Name: &name}, &tag)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", tag.Name)
	}
}

func TestClient_Delete_succeeds_on_204(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Delete(context.Background(), "/tags/tag-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_returns_APIError_on_404(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"detail": "Not found."})
	})

	var tag Tag
	err := c.Get(context.Background(), "/tags/nonexistent", &tag)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false for error: %v", err)
	}
	apiErr := err.(*APIError)
	if apiErr.Detail != "Not found." {
		t.Errorf("expected detail 'Not found.', got %q", apiErr.Detail)
	}
}

func TestClient_List_returns_paginated_results(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PageResponse[Tag]{
			Count: 2,
			Results: []Tag{
				{ID: "1", Name: "tag1"},
				{ID: "2", Name: "tag2"},
			},
		})
	})

	var page PageResponse[Tag]
	err := c.List(context.Background(), "/tags", &page)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 2 {
		t.Errorf("expected count 2, got %d", page.Count)
	}
	if len(page.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(page.Results))
	}
}
