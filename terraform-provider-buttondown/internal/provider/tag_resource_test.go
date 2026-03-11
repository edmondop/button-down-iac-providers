package provider

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTagResource_create_and_read(t *testing.T) {
	tag := tagFixture("tag-abc-123", "Engineering", "#0066cc")
	server := newMockServer(t, []mockRoute{
		{http.MethodPost, "/v1/tags", http.StatusCreated, tag},
		{http.MethodGet, "/v1/tags/tag-abc-123", http.StatusOK, tag},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
resource "buttondown_tag" "test" {
  name  = "Engineering"
  color = "#0066cc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("buttondown_tag.test", "id", "tag-abc-123"),
					resource.TestCheckResourceAttr("buttondown_tag.test", "name", "Engineering"),
					resource.TestCheckResourceAttr("buttondown_tag.test", "color", "#0066cc"),
				),
			},
		},
	})
}

func TestTagResource_update_changes_name(t *testing.T) {
	tagV1 := tagFixture("tag-abc-123", "Engineering", "#0066cc")
	tagV2 := tagFixture("tag-abc-123", "DevOps", "#0066cc")

	server := newMockServer(t, []mockRoute{
		{http.MethodPost, "/v1/tags", http.StatusCreated, tagV1},
		{http.MethodGet, "/v1/tags/tag-abc-123", http.StatusOK, tagV1},
		{http.MethodPatch, "/v1/tags/tag-abc-123", http.StatusOK, tagV2},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
resource "buttondown_tag" "test" {
  name  = "Engineering"
  color = "#0066cc"
}
`,
				Check: resource.TestCheckResourceAttr("buttondown_tag.test", "name", "Engineering"),
			},
			{
				PreConfig: func() {
					// Update mock to return v2 for subsequent reads
					server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mockHandler(t, w, r, []mockRoute{
							{http.MethodPatch, "/v1/tags/tag-abc-123", http.StatusOK, tagV2},
							{http.MethodGet, "/v1/tags/tag-abc-123", http.StatusOK, tagV2},
						})
					})
				},
				Config: providerConfig(server.URL) + `
resource "buttondown_tag" "test" {
  name  = "DevOps"
  color = "#0066cc"
}
`,
				Check: resource.TestCheckResourceAttr("buttondown_tag.test", "name", "DevOps"),
			},
		},
	})
}

func TestTagResource_import(t *testing.T) {
	tag := tagFixture("tag-abc-123", "Engineering", "#0066cc")
	server := newMockServer(t, []mockRoute{
		{http.MethodPost, "/v1/tags", http.StatusCreated, tag},
		{http.MethodGet, "/v1/tags/tag-abc-123", http.StatusOK, tag},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
resource "buttondown_tag" "test" {
  name  = "Engineering"
  color = "#0066cc"
}
`,
			},
			{
				ResourceName:      "buttondown_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
