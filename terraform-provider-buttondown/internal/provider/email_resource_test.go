package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/edmondop/terraform-provider-buttondown/internal/client"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestEmailResource_creates_as_draft(t *testing.T) {
	email := emailFixture("email-123", "Weekly Update", "draft")
	server := newMockServer(t, []mockRoute{
		{http.MethodGet, "/v1/emails/email-123", http.StatusOK, email},
	})

	// Override POST handler to verify status is always "draft"
	originalHandler := server.Config.Handler
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/emails" {
			var input client.EmailInput
			json.NewDecoder(r.Body).Decode(&input)
			if input.Status != "draft" {
				t.Errorf("expected status 'draft', got %q", input.Status)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(email)
			return
		}
		originalHandler.ServeHTTP(w, r)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
resource "buttondown_email" "test" {
  subject = "Weekly Update"
  body    = "Hello world"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("buttondown_email.test", "id", "email-123"),
					resource.TestCheckResourceAttr("buttondown_email.test", "status", "draft"),
					resource.TestCheckResourceAttr("buttondown_email.test", "subject", "Weekly Update"),
				),
			},
		},
	})
}

func TestEmailResource_status_is_computed_only(t *testing.T) {
	// Verify the schema marks status as Computed-only (no Optional)
	r := &EmailResource{}
	schemaResp := fwresource.SchemaResponse{}
	r.Schema(nil, fwresource.SchemaRequest{}, &schemaResp)

	statusAttr := schemaResp.Schema.Attributes["status"]
	if statusAttr == nil {
		t.Fatal("expected 'status' attribute in schema")
	}
	if statusAttr.IsOptional() {
		t.Error("status should NOT be optional — users must not set email status")
	}
	if !statusAttr.IsComputed() {
		t.Error("status should be computed")
	}
}
