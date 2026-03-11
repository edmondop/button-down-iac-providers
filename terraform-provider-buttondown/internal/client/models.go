package client

import "time"

type PageResponse[T any] struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []T     `json:"results"`
}

// Account

type Account struct {
	Username     string `json:"username"`
	EmailAddress string `json:"email_address"`
}

// Tag

type Tag struct {
	ID                 string    `json:"id"`
	CreationDate       time.Time `json:"creation_date"`
	Name               string    `json:"name"`
	Color              string    `json:"color"`
	Description        string    `json:"description"`
	PublicDescription  string    `json:"public_description"`
	SubscriberEditable bool      `json:"subscriber_editable"`
	SecondaryID        int       `json:"secondary_id"`
}

type TagInput struct {
	Name               string `json:"name"`
	Color              string `json:"color"`
	Description        string `json:"description,omitempty"`
	PublicDescription  string `json:"public_description,omitempty"`
	SubscriberEditable *bool  `json:"subscriber_editable,omitempty"`
}

type TagUpdateInput struct {
	Name               *string `json:"name,omitempty"`
	Color              *string `json:"color,omitempty"`
	Description        *string `json:"description,omitempty"`
	PublicDescription  *string `json:"public_description,omitempty"`
	SubscriberEditable *bool   `json:"subscriber_editable,omitempty"`
}

// Email

type Email struct {
	ID                              string       `json:"id"`
	CreationDate                    time.Time    `json:"creation_date"`
	ModificationDate                time.Time    `json:"modification_date"`
	AbsoluteURL                     string       `json:"absolute_url"`
	Body                            string       `json:"body"`
	CanonicalURL                    string       `json:"canonical_url"`
	CommentingMode                  string       `json:"commenting_mode"`
	ReviewMode                      string       `json:"review_mode"`
	Description                     string       `json:"description"`
	Featured                        bool         `json:"featured"`
	Filters                         *FilterGroup `json:"filters"`
	Image                           string       `json:"image"`
	Metadata                        Metadata     `json:"metadata"`
	PublishDate                     *time.Time   `json:"publish_date"`
	RelatedEmailIDs                 []string     `json:"related_email_ids"`
	SecondaryID                     *int         `json:"secondary_id"`
	ShouldTriggerPayPerEmailBilling bool         `json:"should_trigger_pay_per_email_billing"`
	Slug                            string       `json:"slug"`
	Source                          string       `json:"source"`
	Status                          string       `json:"status"`
	Subject                         string       `json:"subject"`
	SuppressionReason               string       `json:"suppression_reason,omitempty"`
	Template                        string       `json:"template,omitempty"`
}

type EmailInput struct {
	Subject                         string       `json:"subject"`
	Body                            string       `json:"body,omitempty"`
	Status                          string       `json:"status,omitempty"`
	Description                     string       `json:"description,omitempty"`
	CanonicalURL                    string       `json:"canonical_url,omitempty"`
	Image                           string       `json:"image,omitempty"`
	Slug                            string       `json:"slug,omitempty"`
	CommentingMode                  string       `json:"commenting_mode,omitempty"`
	Metadata                        Metadata     `json:"metadata,omitempty"`
	Filters                         *FilterGroup `json:"filters,omitempty"`
	RelatedEmailIDs                 []string     `json:"related_email_ids,omitempty"`
	Featured                        *bool        `json:"featured,omitempty"`
	ShouldTriggerPayPerEmailBilling *bool        `json:"should_trigger_pay_per_email_billing,omitempty"`
}

type EmailUpdateInput struct {
	Subject                         *string      `json:"subject,omitempty"`
	Body                            *string      `json:"body,omitempty"`
	Description                     *string      `json:"description,omitempty"`
	CanonicalURL                    *string      `json:"canonical_url,omitempty"`
	Image                           *string      `json:"image,omitempty"`
	Slug                            *string      `json:"slug,omitempty"`
	CommentingMode                  *string      `json:"commenting_mode,omitempty"`
	ReviewMode                      *string      `json:"review_mode,omitempty"`
	Template                        *string      `json:"template,omitempty"`
	Metadata                        Metadata     `json:"metadata,omitempty"`
	Filters                         *FilterGroup `json:"filters,omitempty"`
	RelatedEmailIDs                 []string     `json:"related_email_ids,omitempty"`
	Featured                        *bool        `json:"featured,omitempty"`
	ShouldTriggerPayPerEmailBilling *bool        `json:"should_trigger_pay_per_email_billing,omitempty"`
}

// Automation

