package provider

import (
	"context"
	"fmt"

	"github.com/edmondop/terraform-provider-buttondown/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NewsletterResource{}
	_ resource.ResourceWithImportState = &NewsletterResource{}
)

type NewsletterResource struct{ client *client.Client }

type NewsletterResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Username     types.String `tfsdk:"username"`
	Description  types.String `tfsdk:"description"`
	Domain       types.String `tfsdk:"domain"`
	CSS          types.String `tfsdk:"css"`
	Footer       types.String `tfsdk:"footer"`
	Header       types.String `tfsdk:"header"`
	FromName     types.String `tfsdk:"from_name"`
	Locale       types.String `tfsdk:"locale"`
	Template     types.String `tfsdk:"template"`
	ArchiveTheme types.String `tfsdk:"archive_theme"`
	TintColor    types.String `tfsdk:"tint_color"`
	Timezone     types.String `tfsdk:"timezone"`
}

func NewNewsletterResource() resource.Resource { return &NewsletterResource{} }

func (r *NewsletterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_newsletter"
}

func (r *NewsletterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown newsletter.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":          schema.StringAttribute{Required: true, Description: "Newsletter name."},
			"username":      schema.StringAttribute{Required: true, Description: "Newsletter username (URL slug)."},
			"description":   schema.StringAttribute{Required: true, Description: "Newsletter description."},
			"domain":        schema.StringAttribute{Optional: true, Computed: true, Description: "Custom domain."},
			"css":           schema.StringAttribute{Optional: true, Computed: true, Description: "Custom CSS for emails."},
			"footer":        schema.StringAttribute{Optional: true, Computed: true, Description: "Email footer content."},
			"header":        schema.StringAttribute{Optional: true, Computed: true, Description: "Email header content."},
			"from_name":     schema.StringAttribute{Optional: true, Computed: true, Description: "From name for sent emails."},
			"locale":        schema.StringAttribute{Optional: true, Computed: true, Description: "Newsletter locale (e.g. en, fr, de)."},
			"template":      schema.StringAttribute{Optional: true, Computed: true, Description: "Email template: classic, modern, naked, plaintext."},
			"archive_theme": schema.StringAttribute{Optional: true, Computed: true, Description: "Archive theme: classic, modern, arbus, lovelace, myrna."},
			"tint_color":    schema.StringAttribute{Optional: true, Computed: true, Description: "Brand color as hex (max 7 chars)."},
			"timezone":      schema.StringAttribute{Optional: true, Computed: true, Description: "Timezone (e.g. Etc/UTC, America/New_York)."},
		},
	}
}

func (r *NewsletterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *NewsletterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NewsletterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	input := client.NewsletterInput{
		Name: plan.Name.ValueString(), Username: plan.Username.ValueString(),
		Description: plan.Description.ValueString(),
	}
	var n client.Newsletter
	if err := r.client.Post(ctx, "/v1/newsletters", input, &n); err != nil {
		resp.Diagnostics.AddError("Error creating newsletter", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newsletterToModel(&n))...)
}

func (r *NewsletterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NewsletterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }

	// Newsletters have no GET-by-ID endpoint, so we list and filter
	n, err := r.findNewsletterByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading newsletter", err.Error())
		return
	}
	if n == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newsletterToModel(n))...)
}

func (r *NewsletterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NewsletterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	input := client.NewsletterUpdateInput{}
	if !plan.Name.Equal(state.Name) { v := plan.Name.ValueString(); input.Name = &v }
	if !plan.Description.Equal(state.Description) { v := plan.Description.ValueString(); input.Description = &v }
	if !plan.Domain.Equal(state.Domain) { v := plan.Domain.ValueString(); input.Domain = &v }
	if !plan.CSS.Equal(state.CSS) { v := plan.CSS.ValueString(); input.CSS = &v }
	if !plan.Footer.Equal(state.Footer) { v := plan.Footer.ValueString(); input.Footer = &v }
	if !plan.Header.Equal(state.Header) { v := plan.Header.ValueString(); input.Header = &v }
	if !plan.FromName.Equal(state.FromName) { v := plan.FromName.ValueString(); input.FromName = &v }
	if !plan.Locale.Equal(state.Locale) { v := plan.Locale.ValueString(); input.Locale = &v }
	if !plan.Template.Equal(state.Template) { v := plan.Template.ValueString(); input.Template = &v }
	if !plan.ArchiveTheme.Equal(state.ArchiveTheme) { v := plan.ArchiveTheme.ValueString(); input.ArchiveTheme = &v }
	if !plan.TintColor.Equal(state.TintColor) { v := plan.TintColor.ValueString(); input.TintColor = &v }
	if !plan.Timezone.Equal(state.Timezone) { v := plan.Timezone.ValueString(); input.Timezone = &v }
	var n client.Newsletter
	if err := r.client.Patch(ctx, "/v1/newsletters/"+state.ID.ValueString(), input, &n); err != nil {
		resp.Diagnostics.AddError("Error updating newsletter", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newsletterToModel(&n))...)
}

func (r *NewsletterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NewsletterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	if err := r.client.Delete(ctx, "/v1/newsletters/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting newsletter", err.Error())
	}
}

func (r *NewsletterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NewsletterResource) findNewsletterByID(ctx context.Context, id string) (*client.Newsletter, error) {
	var page client.PageResponse[client.Newsletter]
	if err := r.client.List(ctx, "/v1/newsletters", &page); err != nil {
		return nil, err
	}
	for i := range page.Results {
		if page.Results[i].ID == id {
			return &page.Results[i], nil
		}
	}
	return nil, nil
}

func newsletterToModel(n *client.Newsletter) *NewsletterResourceModel {
	return &NewsletterResourceModel{
		ID: types.StringValue(n.ID), Name: types.StringValue(n.Name),
		Username: types.StringValue(n.Username), Description: types.StringValue(n.Description),
		Domain: types.StringValue(n.Domain), CSS: types.StringValue(n.CSS),
		Footer: types.StringValue(n.Footer), Header: types.StringValue(n.Header),
		FromName: types.StringValue(n.FromName), Locale: types.StringValue(n.Locale),
		Template: types.StringValue(n.Template), ArchiveTheme: types.StringValue(n.ArchiveTheme),
		TintColor: types.StringValue(n.TintColor), Timezone: types.StringValue(n.Timezone),
	}
}
