---
title: "feat: Buttondown IaC Providers (Terraform + Pulumi)"
type: feat
date: 2026-03-11
brainstorm: ../brainstorms/2026-03-11-buttondown-iac-providers-brainstorm.md
---

# Buttondown IaC Providers

## Overview

Build Terraform and Pulumi providers for Buttondown's newsletter API.
Terraform provider uses HashiCorp's open-source codegen pipeline
(`tfplugingen-openapi` + `tfplugingen-framework`) for schema scaffolding,
with hand-written CRUD logic. Pulumi provider wraps the TF provider via
`pulumi-terraform-bridge`. Full CI/CD with mise + GitHub Actions.

## Technical Approach

### Architecture

```
Buttondown OpenAPI spec (git submodule)
  → generator_config.yml (maps operations to TF resources)
  → tfplugingen-openapi (produces provider_code_spec.json)
  → tfplugingen-framework (produces Go schema scaffolding)
  → hand-written CRUD logic + HTTP client
  → terraform-provider-buttondown binary
  → pulumi-terraform-bridge (wraps TF provider)
  → pulumi-provider-buttondown + SDKs (TS, Python, Go, C#)
```

### Key Technical Decisions

- **HashiCorp codegen** over Speakeasy (Speakeasy TF gen is $250/mo)
- **terraform-plugin-framework** (modern, not legacy SDK)
- **Pulumi bridge with `pf.ShimProvider()`** (confirmed compatible)
- **Newsletters**: list-then-filter for Read (no GET-by-ID endpoint)
- **Emails**: default `status: "draft"` on create; no send action in v1

## Implementation Phases

### Phase 1: Repository Scaffolding

Foundation: repo, tooling, CI/CD — no provider code yet.

#### 1.1 Initialize git repo and monorepo structure

Create the directory structure and initialize Go modules.

```
/
├── openapi/                            # git submodule (already exists)
├── terraform-provider-buttondown/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   └── internal/
│       ├── provider/
│       └── client/                     # Buttondown API client
├── pulumi-provider-buttondown/
│   ├── provider/
│   │   ├── resources.go
│   │   ├── go.mod
│   │   └── cmd/
│   │       ├── pulumi-tfgen-buttondown/
│   │       └── pulumi-resource-buttondown/
│   └── sdk/                            # generated (TS, Python, Go, C#)
├── examples/
│   ├── terraform/
│   └── pulumi/
├── docs/
├── generator_config.yml
├── .mise.toml
├── .goreleaser.yml
├── .pre-commit-config.yaml
├── terraform-registry-manifest.json
└── README.md
```

**Files to create:**

- `terraform-provider-buttondown/go.mod` — Go module `github.com/edmondop/terraform-provider-buttondown`
- `terraform-provider-buttondown/main.go` — provider entry point (minimal, just serves the provider)
- `.mise.toml` — tool versions (Go 1.23, Node 22) + task definitions
- `.goreleaser.yml` — Terraform provider release config (ZIP archives, GPG signing, registry manifest)
- `terraform-registry-manifest.json` — `{"version": 1, "metadata": {"protocol_versions": ["6.0"]}}`
- `.pre-commit-config.yaml` — gofmt, golangci-lint, terraform_fmt, conventional commits
- `.gitignore` — Go + Terraform + Node artifacts

**Verification:** `cd terraform-provider-buttondown && go build ./...` compiles.

#### 1.2 Create GitHub repository

```bash
GH_TOKEN=$(gh auth token --user edmondop) gh repo create edmondop/button-down-iac-providers \
  --public --description "Terraform and Pulumi providers for Buttondown" \
  --source . --remote origin --push
```

#### 1.3 Set up GitHub Actions

Four workflow files:

**`.github/workflows/pr.yml`** — on PR to main:
- `jdx/mise-action@v2` installs tools
- Jobs (parallel): lint (`mise run lint`), test (`mise run test`), fmt check

