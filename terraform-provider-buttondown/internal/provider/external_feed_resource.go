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
	_ resource.Resource                = &ExternalFeedResource{}
	_ resource.ResourceWithImportState = &ExternalFeedResource{}
)

type ExternalFeedResource struct{ client *client.Client }

type ExternalFeedResourceModel struct {
	ID       types.String `tfsdk:"id"`
	URL      types.String `tfsdk:"url"`
	Behavior types.String `tfsdk:"behavior"`
	Cadence  types.String `tfsdk:"cadence"`
	Subject  types.String `tfsdk:"subject"`
	Body     types.String `tfsdk:"body"`
	Label    types.String `tfsdk:"label"`
	Status   types.String `tfsdk:"status"`
}

func NewExternalFeedResource() resource.Resource { return &ExternalFeedResource{} }

func (r *ExternalFeedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_feed"
}

func (r *ExternalFeedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown external feed (RSS/Atom import).",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"url":      schema.StringAttribute{Required: true, Description: "Feed URL (max 2000 chars)."},
			"behavior": schema.StringAttribute{Required: true, Description: "What to do with feed items: draft or emails."},
			"cadence":  schema.StringAttribute{Required: true, Description: "Check frequency: every, daily, weekly, monthly."},
			"subject":  schema.StringAttribute{Required: true, Description: "Email subject template (max 255 chars)."},
			"body":     schema.StringAttribute{Required: true, Description: "Email body template."},
			"label":    schema.StringAttribute{Optional: true, Computed: true, Description: "Feed label (max 255 chars)."},
			"status":   schema.StringAttribute{Optional: true, Computed: true, Description: "Feed status."},
		},
	}
}

func (r *ExternalFeedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ExternalFeedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ExternalFeedResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	input := client.ExternalFeedInput{
		URL: plan.URL.ValueString(), Behavior: plan.Behavior.ValueString(),
		Cadence: plan.Cadence.ValueString(), Subject: plan.Subject.ValueString(),
		Body: plan.Body.ValueString(),
		CadenceMetadata: map[string]string{},
		Filters: &client.FilterGroup{Filters: []client.Filter{}, Groups: []client.FilterGroup{}, Predicate: "and"},
	}
	if !plan.Label.IsNull() { input.Label = plan.Label.ValueString() }
	var f client.ExternalFeed
	if err := r.client.Post(ctx, "/v1/external_feeds", input, &f); err != nil {
		resp.Diagnostics.AddError("Error creating external feed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, externalFeedToModel(&f))...)
}

func (r *ExternalFeedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ExternalFeedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	var f client.ExternalFeed
	if err := r.client.Get(ctx, "/v1/external_feeds/"+state.ID.ValueString(), &f); err != nil {
		if client.IsNotFound(err) { resp.State.RemoveResource(ctx); return }
		resp.Diagnostics.AddError("Error reading external feed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, externalFeedToModel(&f))...)
}

func (r *ExternalFeedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ExternalFeedResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	input := client.ExternalFeedUpdateInput{}
	if !plan.Behavior.Equal(state.Behavior) { v := plan.Behavior.ValueString(); input.Behavior = &v }
	if !plan.Cadence.Equal(state.Cadence) { v := plan.Cadence.ValueString(); input.Cadence = &v }
	if !plan.Subject.Equal(state.Subject) { v := plan.Subject.ValueString(); input.Subject = &v }
	if !plan.Body.Equal(state.Body) { v := plan.Body.ValueString(); input.Body = &v }
	if !plan.Label.Equal(state.Label) { v := plan.Label.ValueString(); input.Label = &v }
	if !plan.Status.Equal(state.Status) { v := plan.Status.ValueString(); input.Status = &v }
	var f client.ExternalFeed
	if err := r.client.Patch(ctx, "/v1/external_feeds/"+state.ID.ValueString(), input, &f); err != nil {
		resp.Diagnostics.AddError("Error updating external feed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, externalFeedToModel(&f))...)
}

func (r *ExternalFeedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ExternalFeedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	if err := r.client.Delete(ctx, "/v1/external_feeds/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting external feed", err.Error())
	}
}

func (r *ExternalFeedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func externalFeedToModel(f *client.ExternalFeed) *ExternalFeedResourceModel {
	return &ExternalFeedResourceModel{
		ID: types.StringValue(f.ID), URL: types.StringValue(f.URL), Behavior: types.StringValue(f.Behavior),
		Cadence: types.StringValue(f.Cadence), Subject: types.StringValue(f.Subject), Body: types.StringValue(f.Body),
		Label: types.StringValue(f.Label), Status: types.StringValue(f.Status),
	}
}
