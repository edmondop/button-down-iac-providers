package provider

import (
	"context"
	"fmt"

	"github.com/edmondop/terraform-provider-buttondown/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AccountDataSource{}

type AccountDataSource struct {
	client *client.Client
}

type AccountDataSourceModel struct {
	Username     types.String `tfsdk:"username"`
	EmailAddress types.String `tfsdk:"email_address"`
}

func NewAccountDataSource() datasource.DataSource {
	return &AccountDataSource{}
}

func (d *AccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *AccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve the Buttondown account associated with the current API key.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "The account username.",
			},
			"email_address": schema.StringAttribute{
				Computed:    true,
				Description: "The account email address.",
			},
		},
	}
}

func (d *AccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *AccountDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var account client.Account
	if err := d.client.Get(ctx, "/v1/accounts/me", &account); err != nil {
		resp.Diagnostics.AddError("Error reading account", err.Error())
		return
	}

	state := AccountDataSourceModel{
		Username:     types.StringValue(account.Username),
		EmailAddress: types.StringValue(account.EmailAddress),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
