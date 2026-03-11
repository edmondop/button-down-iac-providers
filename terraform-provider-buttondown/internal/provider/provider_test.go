package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edmondop/terraform-provider-buttondown/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func testAccProtoV6ProviderFactories(serverURL string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"buttondown": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func providerConfig(serverURL string) string {
	return `
provider "buttondown" {
  api_key  = "test-api-key"
  base_url = "` + serverURL + `"
}
`
}

type mockRoute struct {
	method  string
	path    string
	status  int
	body    any
}

func newMockServer(t *testing.T, routes []mockRoute) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range routes {
			if r.Method == route.method && r.URL.Path == route.path {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(route.status)
				if route.body != nil {
					json.NewEncoder(w).Encode(route.body)
				}
				return
			}
		}
		// If we receive DELETE and no route matched, return 204 (common for cleanup)
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Logf("unhandled request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Not found."})
	}))
	t.Cleanup(server.Close)
	return server
}

func tagFixture(id, name, color string) client.Tag {
	return client.Tag{
		ID:    id,
		Name:  name,
		Color: color,
	}
}

func emailFixture(id, subject, status string) client.Email {
	return client.Email{
		ID:      id,
		Subject: subject,
		Status:  status,
	}
}

func accountFixture(username, email string) client.Account {
	return client.Account{
		Username:     username,
		EmailAddress: email,
	}
}

func mockHandler(t *testing.T, w http.ResponseWriter, r *http.Request, routes []mockRoute) {
	t.Helper()
	for _, route := range routes {
		if r.Method == route.method && r.URL.Path == route.path {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(route.status)
			if route.body != nil {
				json.NewEncoder(w).Encode(route.body)
			}
			return
		}
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	t.Logf("unhandled request: %s %s", r.Method, r.URL.Path)
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"detail": "Not found."})
}
