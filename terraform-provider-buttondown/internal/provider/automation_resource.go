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
	_ resource.Resource                = &AutomationResource{}
	_ resource.ResourceWithImportState = &AutomationResource{}
)

type AutomationResource struct{ client *client.Client }

type AutomationResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Status  types.String `tfsdk:"status"`
	Trigger types.String `tfsdk:"trigger"`
}

func NewAutomationResource() resource.Resource { return &AutomationResource{} }

func (r *AutomationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation"
}

func (r *AutomationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown automation.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":    schema.StringAttribute{Required: true, Description: "Automation name (max 100 chars)."},
			"status":  schema.StringAttribute{Optional: true, Computed: true, Description: "Status: active or inactive."},
			"trigger": schema.StringAttribute{Required: true, Description: "Event type that triggers this automation."},
		},
	}
}

func (r *AutomationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AutomationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AutomationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.AutomationInput{
		Name:    plan.Name.ValueString(),
		Trigger: plan.Trigger.ValueString(),
		Actions: []client.Action{},
		Filters: &client.FilterGroup{Filters: []client.Filter{}, Groups: []client.FilterGroup{}, Predicate: "and"},
	}
	var a client.Automation
	if err := r.client.Post(ctx, "/v1/automations", input, &a); err != nil {
		resp.Diagnostics.AddError("Error creating automation", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, automationToModel(&a))...)
}

func (r *AutomationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AutomationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var a client.Automation
	if err := r.client.Get(ctx, "/v1/automations/"+state.ID.ValueString(), &a); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading automation", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, automationToModel(&a))...)
}

func (r *AutomationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AutomationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.AutomationUpdateInput{}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString(); input.Name = &v
	}
	if !plan.Status.Equal(state.Status) {
		v := plan.Status.ValueString(); input.Status = &v
	}
	if !plan.Trigger.Equal(state.Trigger) {
		v := plan.Trigger.ValueString(); input.Trigger = &v
	}
	var a client.Automation
	if err := r.client.Patch(ctx, "/v1/automations/"+state.ID.ValueString(), input, &a); err != nil {
		resp.Diagnostics.AddError("Error updating automation", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, automationToModel(&a))...)
}

func (r *AutomationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AutomationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete(ctx, "/v1/automations/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting automation", err.Error())
	}
}

func (r *AutomationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func automationToModel(a *client.Automation) *AutomationResourceModel {
	return &AutomationResourceModel{
		ID:      types.StringValue(a.ID),
		Name:    types.StringValue(a.Name),
		Status:  types.StringValue(a.Status),
		Trigger: types.StringValue(a.Trigger),
	}
}
