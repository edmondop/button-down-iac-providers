package provider

import (
	"testing"

	"github.com/edmondop/terraform-provider-buttondown/internal/client"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestTagToModel_maps_all_fields(t *testing.T) {
	tag := &client.Tag{
		ID:                 "tag-1",
		Name:               "Engineering",
		Color:              "#0066cc",
		Description:        "desc",
		PublicDescription:  "pub desc",
		SubscriberEditable: true,
	}
	m := tagToModel(tag)
	if m.ID.ValueString() != "tag-1" {
		t.Errorf("ID: got %q", m.ID.ValueString())
	}
	if m.Name.ValueString() != "Engineering" {
		t.Errorf("Name: got %q", m.Name.ValueString())
	}
	if m.Color.ValueString() != "#0066cc" {
		t.Errorf("Color: got %q", m.Color.ValueString())
	}
	if m.SubscriberEditable.ValueBool() != true {
		t.Error("SubscriberEditable: expected true")
	}
}

func TestEmailToModel_maps_all_fields(t *testing.T) {
	email := &client.Email{
		ID:      "email-1",
		Subject: "Hello",
		Body:    "World",
		Status:  "draft",
		Slug:    "hello-world",
	}
	m := emailToModel(email)
	if m.ID.ValueString() != "email-1" {
		t.Errorf("ID: got %q", m.ID.ValueString())
	}
	if m.Subject.ValueString() != "Hello" {
		t.Errorf("Subject: got %q", m.Subject.ValueString())
	}
	if m.Status.ValueString() != "draft" {
		t.Errorf("Status: got %q", m.Status.ValueString())
	}
}

func TestNewsletterToModel_maps_all_fields(t *testing.T) {
	n := &client.Newsletter{
		ID:           "nl-1",
		Name:         "My Newsletter",
		Username:     "test",
		Description:  "A test newsletter",
		TintColor:    "#ff0000",
		Timezone:     "America/New_York",
		Template:     "modern",
		ArchiveTheme: "classic",
	}
	m := newsletterToModel(n)
	if m.ID.ValueString() != "nl-1" {
		t.Errorf("ID: got %q", m.ID.ValueString())
	}
	if m.TintColor.ValueString() != "#ff0000" {
		t.Errorf("TintColor: got %q", m.TintColor.ValueString())
	}
	if m.Timezone.ValueString() != "America/New_York" {
		t.Errorf("Timezone: got %q", m.Timezone.ValueString())
	}
}

func TestTagSchema_name_is_required(t *testing.T) {
	r := &TagResource{}
	resp := fwresource.SchemaResponse{}
	r.Schema(nil, fwresource.SchemaRequest{}, &resp)

	nameAttr := resp.Schema.Attributes["name"]
	if !nameAttr.IsRequired() {
		t.Error("name should be required")
	}
}

func TestTagSchema_id_is_computed(t *testing.T) {
	r := &TagResource{}
	resp := fwresource.SchemaResponse{}
	r.Schema(nil, fwresource.SchemaRequest{}, &resp)

	idAttr := resp.Schema.Attributes["id"]
	if !idAttr.IsComputed() {
		t.Error("id should be computed")
	}
	if idAttr.IsRequired() {
		t.Error("id should not be required")
	}
}

func TestNewsletterSchema_description_is_required(t *testing.T) {
	r := &NewsletterResource{}
	resp := fwresource.SchemaResponse{}
	r.Schema(nil, fwresource.SchemaRequest{}, &resp)

	for _, field := range []string{"name", "username", "description"} {
		attr := resp.Schema.Attributes[field]
		if !attr.IsRequired() {
			t.Errorf("%s should be required", field)
		}
	}
}

func TestProviderSchema_api_key_is_sensitive(t *testing.T) {
	p := &ButtondownProvider{version: "test"}
	resp := fwprovider.SchemaResponse{}
	p.Schema(nil, fwprovider.SchemaRequest{}, &resp)

	apiKeyAttr := resp.Schema.Attributes["api_key"]
	if !apiKeyAttr.IsSensitive() {
		t.Error("api_key should be sensitive")
	}
}