type Automation struct {
	ID                             string       `json:"id"`
	CreationDate                   time.Time    `json:"creation_date"`
	Name                           string       `json:"name"`
	Status                         string       `json:"status"`
	Trigger                        string       `json:"trigger"`
	Actions                        []Action     `json:"actions"`
	Filters                        *FilterGroup `json:"filters"`
	Metadata                       Metadata     `json:"metadata"`
	ShouldEvaluateFilterAfterDelay bool         `json:"should_evaluate_filter_after_delay"`
}

type AutomationInput struct {
	Name                           string       `json:"name"`
	Trigger                        string       `json:"trigger"`
	Actions                        []Action     `json:"actions"`
	Filters                        *FilterGroup `json:"filters"`
	Metadata                       Metadata     `json:"metadata,omitempty"`
	ShouldEvaluateFilterAfterDelay *bool        `json:"should_evaluate_filter_after_delay,omitempty"`
}

type AutomationUpdateInput struct {
	Name                           *string      `json:"name,omitempty"`
	Status                         *string      `json:"status,omitempty"`
	Trigger                        *string      `json:"trigger,omitempty"`
	Actions                        []Action     `json:"actions,omitempty"`
	Filters                        *FilterGroup `json:"filters,omitempty"`
	Metadata                       Metadata     `json:"metadata,omitempty"`
	ShouldEvaluateFilterAfterDelay *bool        `json:"should_evaluate_filter_after_delay,omitempty"`
}

// Webhook

type Webhook struct {
	ID           string    `json:"id"`
	CreationDate time.Time `json:"creation_date"`
	Status       string    `json:"status"`
	EventTypes   []string  `json:"event_types"`
	URL          string    `json:"url"`
	Description  string    `json:"description"`
	SigningKey   string    `json:"signing_key"`
}

type WebhookInput struct {
	EventTypes  []string `json:"event_types"`
	URL         string   `json:"url"`
	Status      string   `json:"status,omitempty"`
	Description string   `json:"description,omitempty"`
	SigningKey  string   `json:"signing_key,omitempty"`
}

// ExternalFeed

type ExternalFeed struct {
	ID              string            `json:"id"`
	CreationDate    time.Time         `json:"creation_date"`
	LastCheckedDate *time.Time        `json:"last_checked_date"`
	Status          string            `json:"status"`
	Behavior        string            `json:"behavior"`
	Cadence         string            `json:"cadence"`
	CadenceMetadata map[string]string `json:"cadence_metadata"`
	Filters         *FilterGroup      `json:"filters"`
	URL             string            `json:"url"`
	Subject         string            `json:"subject"`
	Body            string            `json:"body"`
	Label           string            `json:"label"`
	Metadata        Metadata          `json:"metadata"`
	SkipOldItems    bool              `json:"skip_old_items"`
}

type ExternalFeedInput struct {
	URL             string            `json:"url"`
	Behavior        string            `json:"behavior"`
	Cadence         string            `json:"cadence"`
	CadenceMetadata map[string]string `json:"cadence_metadata"`
	Filters         *FilterGroup      `json:"filters"`
	Subject         string            `json:"subject"`
	Body            string            `json:"body"`
	Label           string            `json:"label,omitempty"`
	Metadata        Metadata          `json:"metadata,omitempty"`
	SkipOldItems    *bool             `json:"skip_old_items,omitempty"`
}

type ExternalFeedUpdateInput struct {
	Behavior        *string           `json:"behavior,omitempty"`
	Cadence         *string           `json:"cadence,omitempty"`
	CadenceMetadata map[string]string `json:"cadence_metadata,omitempty"`
	Filters         *FilterGroup      `json:"filters,omitempty"`
	Subject         *string           `json:"subject,omitempty"`
	Body            *string           `json:"body,omitempty"`
	Label           *string           `json:"label,omitempty"`
	Status          *string           `json:"status,omitempty"`
	Metadata        Metadata          `json:"metadata,omitempty"`
	SkipOldItems    *bool             `json:"skip_old_items,omitempty"`
}

// Form

