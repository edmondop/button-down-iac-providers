terraform {
  required_providers {
    buttondown = {
      source  = "edmondop/buttondown"
      version = "~> 0.1"
    }
  }
}

provider "buttondown" {}

# Configure a newsletter with custom branding
resource "buttondown_newsletter" "my_newsletter" {
  name        = "Edmondo's Engineering Blog"
  username    = "edmondo"
  description = "Weekly insights on infrastructure, DevOps, and software engineering."

  locale        = "en"
  template      = "modern"
  archive_theme = "modern"
  tint_color    = "#0069FF"
  timezone      = "Europe/Rome"

  css = <<-EOT
    .email-body {
      font-family: 'Inter', sans-serif;
      line-height: 1.6;
    }
  EOT

  footer = "You're receiving this because you subscribed at edmondo.lol"
}

# External feed: auto-import blog posts as email drafts
resource "buttondown_external_feed" "blog_rss" {
  url      = "https://edmondo.lol/rss.xml"
  behavior = "draft"
  cadence  = "daily"
  subject  = "New blog post: {{title}}"
  body     = "{{content}}"
  label    = "Blog RSS"
}

# Automation: welcome email for new subscribers
resource "buttondown_automation" "welcome_email" {
  name    = "Welcome Email"
  trigger = "subscriber.created"
}

# Subscription form
resource "buttondown_form" "main_signup" {
  title        = "Subscribe to My Newsletter"
  slug         = "subscribe"
  body         = "Get weekly engineering insights delivered to your inbox."
  success_body = "Thanks for subscribing! Check your email to confirm."
  status       = "active"
}
