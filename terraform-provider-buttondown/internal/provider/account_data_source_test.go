package provider

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccountDataSource_reads_account(t *testing.T) {
	account := accountFixture("edmondo", "edmondo@example.com")
	server := newMockServer(t, []mockRoute{
		{http.MethodGet, "/v1/accounts/me", http.StatusOK, account},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(server.URL),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(server.URL) + `
data "buttondown_account" "me" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.buttondown_account.me", "username", "edmondo"),
					resource.TestCheckResourceAttr("data.buttondown_account.me", "email_address", "edmondo@example.com"),
				),
			},
		},
	})
}