type Form struct {
	ID           string    `json:"id"`
	CreationDate time.Time `json:"creation_date"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Body         string    `json:"body"`
	CSS          string    `json:"css"`
	SuccessBody  string    `json:"success_body"`
	Surveys      []string  `json:"surveys"`
	Admin        bool      `json:"admin"`
	Status       string    `json:"status"`
}

type FormInput struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Body        string   `json:"body,omitempty"`
	CSS         string   `json:"css,omitempty"`
	SuccessBody string   `json:"success_body,omitempty"`
	Surveys     []string `json:"surveys,omitempty"`
	Admin       *bool    `json:"admin,omitempty"`
	Status      string   `json:"status,omitempty"`
}

type FormUpdateInput struct {
	Title       *string  `json:"title,omitempty"`
	Slug        *string  `json:"slug,omitempty"`
	Body        *string  `json:"body,omitempty"`
	CSS         *string  `json:"css,omitempty"`
	SuccessBody *string  `json:"success_body,omitempty"`
	Surveys     []string `json:"surveys,omitempty"`
	Admin       *bool    `json:"admin,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

// Newsletter

type Newsletter struct {
	ID                          string            `json:"id"`
	CreationDate                time.Time         `json:"creation_date"`
	APIKey                      string            `json:"api_key"`
	Name                        string            `json:"name"`
	Username                    string            `json:"username"`
	Description                 string            `json:"description"`
	Domain                      string            `json:"domain"`
	EmailAddress                string            `json:"email_address"`
	EmailDomain                 string            `json:"email_domain"`
	CSS                         string            `json:"css"`
	WebCSS                      string            `json:"web_css"`
	Footer                      string            `json:"footer"`
	Header                      string            `json:"header"`
	FromName                    string            `json:"from_name"`
	ReplyToAddress              string            `json:"reply_to_address"`
	Icon                        string            `json:"icon"`
	Image                       string            `json:"image"`
	Locale                      string            `json:"locale"`
	Template                    string            `json:"template"`
	ArchiveTheme                string            `json:"archive_theme"`
	TintColor                   string            `json:"tint_color"`
	Timezone                    string            `json:"timezone"`
	Metadata                    Metadata          `json:"metadata"`
	EmailThemeConfiguration     map[string]string `json:"email_theme_configuration"`
	ThemeConfiguration          map[string]string `json:"theme_configuration"`
	AnnouncementBarText         string            `json:"announcement_bar_text"`
	AnnouncementBarBgColor      string            `json:"announcement_bar_background_color"`
	AnnouncementBarVisibility   string            `json:"announcement_bar_visibility"`
	AuditingMode                string            `json:"auditing_mode"`
	TestMode                    bool              `json:"test_mode"`
	EnabledFeatures             []string          `json:"enabled_features"`
	SharingNetworks             []string          `json:"sharing_networks"`
	SubscriptionRedirectURL     string            `json:"subscription_redirect_url"`
	SubscriptionConfirmRedirect string            `json:"subscription_confirmation_redirect_url"`
}

type NewsletterInput struct {
	Name        string `json:"name"`
	Username    string `json:"username"`
	Description string `json:"description"`
}

type NewsletterUpdateInput struct {
	Name                      *string           `json:"name,omitempty"`
	Description               *string           `json:"description,omitempty"`
	Domain                    *string           `json:"domain,omitempty"`
	CSS                       *string           `json:"css,omitempty"`
	WebCSS                    *string           `json:"web_css,omitempty"`
	Footer                    *string           `json:"footer,omitempty"`
	Header                    *string           `json:"header,omitempty"`
	FromName                  *string           `json:"from_name,omitempty"`
	ReplyToAddress            *string           `json:"reply_to_address,omitempty"`
	Locale                    *string           `json:"locale,omitempty"`
	Template                  *string           `json:"template,omitempty"`
	ArchiveTheme              *string           `json:"archive_theme,omitempty"`
	TintColor                 *string           `json:"tint_color,omitempty"`
	Timezone                  *string           `json:"timezone,omitempty"`
	Metadata                  Metadata          `json:"metadata,omitempty"`
	EmailThemeConfiguration   map[string]string `json:"email_theme_configuration,omitempty"`
	ThemeConfiguration        map[string]string `json:"theme_configuration,omitempty"`
	AnnouncementBarText       *string           `json:"announcement_bar_text,omitempty"`
	AnnouncementBarBgColor    *string           `json:"announcement_bar_background_color,omitempty"`
	AnnouncementBarVisibility *string           `json:"announcement_bar_visibility,omitempty"`
	AuditingMode              *string           `json:"auditing_mode,omitempty"`
	TestMode                  *bool             `json:"test_mode,omitempty"`
}

// Snippet

type Snippet struct {
	ID             string    `json:"id"`
	CreationDate   time.Time `json:"creation_date"`
	Identifier     string    `json:"identifier"`
	Name           string    `json:"name"`
	Content        string    `json:"content"`
	Mode           string    `json:"mode"`
	ReferenceCount int       `json:"reference_count"`
}

type SnippetInput struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Content    string `json:"content,omitempty"`
	Mode       string `json:"mode,omitempty"`
}