**`.github/workflows/acceptance.yml`** — on push to main:
- Gated by `BUTTONDOWN_API_KEY` secret
- `TF_ACC=1 mise run testacc`
- 120min timeout

**`.github/workflows/release.yml`** — on tag `v*`:
- GoReleaser with GPG signing
- `crazy-max/ghaction-import-gpg@v6` for key import
- `goreleaser/goreleaser-action@v6`

**`.github/workflows/codegen-regen.yml`** — on push to main with `openapi/**` path changes:
- Runs `mise run generate`
- Opens PR via `peter-evans/create-pull-request@v7` if code changed

**Verification:** Push a branch, open a PR, confirm lint + test jobs run green (even if they just pass trivially).

---

### Phase 2: API Client

A thin Go HTTP client for Buttondown's API. Used by all Terraform resources.

#### 2.1 Implement the Buttondown API client

**File:** `terraform-provider-buttondown/internal/client/client.go`

```go
type Client struct {
    BaseURL    string
    APIKey     string
    HTTPClient *http.Client
}

func New(apiKey string, opts ...Option) *Client
func (c *Client) Get(ctx context.Context, path string, result any) error
func (c *Client) Post(ctx context.Context, path string, body any, result any) error
func (c *Client) Patch(ctx context.Context, path string, body any, result any) error
func (c *Client) Delete(ctx context.Context, path string) error
func (c *Client) List(ctx context.Context, path string, result any) error
```

Design:
- API key via `Authorization: Token <key>` header
- Base URL defaults to `https://api.buttondown.com/v1`
- JSON encoding/decoding
- Error type with status code + Buttondown error message
- Pagination support for List (Buttondown uses `next` URL pattern)
- No retry logic in v1 (keep it simple)

**File:** `terraform-provider-buttondown/internal/client/client_test.go`

Unit tests with `httptest.Server` mocking Buttondown responses.

**Verification:** `go test ./internal/client/... -v` passes.

#### 2.2 Define API model types

**File:** `terraform-provider-buttondown/internal/client/models.go`

Go structs matching the OpenAPI schemas for all 11 resources + Account:

- `Email`, `EmailInput`, `EmailUpdateInput`
- `Automation`, `AutomationInput`, `AutomationUpdateInput`
- `Tag`, `TagInput`, `TagUpdateInput`
- `Webhook`, `WebhookInput`
- `ExternalFeed`, `ExternalFeedInput`
- `Form`, `FormInput`
- `Newsletter`, `NewsletterInput`
- `Snippet`, `SnippetInput`
- `Survey`, `SurveyInput`
- `Book`, `BookInput`
- `User`, `UserInput`
- `Account`
- `PageResponse[T]` (generic pagination wrapper)

Derive field names and types from the OpenAPI spec schemas. Use `json` struct tags matching the API's field names.

**Verification:** Models compile and can be round-tripped through `json.Marshal` / `json.Unmarshal`.

---

### Phase 3: Code Generation + Provider Scaffolding

Use HashiCorp's codegen tools to generate Terraform schema code from the OpenAPI spec.

#### 3.1 Write generator_config.yml

**File:** `generator_config.yml`

Maps OpenAPI operations to Terraform resources and data sources. Example structure:

```yaml
provider:
  name: buttondown

resources:
  email:
    create:
      path: /emails
      method: POST
    read:
      path: /emails/{id}
      method: GET
    update:
      path: /emails/{id}
      method: PATCH
    delete:
      path: /emails/{id}
      method: DELETE
  # ... repeat for all 11 resources

data_sources:
  account:
    read:
      path: /accounts/me
      method: GET
```

**Special cases:**
- `newsletter`: Read maps to `GET /newsletters` (list) — no GET-by-ID
- `email`: `EmailInput` should override `status` default to `"draft"`

**Note:** Exact YAML schema depends on `tfplugingen-openapi` version. Consult
`go doc` or `--help` output for the current config format.

#### 3.2 Run code generation

