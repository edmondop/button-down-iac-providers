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
	_ resource.Resource                = &SurveyResource{}
	_ resource.ResourceWithImportState = &SurveyResource{}
)

type SurveyResource struct{ client *client.Client }

type SurveyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Identifier      types.String `tfsdk:"identifier"`
	Question        types.String `tfsdk:"question"`
	Answers         types.List   `tfsdk:"answers"`
	Notes           types.String `tfsdk:"notes"`
	ResponseCadence types.String `tfsdk:"response_cadence"`
	Status          types.String `tfsdk:"status"`
	InputType       types.String `tfsdk:"input_type"`
}

func NewSurveyResource() resource.Resource { return &SurveyResource{} }

func (r *SurveyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_survey"
}

func (r *SurveyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown survey.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"identifier":       schema.StringAttribute{Required: true, Description: "Unique identifier (max 100 chars)."},
			"question":         schema.StringAttribute{Required: true, Description: "Survey question (max 500 chars)."},
			"answers":          schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "List of answer options."},
			"notes":            schema.StringAttribute{Optional: true, Computed: true, Description: "Internal notes."},
			"response_cadence": schema.StringAttribute{Optional: true, Computed: true, Description: "Response frequency: once or once_per_email."},
			"status":           schema.StringAttribute{Optional: true, Computed: true, Description: "Status: active or inactive."},
			"input_type":       schema.StringAttribute{Optional: true, Computed: true, Description: "Input type: radio, checkbox, or text."},
		},
	}
}

func (r *SurveyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SurveyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SurveyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var answers []string
	resp.Diagnostics.Append(plan.Answers.ElementsAs(ctx, &answers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.SurveyInput{
		Identifier: plan.Identifier.ValueString(), Question: plan.Question.ValueString(), Answers: answers,
	}
	if !plan.Notes.IsNull() {
		input.Notes = plan.Notes.ValueString()
	}
	if !plan.ResponseCadence.IsNull() {
		input.ResponseCadence = plan.ResponseCadence.ValueString()
	}
	if !plan.InputType.IsNull() {
		input.InputType = plan.InputType.ValueString()
	}
	var s client.Survey
	if err := r.client.Post(ctx, "/v1/surveys", input, &s); err != nil {
		resp.Diagnostics.AddError("Error creating survey", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, surveyToModel(ctx, &s, &resp.Diagnostics))...)
}

func (r *SurveyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SurveyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var s client.Survey
	if err := r.client.Get(ctx, "/v1/surveys/"+state.ID.ValueString(), &s); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading survey", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, surveyToModel(ctx, &s, &resp.Diagnostics))...)
}

func (r *SurveyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SurveyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.SurveyUpdateInput{}
	if !plan.Notes.Equal(state.Notes) {
		v := plan.Notes.ValueString()
		input.Notes = &v
	}
	if !plan.ResponseCadence.Equal(state.ResponseCadence) {
		v := plan.ResponseCadence.ValueString()
		input.ResponseCadence = &v
	}
	if !plan.Status.Equal(state.Status) {
		v := plan.Status.ValueString()
		input.Status = &v
	}
	if !plan.InputType.Equal(state.InputType) {
		v := plan.InputType.ValueString()
		input.InputType = &v
	}
	if !plan.Answers.Equal(state.Answers) {
		var answers []string
		resp.Diagnostics.Append(plan.Answers.ElementsAs(ctx, &answers, false)...)
		input.Answers = answers
	}
	var s client.Survey
	if err := r.client.Patch(ctx, "/v1/surveys/"+state.ID.ValueString(), input, &s); err != nil {
		resp.Diagnostics.AddError("Error updating survey", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, surveyToModel(ctx, &s, &resp.Diagnostics))...)
}

func (r *SurveyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SurveyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete(ctx, "/v1/surveys/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting survey", err.Error())
	}
}

func (r *SurveyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func surveyToModel(ctx context.Context, s *client.Survey, diags *diag.Diagnostics) *SurveyResourceModel {
	answers, d := types.ListValueFrom(ctx, types.StringType, s.Answers)
	diags.Append(d...)
	return &SurveyResourceModel{
		ID: types.StringValue(s.ID), Identifier: types.StringValue(s.Identifier),
		Question: types.StringValue(s.Question), Answers: answers,
		Notes: types.StringValue(s.Notes), ResponseCadence: types.StringValue(s.ResponseCadence),
		Status: types.StringValue(s.Status), InputType: types.StringValue(s.InputType),
	}
}