type SnippetUpdateInput struct {
	Identifier *string `json:"identifier,omitempty"`
	Name       *string `json:"name,omitempty"`
	Content    *string `json:"content,omitempty"`
	Mode       *string `json:"mode,omitempty"`
}

// Survey

type Survey struct {
	ID                        string    `json:"id"`
	CreationDate              time.Time `json:"creation_date"`
	Identifier                string    `json:"identifier"`
	Question                  string    `json:"question"`
	ResponseCount             int       `json:"response_count"`
	Answers                   []string  `json:"answers"`
	Notes                     string    `json:"notes"`
	RandomizeAnswers          bool      `json:"randomize_answers"`
	ResponseCadence           string    `json:"response_cadence"`
	Status                    string    `json:"status"`
	IsFreeformResponseEnabled bool      `json:"is_freeform_response_enabled"`
	InputType                 string    `json:"input_type"`
}

type SurveyInput struct {
	Identifier                string   `json:"identifier"`
	Question                  string   `json:"question"`
	Answers                   []string `json:"answers"`
	Notes                     string   `json:"notes,omitempty"`
	ResponseCadence           string   `json:"response_cadence,omitempty"`
	IsFreeformResponseEnabled *bool    `json:"is_freeform_response_enabled,omitempty"`
	InputType                 string   `json:"input_type,omitempty"`
	RandomizeAnswers          *bool    `json:"randomize_answers,omitempty"`
}

type SurveyUpdateInput struct {
	Answers                   []string `json:"answers,omitempty"`
	Notes                     *string  `json:"notes,omitempty"`
	ResponseCadence           *string  `json:"response_cadence,omitempty"`
	Status                    *string  `json:"status,omitempty"`
	IsFreeformResponseEnabled *bool    `json:"is_freeform_response_enabled,omitempty"`
	InputType                 *string  `json:"input_type,omitempty"`
	RandomizeAnswers          *bool    `json:"randomize_answers,omitempty"`
}

// Book

type Book struct {
	ID           string    `json:"id"`
	CreationDate time.Time `json:"creation_date"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	ImageURL     string    `json:"image_url"`
	Description  string    `json:"description"`
	Year         *int      `json:"year"`
	ISBN         string    `json:"isbn"`
	Shared       bool      `json:"shared"`
}

type BookInput struct {
	Title       string `json:"title"`
	URL         string `json:"url,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Description string `json:"description,omitempty"`
	Year        *int   `json:"year,omitempty"`
	ISBN        string `json:"isbn,omitempty"`
	Shared      *bool  `json:"shared,omitempty"`
}

type BookUpdateInput struct {
	Title       *string `json:"title,omitempty"`
	URL         *string `json:"url,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Description *string `json:"description,omitempty"`
	Year        *int    `json:"year,omitempty"`
	ISBN        *string `json:"isbn,omitempty"`
	Shared      *bool   `json:"shared,omitempty"`
}

// User

type User struct {
	ID           string      `json:"id"`
	CreationDate time.Time   `json:"creation_date"`
	EmailAddress string      `json:"email_address"`
	Status       string      `json:"status"`
	LastLoggedIn *time.Time  `json:"last_logged_in"`
	Permissions  Permissions `json:"permissions"`
}

type UserInput struct {
	EmailAddress string      `json:"email_address"`
	Permissions  Permissions `json:"permissions"`
}

type UserUpdateInput struct {
	Permissions map[string]string `json:"permissions"`
}

// Shared types

type Metadata map[string]any

type FilterGroup struct {
	Filters   []Filter      `json:"filters"`
	Groups    []FilterGroup `json:"groups"`
	Predicate string        `json:"predicate"`
}

type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Action struct {
	Type     string   `json:"type"`
	Metadata Metadata `json:"metadata"`
	Timing   *Timing  `json:"timing,omitempty"`
}

type Timing struct {
	Time  string `json:"time"`
	Delay *Delay `json:"delay,omitempty"`
}

type Delay struct {
	Value     string `json:"value"`
	Unit      string `json:"unit"`
	TimeOfDay string `json:"time_of_day,omitempty"`
}

type Permissions struct {
	Subscriber    string `json:"subscriber"`
	Email         string `json:"email"`
	Sending       string `json:"sending"`
	Styling       string `json:"styling"`
	Administrivia string `json:"administrivia"`
	Automations   string `json:"automations"`
	Surveys       string `json:"surveys"`
	Forms         string `json:"forms"`
}

// Error response from the API

type APIError struct {
	StatusCode int
	Message    string
	Detail     string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return e.Message
}
