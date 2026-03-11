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
	_ resource.Resource                = &SnippetResource{}
	_ resource.ResourceWithImportState = &SnippetResource{}
)

type SnippetResource struct{ client *client.Client }

type SnippetResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Content    types.String `tfsdk:"content"`
	Mode       types.String `tfsdk:"mode"`
}

func NewSnippetResource() resource.Resource { return &SnippetResource{} }

func (r *SnippetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snippet"
}

func (r *SnippetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown snippet (reusable content block).",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"identifier": schema.StringAttribute{Required: true, Description: "Unique identifier for the snippet (max 100 chars)."},
			"name":       schema.StringAttribute{Required: true, Description: "Display name (max 255 chars)."},
			"content":    schema.StringAttribute{Optional: true, Computed: true, Description: "Snippet content."},
			"mode":       schema.StringAttribute{Optional: true, Computed: true, Description: "Rendering mode: fancy, naked, or plaintext."},
		},
	}
}

func (r *SnippetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *SnippetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SnippetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.SnippetInput{
		Identifier: plan.Identifier.ValueString(),
		Name:       plan.Name.ValueString(),
	}
	if !plan.Content.IsNull() {
		input.Content = plan.Content.ValueString()
	}
	if !plan.Mode.IsNull() {
		input.Mode = plan.Mode.ValueString()
	}
	var s client.Snippet
	if err := r.client.Post(ctx, "/v1/snippets", input, &s); err != nil {
		resp.Diagnostics.AddError("Error creating snippet", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, snippetToModel(&s))...)
}

func (r *SnippetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SnippetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var s client.Snippet
	if err := r.client.Get(ctx, "/v1/snippets/"+state.ID.ValueString(), &s); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading snippet", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, snippetToModel(&s))...)
}

func (r *SnippetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SnippetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.SnippetUpdateInput{}
	if !plan.Identifier.Equal(state.Identifier) {
		v := plan.Identifier.ValueString(); input.Identifier = &v
	}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString(); input.Name = &v
	}
	if !plan.Content.Equal(state.Content) {
		v := plan.Content.ValueString(); input.Content = &v
	}
	if !plan.Mode.Equal(state.Mode) {
		v := plan.Mode.ValueString(); input.Mode = &v
	}
	var s client.Snippet
	if err := r.client.Patch(ctx, "/v1/snippets/"+state.ID.ValueString(), input, &s); err != nil {
		resp.Diagnostics.AddError("Error updating snippet", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, snippetToModel(&s))...)
}

func (r *SnippetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SnippetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete(ctx, "/v1/snippets/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting snippet", err.Error())
	}
}

func (r *SnippetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func snippetToModel(s *client.Snippet) *SnippetResourceModel {
	return &SnippetResourceModel{
		ID:         types.StringValue(s.ID),
		Identifier: types.StringValue(s.Identifier),
		Name:       types.StringValue(s.Name),
		Content:    types.StringValue(s.Content),
		Mode:       types.StringValue(s.Mode),
	}
}
