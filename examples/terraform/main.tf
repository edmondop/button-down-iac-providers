terraform {
  required_providers {
    buttondown = {
      source  = "edmondop/buttondown"
      version = "~> 0.1"
    }
  }
}

provider "buttondown" {
  # API key can be set here or via BUTTONDOWN_API_KEY env var
  # api_key = "your-api-key-here"
}

# Read your account info
data "buttondown_account" "me" {}

output "account_username" {
  value = data.buttondown_account.me.username
}

# Create tags for organizing content
resource "buttondown_tag" "engineering" {
  name  = "Engineering"
  color = "#0066cc"
}

resource "buttondown_tag" "announcements" {
  name  = "Announcements"
  color = "#ff6600"
}

# Create a reusable snippet
resource "buttondown_snippet" "footer_cta" {
  identifier = "footer-cta"
  name       = "Footer CTA"
  content    = "Thanks for reading! Reply to this email if you have questions."
  mode       = "fancy"
}

# Create an email draft
resource "buttondown_email" "weekly_update" {
  subject     = "Weekly Engineering Update"
  body        = <<-EOT
    # This Week in Engineering

    Here's what happened this week...

    {{ footer-cta }}
  EOT
  description = "Weekly engineering team update"
}

# Set up a webhook for new subscribers
resource "buttondown_webhook" "new_subscriber_notify" {
  url         = "https://hooks.example.com/buttondown"
  event_types = ["subscriber.created", "subscriber.updated"]
  description = "Notify our CRM when subscribers change"
}

# Create a survey
resource "buttondown_survey" "content_feedback" {
  identifier      = "content-feedback"
  question        = "How useful was this email?"
  answers         = ["Very useful", "Somewhat useful", "Not useful"]
  response_cadence = "once_per_email"
  input_type      = "radio"
}
