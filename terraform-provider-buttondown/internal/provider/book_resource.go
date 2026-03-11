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
	_ resource.Resource                = &BookResource{}
	_ resource.ResourceWithImportState = &BookResource{}
)

type BookResource struct{ client *client.Client }

type BookResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	URL         types.String `tfsdk:"url"`
	ImageURL    types.String `tfsdk:"image_url"`
	Description types.String `tfsdk:"description"`
	Year        types.Int64  `tfsdk:"year"`
	ISBN        types.String `tfsdk:"isbn"`
	Shared      types.Bool   `tfsdk:"shared"`
}

func NewBookResource() resource.Resource { return &BookResource{} }

func (r *BookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_book"
}

func (r *BookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Buttondown book (reading list item).",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"title":       schema.StringAttribute{Required: true, Description: "Book title."},
			"url":         schema.StringAttribute{Optional: true, Computed: true, Description: "Link to the book."},
			"image_url":   schema.StringAttribute{Optional: true, Computed: true, Description: "Cover image URL."},
			"description": schema.StringAttribute{Optional: true, Computed: true, Description: "Book description."},
			"year":        schema.Int64Attribute{Optional: true, Computed: true, Description: "Publication year."},
			"isbn":        schema.StringAttribute{Optional: true, Computed: true, Description: "ISBN."},
			"shared":      schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the book is publicly visible."},
		},
	}
}

func (r *BookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.BookInput{Title: plan.Title.ValueString()}
	if !plan.URL.IsNull() {
		input.URL = plan.URL.ValueString()
	}
	if !plan.ImageURL.IsNull() {
		input.ImageURL = plan.ImageURL.ValueString()
	}
	if !plan.Description.IsNull() {
		input.Description = plan.Description.ValueString()
	}
	if !plan.Year.IsNull() {
		v := int(plan.Year.ValueInt64())
		input.Year = &v
	}
	if !plan.ISBN.IsNull() {
		input.ISBN = plan.ISBN.ValueString()
	}
	if !plan.Shared.IsNull() {
		v := plan.Shared.ValueBool()
		input.Shared = &v
	}
	var b client.Book
	if err := r.client.Post(ctx, "/v1/books", input, &b); err != nil {
		resp.Diagnostics.AddError("Error creating book", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, bookToModel(&b))...)
}

func (r *BookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var b client.Book
	if err := r.client.Get(ctx, "/v1/books/"+state.ID.ValueString(), &b); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading book", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, bookToModel(&b))...)
}

func (r *BookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := client.BookUpdateInput{}
	if !plan.Title.Equal(state.Title) {
		v := plan.Title.ValueString()
		input.Title = &v
	}
	if !plan.URL.Equal(state.URL) {
		v := plan.URL.ValueString()
		input.URL = &v
	}
	if !plan.ImageURL.Equal(state.ImageURL) {
		v := plan.ImageURL.ValueString()
		input.ImageURL = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		input.Description = &v
	}
	if !plan.Year.Equal(state.Year) {
		v := int(plan.Year.ValueInt64())
		input.Year = &v
	}
	if !plan.ISBN.Equal(state.ISBN) {
		v := plan.ISBN.ValueString()
		input.ISBN = &v
	}
	if !plan.Shared.Equal(state.Shared) {
		v := plan.Shared.ValueBool()
		input.Shared = &v
	}
	var b client.Book
	if err := r.client.Patch(ctx, "/v1/books/"+state.ID.ValueString(), input, &b); err != nil {
		resp.Diagnostics.AddError("Error updating book", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, bookToModel(&b))...)
}

func (r *BookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete(ctx, "/v1/books/"+state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting book", err.Error())
	}
}

func (r *BookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func bookToModel(b *client.Book) *BookResourceModel {
	m := &BookResourceModel{
		ID: types.StringValue(b.ID), Title: types.StringValue(b.Title),
		URL: types.StringValue(b.URL), ImageURL: types.StringValue(b.ImageURL),
		Description: types.StringValue(b.Description), ISBN: types.StringValue(b.ISBN),
		Shared: types.BoolValue(b.Shared),
	}
	if b.Year != nil {
		m.Year = types.Int64Value(int64(*b.Year))
	} else {
		m.Year = types.Int64Null()
	}
	return m
}
