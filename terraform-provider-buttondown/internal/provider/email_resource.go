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
	_ resource.Resource                = &EmailResource{}
	_ resource.ResourceWithImportState = &EmailResource{}
)

type EmailResource struct{ client *client.Client }

type EmailResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Subject        types.String `tfsdk:"subject"`
	Body           types.String `tfsdk:"body"`
	Status         types.String `tfsdk:"status"`
	Description    types.String `tfsdk:"description"`
	Slug           types.String `tfsdk:"slug"`
	CanonicalURL   types.String `tfsdk:"canonical_url"`
	Image          types.String `tfsdk:"image"`
	CommentingMode types.String `tfsdk:"commenting_mode"`
	Template       types.String `tfsdk:"template"`
}

func NewEmailResource() resource.Resource { return &EmailResource{} }

func (r *EmailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email"
}

func (r *EmailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown email draft. Emails are created with status 'draft' by default and cannot be sent via Terraform.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"subject":         schema.StringAttribute{Required: true, Description: "Email subject line (max 2000 chars)."},
			"body":            schema.StringAttribute{Optional: true, Computed: true, Description: "Email body content (HTML or Markdown)."},
			"status":          schema.StringAttribute{Computed: true, Description: "Email status. Always 'draft' when created via Terraform."},
			"description":     schema.StringAttribute{Optional: true, Computed: true, Description: "Email description / subtitle."},
			"slug":            schema.StringAttribute{Optional: true, Computed: true, Description: "URL slug (max 100 chars)."},
			"canonical_url":   schema.StringAttribute{Optional: true, Computed: true, Description: "Canonical URL for the email."},
			"image":           schema.StringAttribute{Optional: true, Computed: true, Description: "Hero image URL."},
			"commenting_mode": schema.StringAttribute{Optional: true, Computed: true, Description: "Commenting mode: disabled, enabled, enabled_for_paid_subscribers."},
			"template":        schema.StringAttribute{Optional: true, Computed: true, Description: "Email template: classic, modern, naked, plaintext."},
		},
	}
}

func (r *EmailResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EmailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EmailResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.EmailInput{
		Subject: plan.Subject.ValueString(),
		Status:  "draft",
	}
	if !plan.Body.IsNull() {
		input.Body = plan.Body.ValueString()
	}
	if !plan.Description.IsNull() {
		input.Description = plan.Description.ValueString()
	}
	if !plan.Slug.IsNull() {
		input.Slug = plan.Slug.ValueString()
	}
	if !plan.CanonicalURL.IsNull() {
		input.CanonicalURL = plan.CanonicalURL.ValueString()
	}
	if !plan.Image.IsNull() {
		input.Image = plan.Image.ValueString()
	}
	if !plan.CommentingMode.IsNull() {
		input.CommentingMode = plan.CommentingMode.ValueString()
	}

	var email client.Email
	if err := r.client.Post(ctx, "/v1/emails", input, &email); err != nil {
		resp.Diagnostics.AddError("Error creating email", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, emailToModel(&email))...)
}

func (r *EmailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EmailResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var email client.Email
	if err := r.client.Get(ctx, "/v1/emails/"+state.ID.ValueString(), &email); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading email", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, emailToModel(&email))...)
}

func (r *EmailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EmailResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.EmailUpdateInput{}
	if !plan.Subject.Equal(state.Subject) {
		v := plan.Subject.ValueString()
		input.Subject = &v
	}
	if !plan.Body.Equal(state.Body) {
		v := plan.Body.ValueString()
		input.Body = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		input.Description = &v
	}
	if !plan.Slug.Equal(state.Slug) {
		v := plan.Slug.ValueString()
		input.Slug = &v
	}
	if !plan.CanonicalURL.Equal(state.CanonicalURL) {
		v := plan.CanonicalURL.ValueString()
		input.CanonicalURL = &v
	}
	if !plan.Image.Equal(state.Image) {
		v := plan.Image.ValueString()
		input.Image = &v
	}
	if !plan.CommentingMode.Equal(state.CommentingMode) {
		v := plan.CommentingMode.ValueString()
		input.CommentingMode = &v
	}
	if !plan.Template.Equal(state.Template) {
		v := plan.Template.ValueString()
		input.Template = &v
	}
	var email client.Email
	if err := r.client.Patch(ctx, "/v1/emails/"+state.ID.ValueString(), input, &email); err != nil {
		resp.Diagnostics.AddError("Error updating email", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, emailToModel(&email))...)
}

func (r *EmailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EmailResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete(ctx, "/v1/emails/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting email", err.Error())
	}
}

func (r *EmailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func emailToModel(e *client.Email) *EmailResourceModel {
	return &EmailResourceModel{
		ID:             types.StringValue(e.ID),
		Subject:        types.StringValue(e.Subject),
		Body:           types.StringValue(e.Body),
		Status:         types.StringValue(e.Status),
		Description:    types.StringValue(e.Description),
		Slug:           types.StringValue(e.Slug),
		CanonicalURL:   types.StringValue(e.CanonicalURL),
		Image:          types.StringValue(e.Image),
		CommentingMode: types.StringValue(e.CommentingMode),
		Template:       types.StringValue(e.Template),
	}
}
