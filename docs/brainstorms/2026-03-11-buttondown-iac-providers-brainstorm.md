# Buttondown IaC Providers

**Date:** 2026-03-11
**Status:** Design complete, ready for planning

## What We're Building

Terraform and Pulumi providers for Buttondown (buttondown.com) that let you manage
your newsletter entirely as code — configuration, automation rules, email drafts,
webhooks, and more. Version-controlled, PR-reviewed, reproducible.

### Scope

**Terraform provider** (`terraform-provider-buttondown`):
- Generated using HashiCorp's open-source codegen pipeline (tech preview):
  1. `tfplugingen-openapi` — OpenAPI spec → Provider Code Specification (JSON IR)
  2. `tfplugingen-framework` — Provider Code Spec → Go code (terraform-plugin-framework)
- We write a `generator_config.yml` mapping OpenAPI operations to TF resources
- Generated code provides schema scaffolding; we implement CRUD logic + HTTP client

**Pulumi provider** (`pulumi-provider-buttondown`):
- Built via `pulumi-terraform-bridge`, not a native provider
- Included in v1 scope (TypeScript, Python, Go, C# SDKs)

**Blog post** (deferred):
- Outline captured below; actual writing happens after the provider works
- Astro blog at ~/code/edmondoblog, markdown with YAML frontmatter

### Terraform Resources (v1)

| Resource | CRUD | Notes |
|----------|------|-------|
| Emails | CRU + D | Default to `status: draft`. Send is out of scope for v1. |
| Automations | Full CRUD | + analytics (data source potential) |
| Tags | Full CRUD | |
| Webhooks | Full CRUD | |
| External Feeds | Full CRUD | |
| Forms | Full CRUD | |
| Newsletters | CRUD (no GET single) | List-based read workaround |
| Snippets | Full CRUD | |
| Surveys | Full CRUD | |
| Books | Full CRUD | |
| Users | Full CRUD | Team member management |

### Terraform Data Sources (v1)

| Data Source | Notes |
|-------------|-------|
| Account | Singleton GET `/accounts/me`. Read-only. |

### Explicitly Excluded

| Resource | Reason |
|----------|--------|
| Subscribers | Highly mutable external state, not IaC-appropriate |
| Comments | User-generated content, not configuration |
| Survey Responses | User-generated, not manageable |
| Bulk Actions | Trigger/fire-and-forget, not stateful |
| Exports | Trigger, not stateful |
| Notes | No read or update, weak TF fit |
| Images | No read or update, weak TF fit |
| Prices | Create-only, no delete |
| API Requests | Read-only audit log |
| Events | Read-only event log |
| Coupons | Read-only Stripe mirror |
| Ping | Health check, not a resource |

## Why This Approach

### HashiCorp codegen pipeline (not Speakeasy)

- Speakeasy Terraform generation requires paid Enterprise plan ($250/mo) — rejected
- HashiCorp's `tfplugingen-openapi` + `tfplugingen-framework` are free and open source
- `generator_config.yml` maps OpenAPI operations to TF resources/data sources
  (acts as our inclusion filter — excluded resources simply aren't mapped)
- Generated code provides type-safe schema scaffolding using `terraform-plugin-framework`
- We implement CRUD logic manually, wiring up an HTTP client to Buttondown's API
- Upstream OpenAPI spec stays pristine as a git submodule

### Pulumi via Terraform Bridge (not native)

- One codebase, four SDK languages for free
- Bridge is the standard approach for community providers
- Terraform provider is the source of truth; Pulumi inherits it

### Email as Draft-Only (no send action in v1)

- `terraform apply` creating an email defaults to `status: draft`
- Sending is irreversible — bad fit for `terraform plan` → `apply` cycle
- Users send via UI or direct API call after reviewing the draft
- v2 could add a `buttondown_email_send` resource with explicit intent

### mise + GitHub Actions (not Dagger)

- Pipeline is linear: generate → build → test → release
- Dagger adds complexity without proportional value for this shape
- mise handles local polyglot tasks (Go, Node for Pulumi SDKs)
- GitHub Actions for CI with mise tasks as the execution layer

## Key Decisions

1. **HashiCorp codegen pipeline** — `tfplugingen-openapi` + `tfplugingen-framework` (free, open source)
2. **generator_config.yml** as the inclusion filter — no spec forking needed
3. **11 resources + 1 data source** in v1 (see tables above)
4. **Email defaults to draft** — no send capability in v1
5. **Pulumi bridge in v1** — ships alongside Terraform provider
6. **mise for tasks, GitHub Actions for CI** — no Dagger
7. **Full CI/CD pipeline** — lint, test on PR; acceptance tests on main (API key gated); GoReleaser; codegen regen workflow; pre-commit hooks
8. **Blog post deferred** — outline captured, written after provider works
9. **Repo created during implementation** — not upfront

## Open Questions (Resolved by Research)

1. ~~Speakeasy~~ → Using HashiCorp codegen pipeline instead (free, open source)
2. **Newsletters lack GET single** — must implement list-then-filter in
   the hand-written CRUD Read function (list all, match by ID)
3. **Pulumi bridge compatibility** — CONFIRMED: bridge supports
   `terraform-plugin-framework` natively via `pf.ShimProvider()`
4. **Acceptance test strategy** — hand-written tests using
   `terraform-plugin-testing` (codegen only generates schema, not tests)
5. **Shim layer for Pulumi** — if provider constructor is in `internal/`,
   need a `provider/shim/` package to re-export it for the bridge

## Blog Post Outline (Deferred)

**Working title:** "Git is Your Best Configuration Management Tool"

**Target:** `~/code/edmondoblog/src/content/blog/YYYY-MM-DD-git-configuration-management.md`

**Frontmatter:**
```yaml
title: "Git is Your Best Configuration Management Tool"
description: "Stop clicking through dashboards. Manage your SaaS tools the same way you manage infrastructure — with code, version control, and pull requests."
pubDate: "TBD"
tags: ["infrastructure-as-code", "terraform", "git", "devops"]
draft: true
```

**Structure:**
1. The problem: SaaS configuration drift (Buttondown as concrete example)
2. The thesis: Git's architecture (storage + UI + collab + hooks + data model)
   makes it the universal version control layer
3. The pattern: IaC for "non-infrastructure" services
4. Concrete example: Buttondown Terraform provider — before/after
5. The ecosystem: Kubernetes YAML, DNS via Terraform, feature flags as code,
   database migrations
6. When UIs are actually better (counterarguments)
7. Call to action: pick one SaaS tool you manage via UI, write a provider

## Repository Structure

```
/
├── openapi/                           # Buttondown OpenAPI spec (git submodule)
├── terraform-provider-buttondown/     # Speakeasy-generated Terraform provider
├── pulumi-provider-buttondown/        # Pulumi bridge wrapping the TF provider
├── examples/
│   ├── terraform/                     # Realistic HCL examples
│   └── pulumi/                        # TypeScript/Python examples
├── docs/
│   └── brainstorms/                   # This document
├── .github/workflows/                 # CI/CD pipelines
├── generator_config.yml                # OpenAPI → TF resource mapping
├── provider_code_spec.json            # Generated intermediate representation
├── .pre-commit-config.yaml
├── .goreleaser.yml
├── .mise.toml                         # Task runner config
└── README.md
```
