package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/edmondop/terraform-provider-buttondown/internal/client"
)

var _ provider.Provider = &ButtondownProvider{}

type ButtondownProvider struct {
	version string
}

type ButtondownProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ButtondownProvider{
			version: version,
		}
	}
}

func (p *ButtondownProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "buttondown"
	resp.Version = p.version
}

func (p *ButtondownProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Buttondown newsletter resources.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Buttondown API key. Can also be set via BUTTONDOWN_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Buttondown API base URL. Defaults to https://api.buttondown.com.",
				Optional:    true,
			},
		},
	}
}

func (p *ButtondownProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ButtondownProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := resolveEnvDefault(config.APIKey, "BUTTONDOWN_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The Buttondown API key must be set in the provider configuration or via the BUTTONDOWN_API_KEY environment variable.",
		)
		return
	}

	baseURL := resolveEnvDefault(config.BaseURL, "BUTTONDOWN_BASE_URL")

	var opts []client.Option
	if baseURL != "" {
		opts = append(opts, client.WithBaseURL(baseURL))
	}

	c := client.New(apiKey, opts...)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *ButtondownProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTagResource,
		NewEmailResource,
		NewAutomationResource,
		NewWebhookResource,
		NewSnippetResource,
		NewFormResource,
		NewExternalFeedResource,
		NewSurveyResource,
		NewBookResource,
		NewUserResource,
		NewNewsletterResource,
	}
}

func (p *ButtondownProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountDataSource,
	}
}
