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
	_ resource.Resource                = &TagResource{}
	_ resource.ResourceWithImportState = &TagResource{}
)

type TagResource struct {
	client *client.Client
}

type TagResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Color              types.String `tfsdk:"color"`
	Description        types.String `tfsdk:"description"`
	PublicDescription  types.String `tfsdk:"public_description"`
	SubscriberEditable types.Bool   `tfsdk:"subscriber_editable"`
}

func NewTagResource() resource.Resource {
	return &TagResource{}
}

func (r *TagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *TagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown tag.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The tag ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The tag name (max 100 characters).",
			},
			"color": schema.StringAttribute{
				Required:    true,
				Description: "The tag color as a hex string (max 10 characters).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The tag description.",
			},
			"public_description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The public-facing tag description.",
			},
			"subscriber_editable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether subscribers can self-manage this tag.",
			},
		},
	}
}

func (r *TagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.TagInput{
		Name:  plan.Name.ValueString(),
		Color: plan.Color.ValueString(),
	}
	if !plan.Description.IsNull() {
		input.Description = plan.Description.ValueString()
	}
	if !plan.PublicDescription.IsNull() {
		input.PublicDescription = plan.PublicDescription.ValueString()
	}
	if !plan.SubscriberEditable.IsNull() {
		v := plan.SubscriberEditable.ValueBool()
		input.SubscriberEditable = &v
	}

	var tag client.Tag
	if err := r.client.Post(ctx, "/v1/tags", input, &tag); err != nil {
		resp.Diagnostics.AddError("Error creating tag", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tagToModel(&tag))...)
}

func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tag client.Tag
	err := r.client.Get(ctx, "/v1/tags/"+state.ID.ValueString(), &tag)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading tag", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tagToModel(&tag))...)
}

func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TagResourceModel
	var state TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.TagUpdateInput{}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		input.Name = &v
	}
	if !plan.Color.Equal(state.Color) {
		v := plan.Color.ValueString()
		input.Color = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		input.Description = &v
	}
	if !plan.PublicDescription.Equal(state.PublicDescription) {
		v := plan.PublicDescription.ValueString()
		input.PublicDescription = &v
	}
	if !plan.SubscriberEditable.Equal(state.SubscriberEditable) {
		v := plan.SubscriberEditable.ValueBool()
		input.SubscriberEditable = &v
	}

	var tag client.Tag
	if err := r.client.Patch(ctx, "/v1/tags/"+state.ID.ValueString(), input, &tag); err != nil {
		resp.Diagnostics.AddError("Error updating tag", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tagToModel(&tag))...)
}

func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, "/v1/tags/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting tag", err.Error())
		return
	}
}

func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func tagToModel(tag *client.Tag) *TagResourceModel {
	return &TagResourceModel{
		ID:                 types.StringValue(tag.ID),
		Name:               types.StringValue(tag.Name),
		Color:              types.StringValue(tag.Color),
		Description:        types.StringValue(tag.Description),
		PublicDescription:  types.StringValue(tag.PublicDescription),
		SubscriberEditable: types.BoolValue(tag.SubscriberEditable),
	}
}