```bash
# Step 1: OpenAPI → Provider Code Spec
tfplugingen-openapi generate \
  --config generator_config.yml \
  --output provider_code_spec.json \
  openapi/openapi.json

# Step 2: Provider Code Spec → Go scaffolding
tfplugingen-framework generate all \
  --input provider_code_spec.json \
  --output terraform-provider-buttondown/internal/provider
```

Add these as `mise run generate` task.

**Verification:** Generated `.go` files exist in `internal/provider/` and compile.

#### 3.3 Wire up the provider

**File:** `terraform-provider-buttondown/internal/provider/provider.go`

The provider configuration:
- Schema: `api_key` (sensitive, optional — falls back to `BUTTONDOWN_API_KEY` env var), `base_url` (optional)
- `Configure`: creates `client.Client` and stores it in provider data
- `Resources`: returns list of all 11 resource constructors
- `DataSources`: returns `[]*schema.DataSource{NewAccountDataSource}`

**File:** `terraform-provider-buttondown/main.go`

```go
func main() {
    providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
        Address: "registry.terraform.io/edmondop/buttondown",
    })
}
```

**Verification:** `go build ./...` succeeds. `terraform providers schema -json` shows the provider schema.

---

### Phase 4: Terraform Resources (CRUD Implementation)

Implement each resource. Each resource follows the same pattern:
1. Schema (from codegen or hand-written)
2. Create / Read / Update / Delete methods
3. Unit test with mocked HTTP
4. Acceptance test with real API (gated)

#### 4.1 Account data source

**Files:**
- `internal/provider/account_data_source.go`
- `internal/provider/account_data_source_test.go`

Simplest resource — singleton GET only. Good first implementation to validate the pattern.

Schema: `username` (computed string), `email_address` (computed string).
Read: `GET /accounts/me` → populate state.

**Verification:** `TF_ACC=1 go test ./internal/provider/ -run TestAccAccountDataSource -v`

#### 4.2 Tags resource

**Files:**
- `internal/provider/tag_resource.go`
- `internal/provider/tag_resource_test.go`

Clean full CRUD, simplest resource type. Good second implementation.

Schema: `id` (computed), `name` (required), `color` (optional), `description` (optional).
- Create: `POST /tags`
- Read: `GET /tags/{id}`
- Update: `PATCH /tags/{id}`
- Delete: `DELETE /tags/{id}`

**Verification:** Unit tests + acceptance test creating, reading, updating, deleting a tag.

#### 4.3 Webhooks resource

**Files:**
- `internal/provider/webhook_resource.go`
- `internal/provider/webhook_resource_test.go`

Full CRUD. Schema: `id`, `url` (required), `events` (list of strings), `status` (enum).

#### 4.4 Snippets resource

**Files:**
- `internal/provider/snippet_resource.go`
- `internal/provider/snippet_resource_test.go`

Full CRUD. Schema: `id`, `name`, `body`, `mode` (enum: fancy/naked/plaintext).

#### 4.5 Automations resource

**Files:**
- `internal/provider/automation_resource.go`
- `internal/provider/automation_resource_test.go`

Full CRUD. Complex schema — has `timing`, `actions`, `filters`. Derive from OpenAPI.

#### 4.6 Forms resource

**Files:**
- `internal/provider/form_resource.go`
- `internal/provider/form_resource_test.go`

Full CRUD. Schema: `id`, `name`, `description`, `status`.

#### 4.7 External Feeds resource

**Files:**
- `internal/provider/external_feed_resource.go`
- `internal/provider/external_feed_resource_test.go`

Full CRUD. Schema: `id`, `url`, `automation_behavior`, `automation_cadence`, etc.

#### 4.8 Surveys resource

**Files:**
- `internal/provider/survey_resource.go`
- `internal/provider/survey_resource_test.go`

Full CRUD. Schema: `id`, `question`, `input_type`, `status`, `response_cadence`.

#### 4.9 Books resource

**Files:**
- `internal/provider/book_resource.go`
- `internal/provider/book_resource_test.go`

