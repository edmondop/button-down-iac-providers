# Buttondown IaC Providers

Terraform and Pulumi providers for [Buttondown](https://buttondown.com) — manage your newsletter as code.

## Why

Stop clicking through dashboards. Version-control your newsletter configuration, email drafts, automation rules, and more. Review changes in PRs. Roll back mistakes with `git revert`.

## Terraform Provider

### Quick Start

```hcl
terraform {
  required_providers {
    buttondown = {
      source  = "edmondop/buttondown"
      version = "~> 0.1"
    }
  }
}

provider "buttondown" {
  # Set via BUTTONDOWN_API_KEY env var or here
}

resource "buttondown_email" "weekly_update" {
  subject = "Weekly Update"
  body    = file("emails/weekly-update.md")
}
```

### Resources

| Resource | Description |
|----------|-------------|
| `buttondown_email` | Email drafts (created as draft, never auto-sent) |
| `buttondown_automation` | Automation rules |
| `buttondown_tag` | Tags for organizing content |
| `buttondown_webhook` | Webhook subscriptions |
| `buttondown_snippet` | Reusable content blocks |
| `buttondown_form` | Subscription forms |
| `buttondown_external_feed` | RSS/Atom feed imports |
| `buttondown_survey` | Reader surveys |
| `buttondown_book` | Reading list items |
| `buttondown_user` | Team members |
| `buttondown_newsletter` | Newsletter configuration |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `buttondown_account` | Current account info |

## Pulumi Provider

Use any Buttondown resource from TypeScript, Python, Go, or C# via Pulumi's Terraform provider bridge.

### Quick Start

```bash
# Add the Buttondown provider to your Pulumi project
pulumi package add terraform-provider edmondop/buttondown
```

### TypeScript Example

```typescript
import * as buttondown from "@pulumi/buttondown";

const tag = new buttondown.Tag("engineering", {
    name: "Engineering",
    color: "#0066cc",
});

const email = new buttondown.Email("weekly-update", {
    subject: "Weekly Update",
    body: "# This Week\n\nHere's what happened...",
});
```

### Python Example

```python
import pulumi_buttondown as buttondown

tag = buttondown.Tag("engineering",
    name="Engineering",
    color="#0066cc",
)

email = buttondown.Email("weekly-update",
    subject="Weekly Update",
    body="# This Week\n\nHere's what happened...",
)
```

All 11 Terraform resources and the account data source are available in every Pulumi language.

## Development

```bash
# Install tools
mise install

# Build
mise run build

# Test
mise run test

# Acceptance tests (requires BUTTONDOWN_API_KEY)
mise run testacc

# Lint
mise run lint
```

## License

MIT
