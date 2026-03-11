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
	_ resource.Resource                = &FormResource{}
	_ resource.ResourceWithImportState = &FormResource{}
)

type FormResource struct{ client *client.Client }

type FormResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Slug        types.String `tfsdk:"slug"`
	Body        types.String `tfsdk:"body"`
	CSS         types.String `tfsdk:"css"`
	SuccessBody types.String `tfsdk:"success_body"`
	Status      types.String `tfsdk:"status"`
}

func NewFormResource() resource.Resource { return &FormResource{} }

func (r *FormResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_form"
}

func (r *FormResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown subscription form.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"title":        schema.StringAttribute{Required: true, Description: "Form title (max 255 chars)."},
			"slug":         schema.StringAttribute{Required: true, Description: "URL slug (max 100 chars)."},
			"body":         schema.StringAttribute{Optional: true, Computed: true, Description: "Form body content."},
			"css":          schema.StringAttribute{Optional: true, Computed: true, Description: "Custom CSS for the form."},
			"success_body": schema.StringAttribute{Optional: true, Computed: true, Description: "Message shown after successful subscription."},
			"status":       schema.StringAttribute{Optional: true, Computed: true, Description: "Form status: active or inactive."},
		},
	}
}

func (r *FormResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FormResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FormResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.FormInput{Title: plan.Title.ValueString(), Slug: plan.Slug.ValueString()}
	if !plan.Body.IsNull() { input.Body = plan.Body.ValueString() }
	if !plan.CSS.IsNull() { input.CSS = plan.CSS.ValueString() }
	if !plan.SuccessBody.IsNull() { input.SuccessBody = plan.SuccessBody.ValueString() }
	if !plan.Status.IsNull() { input.Status = plan.Status.ValueString() }
	var f client.Form
	if err := r.client.Post(ctx, "/v1/forms", input, &f); err != nil {
		resp.Diagnostics.AddError("Error creating form", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, formToModel(&f))...)
}

func (r *FormResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FormResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var f client.Form
	if err := r.client.Get(ctx, "/v1/forms/"+state.ID.ValueString(), &f); err != nil {
		if client.IsNotFound(err) { resp.State.RemoveResource(ctx); return }
		resp.Diagnostics.AddError("Error reading form", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, formToModel(&f))...)
}

func (r *FormResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FormResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.FormUpdateInput{}
	if !plan.Title.Equal(state.Title) { v := plan.Title.ValueString(); input.Title = &v }
	if !plan.Slug.Equal(state.Slug) { v := plan.Slug.ValueString(); input.Slug = &v }
	if !plan.Body.Equal(state.Body) { v := plan.Body.ValueString(); input.Body = &v }
	if !plan.CSS.Equal(state.CSS) { v := plan.CSS.ValueString(); input.CSS = &v }
	if !plan.SuccessBody.Equal(state.SuccessBody) { v := plan.SuccessBody.ValueString(); input.SuccessBody = &v }
	if !plan.Status.Equal(state.Status) { v := plan.Status.ValueString(); input.Status = &v }
	var f client.Form
	if err := r.client.Patch(ctx, "/v1/forms/"+state.ID.ValueString(), input, &f); err != nil {
		resp.Diagnostics.AddError("Error updating form", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, formToModel(&f))...)
}

func (r *FormResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FormResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	if err := r.client.Delete(ctx, "/v1/forms/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting form", err.Error())
	}
}

func (r *FormResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func formToModel(f *client.Form) *FormResourceModel {
	return &FormResourceModel{
		ID: types.StringValue(f.ID), Title: types.StringValue(f.Title), Slug: types.StringValue(f.Slug),
		Body: types.StringValue(f.Body), CSS: types.StringValue(f.CSS), SuccessBody: types.StringValue(f.SuccessBody),
		Status: types.StringValue(f.Status),
	}
}