Full CRUD. Schema from OpenAPI BookInput/Book schemas.

#### 4.10 Users resource

**Files:**
- `internal/provider/user_resource.go`
- `internal/provider/user_resource_test.go`

Full CRUD. Schema: `id`, `email`, `permissions`, etc.

#### 4.11 Emails resource

**Files:**
- `internal/provider/email_resource.go`
- `internal/provider/email_resource_test.go`

CRUD with special behavior:
- Create defaults `status` to `"draft"` (overriding API default of `"about_to_send"`)
- Schema: `id`, `subject` (required, max 2000), `body` (required, HTML/markdown),
  `status` (computed, always "draft"), `slug`, `description`, `template` (enum),
  `commenting_mode` (enum), `metadata` (JSON object), `publish_date`, `image`
- Read: `GET /emails/{id}`
- Update: `PATCH /emails/{id}` — cannot change status
- Delete: `DELETE /emails/{id}`

#### 4.12 Newsletters resource

**Files:**
- `internal/provider/newsletter_resource.go`
- `internal/provider/newsletter_resource_test.go`

CRUD with list-based Read:
- Create: `POST /newsletters`
- Read: `GET /newsletters` → iterate list, find by ID
- Update: `PATCH /newsletters/{id}`
- Delete: `DELETE /newsletters/{id}`

The Read implementation must handle pagination if the user has many newsletters.

---

### Phase 5: Examples and Documentation

#### 5.1 HCL examples

**File:** `examples/terraform/main.tf`

Realistic example showing:
- Provider configuration
- Email draft resource
- Automation resource
- Tag resource
- Webhook resource
- Account data source

**File:** `examples/terraform/newsletter-setup/main.tf`

Full newsletter configuration example.

#### 5.2 Generate Terraform registry documentation

```bash
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
tfplugindocs generate --provider-name buttondown
```

Produces `docs/` directory with resource and data source documentation.

Add as `mise run docs` task.

#### 5.3 README

Top-level README with:
- What this is
- Quick start (provider config + first resource)
- Development setup (`mise install && mise run build`)
- Links to Terraform Registry and Pulumi Registry

---

### Phase 6: Pulumi Bridge

#### 6.1 Set up Pulumi bridge scaffolding

Clone from `pulumi-tf-provider-boilerplate` or set up manually.

**File:** `pulumi-provider-buttondown/provider/resources.go`

```go
func ButtondownProvider() tfbridge.ProviderInfo {
    info := tfbridge.ProviderInfo{
        P:       pf.ShimProvider(<buttondown provider constructor>),
        Name:    "buttondown",
        Version: "1.0.0",
        MetadataInfo: tfbridge.NewProviderMetadata(bridgeMetadata),
    }
    info.MustComputeTokens(tokens.SingleModule("buttondown_", "index",
        tokens.MakeStandard("buttondown")))
    return info
}
```

**If provider constructor is in `internal/`:** Create a shim package at
`pulumi-provider-buttondown/provider/shim/` that imports and re-exports it.

**Files:**
- `cmd/pulumi-tfgen-buttondown/main.go` — build-time schema generator
- `cmd/pulumi-resource-buttondown/main.go` — runtime resource plugin

#### 6.2 Generate SDKs

```bash
make tfgen      # generates schema.json + bridge-metadata.json
make provider   # builds pulumi-resource-buttondown binary
make build_sdks # generates TS, Python, Go, C# SDKs
```

Add as `mise run pulumi-build` task.

**Verification:** SDKs generate without errors. TypeScript SDK compiles.

#### 6.3 Pulumi examples

**File:** `examples/pulumi/typescript/index.ts`

```typescript
import * as buttondown from "@edmondop/pulumi-buttondown";

const tag = new buttondown.Tag("engineering", {
    name: "Engineering",
    color: "#0066cc",
});

const draft = new buttondown.Email("weekly-update", {
    subject: "Weekly Engineering Update",
    body: "# This Week...",
});
```

