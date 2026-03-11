package provider

import (
	"context"
	"fmt"

	"github.com/edmondop/terraform-provider-buttondown/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &WebhookResource{}
	_ resource.ResourceWithImportState = &WebhookResource{}
)

type WebhookResource struct {
	client *client.Client
}

type WebhookResourceModel struct {
	ID          types.String `tfsdk:"id"`
	URL         types.String `tfsdk:"url"`
	EventTypes  types.List   `tfsdk:"event_types"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
	SigningKey   types.String `tfsdk:"signing_key"`
}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The webhook ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The webhook URL to receive events.",
			},
			"event_types": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of event types to subscribe to.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Webhook status: enabled or disabled.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Webhook description.",
			},
			"signing_key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Secret key for signing webhook payloads.",
			},
		},
	}
}

func (r *WebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var eventTypes []string
	resp.Diagnostics.Append(plan.EventTypes.ElementsAs(ctx, &eventTypes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.WebhookInput{
		URL:        plan.URL.ValueString(),
		EventTypes: eventTypes,
	}
	if !plan.Status.IsNull() {
		input.Status = plan.Status.ValueString()
	}
	if !plan.Description.IsNull() {
		input.Description = plan.Description.ValueString()
	}
	if !plan.SigningKey.IsNull() {
		input.SigningKey = plan.SigningKey.ValueString()
	}

	var webhook client.Webhook
	if err := r.client.Post(ctx, "/v1/webhooks", input, &webhook); err != nil {
		resp.Diagnostics.AddError("Error creating webhook", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, webhookToModel(ctx, &webhook, &resp.Diagnostics))...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook client.Webhook
	if err := r.client.Get(ctx, "/v1/webhooks/"+state.ID.ValueString(), &webhook); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading webhook", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, webhookToModel(ctx, &webhook, &resp.Diagnostics))...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookResourceModel
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var eventTypes []string
	resp.Diagnostics.Append(plan.EventTypes.ElementsAs(ctx, &eventTypes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.WebhookInput{
		URL:        plan.URL.ValueString(),
		EventTypes: eventTypes,
	}
	if !plan.Status.IsNull() {
		input.Status = plan.Status.ValueString()
	}
	if !plan.Description.IsNull() {
		input.Description = plan.Description.ValueString()
	}
	if !plan.SigningKey.IsNull() {
		input.SigningKey = plan.SigningKey.ValueString()
	}

	var webhook client.Webhook
	if err := r.client.Patch(ctx, "/v1/webhooks/"+state.ID.ValueString(), input, &webhook); err != nil {
		resp.Diagnostics.AddError("Error updating webhook", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, webhookToModel(ctx, &webhook, &resp.Diagnostics))...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete(ctx, "/v1/webhooks/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting webhook", err.Error())
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func webhookToModel(ctx context.Context, w *client.Webhook, diags *diag.Diagnostics) *WebhookResourceModel {
	eventTypes, d := types.ListValueFrom(ctx, types.StringType, w.EventTypes)
	diags.Append(d...)
	return &WebhookResourceModel{
		ID:          types.StringValue(w.ID),
		URL:         types.StringValue(w.URL),
		EventTypes:  eventTypes,
		Status:      types.StringValue(w.Status),
		Description: types.StringValue(w.Description),
		SigningKey:   types.StringValue(w.SigningKey),
	}
}
