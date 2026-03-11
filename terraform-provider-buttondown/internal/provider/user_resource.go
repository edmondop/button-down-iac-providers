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
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

type UserResource struct{ client *client.Client }

type UserResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	EmailAddress          types.String `tfsdk:"email_address"`
	Status                types.String `tfsdk:"status"`
	SubscriberPermission  types.String `tfsdk:"permission_subscriber"`
	EmailPermission       types.String `tfsdk:"permission_email"`
	SendingPermission     types.String `tfsdk:"permission_sending"`
	StylingPermission     types.String `tfsdk:"permission_styling"`
	AdminPermission       types.String `tfsdk:"permission_administrivia"`
	AutomationsPermission types.String `tfsdk:"permission_automations"`
	SurveysPermission     types.String `tfsdk:"permission_surveys"`
	FormsPermission       types.String `tfsdk:"permission_forms"`
}

func NewUserResource() resource.Resource { return &UserResource{} }

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown team member.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"email_address": schema.StringAttribute{Required: true, Description: "Team member email address."},
			"status":        schema.StringAttribute{Computed: true, Description: "Invitation status."},
			"permission_subscriber":    schema.StringAttribute{Optional: true, Computed: true, Description: "Subscriber access: none, read, write."},
			"permission_email":         schema.StringAttribute{Optional: true, Computed: true, Description: "Email access: none, read, write."},
			"permission_sending":       schema.StringAttribute{Optional: true, Computed: true, Description: "Sending access: none, read, write."},
			"permission_styling":       schema.StringAttribute{Optional: true, Computed: true, Description: "Styling access: none, read, write."},
			"permission_administrivia": schema.StringAttribute{Optional: true, Computed: true, Description: "Admin access: none, read, write."},
			"permission_automations":   schema.StringAttribute{Optional: true, Computed: true, Description: "Automations access: none, read, write."},
			"permission_surveys":       schema.StringAttribute{Optional: true, Computed: true, Description: "Surveys access: none, read, write."},
			"permission_forms":         schema.StringAttribute{Optional: true, Computed: true, Description: "Forms access: none, read, write."},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func permissionsFromModel(m *UserResourceModel) client.Permissions {
	p := client.Permissions{}
	if !m.SubscriberPermission.IsNull() { p.Subscriber = m.SubscriberPermission.ValueString() }
	if !m.EmailPermission.IsNull() { p.Email = m.EmailPermission.ValueString() }
	if !m.SendingPermission.IsNull() { p.Sending = m.SendingPermission.ValueString() }
	if !m.StylingPermission.IsNull() { p.Styling = m.StylingPermission.ValueString() }
	if !m.AdminPermission.IsNull() { p.Administrivia = m.AdminPermission.ValueString() }
	if !m.AutomationsPermission.IsNull() { p.Automations = m.AutomationsPermission.ValueString() }
	if !m.SurveysPermission.IsNull() { p.Surveys = m.SurveysPermission.ValueString() }
	if !m.FormsPermission.IsNull() { p.Forms = m.FormsPermission.ValueString() }
	return p
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	input := client.UserInput{
		EmailAddress: plan.EmailAddress.ValueString(),
		Permissions:  permissionsFromModel(&plan),
	}
	var u client.User
	if err := r.client.Post(ctx, "/v1/users", input, &u); err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(&u))...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	var u client.User
	if err := r.client.Get(ctx, "/v1/users/"+state.ID.ValueString(), &u); err != nil {
		if client.IsNotFound(err) { resp.State.RemoveResource(ctx); return }
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(&u))...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	perms := permissionsFromModel(&plan)
	input := client.UserUpdateInput{
		Permissions: map[string]string{
			"subscriber": perms.Subscriber, "email": perms.Email,
			"sending": perms.Sending, "styling": perms.Styling,
			"administrivia": perms.Administrivia, "automations": perms.Automations,
			"surveys": perms.Surveys, "forms": perms.Forms,
		},
	}
	var u client.User
	if err := r.client.Patch(ctx, "/v1/users/"+state.ID.ValueString(), input, &u); err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userToModel(&u))...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	if err := r.client.Delete(ctx, "/v1/users/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func userToModel(u *client.User) *UserResourceModel {
	return &UserResourceModel{
		ID: types.StringValue(u.ID), EmailAddress: types.StringValue(u.EmailAddress),
		Status: types.StringValue(u.Status),
		SubscriberPermission: types.StringValue(u.Permissions.Subscriber),
		EmailPermission: types.StringValue(u.Permissions.Email),
		SendingPermission: types.StringValue(u.Permissions.Sending),
		StylingPermission: types.StringValue(u.Permissions.Styling),
		AdminPermission: types.StringValue(u.Permissions.Administrivia),
		AutomationsPermission: types.StringValue(u.Permissions.Automations),
		SurveysPermission: types.StringValue(u.Permissions.Surveys),
		FormsPermission: types.StringValue(u.Permissions.Forms),
	}
}