---

### Phase 7: Blog Post Outline (Deferred Execution)

- [ ] Verify blog format conventions at `~/code/edmondoblog/src/content/config.ts`
- [ ] Write draft post at `~/code/edmondoblog/src/content/blog/YYYY-MM-DD-git-configuration-management.md`
- [ ] Include real HCL examples from the working provider
- [ ] Include before/after screenshots (Buttondown UI vs HCL)
- [ ] Review with `draft: true` frontmatter flag

See brainstorm document for full outline.

## Acceptance Criteria

### Functional Requirements

- [ ] `terraform init` with the provider succeeds
- [ ] `terraform plan` and `terraform apply` work for all 11 resources
- [ ] Email resources default to `status: "draft"` — never auto-send
- [ ] Newsletter resource handles list-based Read correctly
- [ ] Account data source returns correct username and email
- [ ] `terraform import` works for all resources with known IDs
- [ ] `terraform destroy` cleans up all created resources
- [ ] Pulumi bridge generates working TypeScript, Python, Go, C# SDKs
- [ ] `pulumi up` works with the TypeScript example

### Non-Functional Requirements

- [ ] Acceptance tests pass against real Buttondown API
- [ ] CI passes on PR (lint + unit tests)
- [ ] CI runs acceptance tests on merge to main
- [ ] GoReleaser produces valid release artifacts on tag
- [ ] Provider documentation generates for Terraform Registry

### Quality Gates

- [ ] Every resource has unit tests (mocked HTTP)
- [ ] Every resource has at least one acceptance test (real API)
- [ ] `golangci-lint` passes with zero warnings
- [ ] `terraform fmt` produces no diffs on examples
- [ ] Pre-commit hooks pass

## Dependencies and Prerequisites

| Dependency | Status | Notes |
|------------|--------|-------|
| Buttondown API key | Required for acceptance tests | Set as `BUTTONDOWN_API_KEY` |
| GPG key pair | Required for releases | For Terraform Registry signing |
| Go 1.23+ | Managed by mise | |
| Node 22+ | Managed by mise | For Pulumi SDK generation |
| `tfplugingen-openapi` | Install via `go install` | HashiCorp tech preview |
| `tfplugingen-framework` | Install via `go install` | HashiCorp tech preview |
| `pulumi-terraform-bridge` v3 | Go dependency | Confirmed: supports plugin-framework |

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| HashiCorp codegen is tech preview — may have bugs | Medium | Medium | Fall back to fully hand-written schemas. Codegen saves time but isn't required. |
| Newsletters list-based Read is slow | Low | Low | Personal account has few newsletters. Optimize with caching later if needed. |
| Pulumi bridge shim issues with `internal/` packages | Medium | Low | Documented pattern: `provider/shim/` package re-exports constructor. |
| Buttondown API changes break provider | Low | Medium | Pin to OpenAPI spec version. Acceptance tests catch regressions. |
| `tfplugingen-openapi` config format doesn't match our needs | Medium | Medium | Write schemas by hand — the codegen is a shortcut, not a hard dependency. |

## References

### Tools

- `tfplugingen-openapi`: `github.com/hashicorp/terraform-plugin-codegen-openapi`
- `tfplugingen-framework`: `github.com/hashicorp/terraform-plugin-codegen-framework`
- `terraform-plugin-framework`: `github.com/hashicorp/terraform-plugin-framework`
- `pulumi-terraform-bridge`: `github.com/pulumi/pulumi-terraform-bridge`
- `pulumi-tf-provider-boilerplate`: `github.com/pulumi/pulumi-tf-provider-boilerplate`

### Buttondown

- OpenAPI spec: `./openapi/openapi.json`
- API docs: `https://docs.buttondown.com/api-introduction`
- Fixtures: `./openapi/fixtures.json`
- Enums: `./openapi/enums.json`

### Brainstorm

- `docs/brainstorms/2026-03-11-buttondown-iac-providers-brainstorm.md`
