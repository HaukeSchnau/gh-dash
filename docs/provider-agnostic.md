# Provider-Agnostic Roadmap (GitHub + GitLab)

This document describes a concrete plan to evolve `gh-dash` into a provider-agnostic tool that can aggregate and manage work items (PRs/MRs + issues) across GitHub and multiple GitLab instances in a unified TUI.

It is written as an implementation plan (architecture + milestones), not as a design manifesto.

---

## Goals

1. **Provider-agnostic core**: A single UI that can show work items from multiple providers in the same screen.
2. **Multiple instances**: Support multiple GitLab instances (as configured via `glab`), and GitHub Enterprise if configured via `gh`.
3. **Unified query DSL**: Replace provider-specific search strings with a small DSL that can be translated into:
   - GitHub GraphQL search queries
   - GitLab REST API query parameters (or `glab api` as a fallback)
4. **Actions parity**: Built-in actions should work across providers as much as possible (minimum GitLab action set in milestone 1).
5. **Repo view parity**: Local branches view should integrate with GitLab too (branch-name matching is acceptable initially).
6. **Optional grouping**: Users can toggle “group by provider” via a **global keybinding**.

## Non-goals (initially)

- Strict global sorting across providers (concat per provider is acceptable for now).
- Backward compatibility with existing GitHub search strings in `filters:` (configs should migrate to DSL).
- Perfect feature parity on day 1 (but follow-up milestones should land quickly).

---

## Current Coupling (What Must Change)

The codebase is GitHub-coupled in three ways:

1. **Data fetching is GitHub GraphQL-specific**
   - Uses `github.com/cli/go-gh` GraphQL client and GitHub GraphQL schema types.
   - Query strings assume GitHub search syntax (e.g. `is:pr … sort:updated`).
2. **UI injects GitHub-specific filters and repo detection**
   - Smart filtering injects `repo:owner/name` using `go-gh`’s `repository.Current()`.
   - Repo short-name parsing assumes `https://github.com/...`.
3. **Actions shell out to `gh`**
   - Built-in PR/issue actions invoke `gh` directly and are named `GitHubTask`.

The plan below replaces these with provider-aware abstractions while preserving the UI structure and most UX behavior.

---

## Target Architecture

### Package layout (proposed)

- `internal/domain`
  - Provider-agnostic types the UI renders and updates.
  - Examples: `WorkItemKey`, `PullRequest`, `Issue`, `RepoRef`, `UserRef`, `LabelRef`, `ReviewState`, `CIState`, etc.
- `internal/providers`
  - Provider registry + interfaces.
  - `internal/providers/github` (backed by `gh` auth + GitHub GraphQL)
  - `internal/providers/gitlab` (backed by `glab` auth + GitLab API; multiple hosts)
- `internal/dsl`
  - Lexer/parser for the query DSL.
  - Normalized AST.
  - Translators:
    - `internal/dsl/github` → GitHub search query string
    - `internal/dsl/gitlab` → GitLab API params (and/or fallback `glab api`)
- `internal/actions`
  - Provider-agnostic action definitions + capability checks.
  - Provider-specific implementations (via CLI or API).
- `internal/git`
  - Replace GitHub-only remote parsing with a remote URL parser that can produce:
    - `host` (e.g. `github.com`, `gitlab.com`, `gitlab.mycorp.com`)
    - `projectPath` (e.g. `org/repo`, `group/subgroup/repo`)
    - `scheme` + `transport` (https/ssh)

### Provider interface (proposed)

Each provider instance is identified by a stable `ProviderID` (e.g. `github:github.com`, `gitlab:gitlab.com`, `gitlab:corp`).

Minimal interfaces for milestone 1:

- `Provider`
  - `ID() ProviderID`
  - `DisplayName() string`
  - `Host() string`
  - `Capabilities() Capabilities`
  - `CurrentUser(ctx) (UserRef, error)` (for `me/@me`)
- `WorkItemReader`
  - `ListPullRequests(ctx, QueryAST, Limit, Cursor?) ([]PullRequest, PageInfo, error)`
  - `ListIssues(ctx, QueryAST, Limit, Cursor?) ([]Issue, PageInfo, error)`
- `WorkItemActions`
  - PR/MR actions: approve, comment, close/reopen, merge, assign/unassign, label (as supported).
  - Issue actions: comment, close/reopen, assign/unassign, label (as supported).

Details views (files/checks/reviews) can be optional capabilities; the UI should degrade gracefully.

### UI strategy

- Keep the existing section/table UI structure, but change the row data type:
  - Replace direct usage of GitHub GraphQL structs with `internal/domain` models.
- Add a small provider marker (icon or short label) in row rendering to disambiguate mixed results.
- Add global keybinding to toggle “group by provider”:
  - OFF: sections show a single concatenated list: `[provider1 items..., provider2 items..., ...]`
  - ON: sections render provider subheaders and lists per provider (still within a unified screen).

### Data flow

1. Parse section `filters:` DSL → AST.
2. Expand `me/@me` per provider instance (provider-specific current user lookup).
3. Per provider instance:
   - Translate AST → provider query.
   - Fetch items.
4. Concatenate results in provider iteration order (order is not guaranteed stable initially).
5. Render domain objects into tables.

---

## Query DSL (v1)

### Requirements

- Provider-agnostic surface area.
- Supports `me` / `@me` (expanded per provider).
- Supports provider scoping in the DSL:
  - Example: `provider in ["github", "gitlab:corp"]`
- Supports project scoping with a unified field:
  - `project = "path"` where `path` is:
    - GitHub: `owner/repo`
    - GitLab: `group/subgroup/repo`
- Supports all three time syntaxes:
  - Relative: `updated > -7d`
  - Absolute: `updated >= 2025-12-01`
  - Function: `updated in last(7d)`

### Proposed DSL shape (illustrative, not final)

Operators:
- Comparisons: `= != > >= < <=`
- Boolean: `and`, `or`, `not`
- Membership: `in`, `not in`

Fields (initial candidates):
- `provider`
- `type` (pr/issue; in the UI the section type already scopes this)
- `project` (string)
- `state` (open/closed/merged; merged only applies to PR/MR)
- `author`, `assignee`, `review_requested`, `involves`
- `label` (string or list membership)
- `draft` (bool)
- `archived` (bool) (if supported)
- `updated`, `created` (time)
- `text` (free-text; optional)

### DSL grammar (v1)

This is the concrete spec we should implement (and test) so translation is deterministic.

Lexical:
- Whitespace separates tokens.
- Strings: double-quoted with `\"` and `\\` escapes.
- Identifiers: `[a-zA-Z_][a-zA-Z0-9_]*` (e.g. `review_requested`).
- Durations: `-?` + integer + unit in `{m,h,d,w}` (minutes/hours/days/weeks), e.g. `-7d`, `-3w`.
- Dates: `YYYY-MM-DD` (time-of-day and offsets can be added later if needed).

Grammar (EBNF-ish):
- `expr        := or_expr`
- `or_expr     := and_expr (("or" | "||") and_expr)*`
- `and_expr    := unary_expr (("and" | "&&") unary_expr)*`
- `unary_expr  := ("not" | "!") unary_expr | primary`
- `primary     := "(" expr ")" | predicate`
- `predicate   := ident op value | ident ("in" | "not in") list`
- `op          := "=" | "!=" | ">" | ">=" | "<" | "<="`
- `list        := "[" (value ("," value)*)? "]"`
- `value       := string | boolean | number | date | duration | function`
- `function    := "last(" duration ")"`

Type system (v1):
- `provider`: string or list[string]
- `project`: string
- `state`: string
- `author` / `assignee` / `review_requested` / `involves`: string (supports `me` and `@me`)
- `label`: string or list[string]
- `draft` / `archived`: boolean
- `updated` / `created`: date or duration or function `last(duration)`
- `text`: string

Normalization rules:
- Treat `me` and `@me` as the same token.
- Coerce `updated in last(7d)` and `updated > -7d` into canonical range predicates during normalization (e.g. `updated >= now-7d`).
- Expand provider shorthands (e.g. `github`) into concrete provider instance ids before provider filtering.

### Translation support matrix (v1)

We should keep a living matrix in code/docs to prevent accidental silent mismatches.

Legend:
- ✅ server-side (direct)
- 🟡 server-side (approx) + client-side refine
- 🟠 requires enrichment call(s) per item
- ❌ unsupported (error)

| DSL predicate | GitHub | GitLab |
|---|---:|---:|
| `project = "path"` | ✅ | ✅ |
| `state = open/closed` | ✅ | ✅ |
| `state = merged` | ✅ | ✅ |
| `author = me` | ✅ | ✅ |
| `assignee = me` | ✅ | ✅ |
| `review_requested = me` | ✅ | 🟡 |
| `involves = me` | ✅ | 🟡 |
| `label in ["a","b"]` | ✅ | ✅ |
| `draft = true/false` | ✅ | ✅ |
| `updated >= …` | ✅ | ✅ |
| `text = "foo"` | ✅ | 🟡 |

Policy:
- If a predicate is ❌ for a provider instance, fail that provider fetch with a clear “unsupported filter” error that identifies the predicate and provider.
- Never silently drop predicates.

### Translation notes

GitHub:
- Translate into GitHub search qualifiers + keywords where possible.
- For features that require GraphQL fields beyond search, fetch minimal list via search, then optionally enrich per item (defer this until after milestone 1 unless required).

GitLab:
- Prefer GitLab API list endpoints that can filter server-side:
  - `GET /merge_requests` and `GET /issues` (global scope=all)
  - Optionally `GET /projects/:id/merge_requests` for project-scoped queries and `source_branch=...` for repo view integration.
- When the API cannot express a DSL predicate, either:
  1) approximate server-side + filter client-side, or
  2) mark as unsupported and surface an error (avoid silent wrong results).

`me/@me` expansion:
- GitHub: via `gh`-backed GraphQL query for viewer login.
- GitLab: via `glab` auth context, then `GET /user` per host to resolve username.

---

## Configuration Plan

### Providers source of truth

- GitHub hosts and auth remain sourced from `gh` configuration (including GHES).
- GitLab hosts and auth remain sourced from `glab` configuration (`~/.config/glab-cli/config.yml`), supporting multiple instances.

### `gh-dash` config changes (proposed)

Add a `providers` block for selection and defaults, without storing credentials:

```yaml
providers:
  include:
    - github:*          # all gh hosts
    - gitlab:*          # all glab hosts
  exclude: []
  defaults:
    groupByProvider: false
```

Section configs remain, but `filters:` becomes DSL (no legacy GitHub strings).

Repo view:
- Determine provider instance from local `origin` remote host.

### Provider instance discovery (exact rules)

- GitHub instances:
  - Discovered via `gh` auth/config (GitHub.com and any configured GHES hosts).
  - Each host becomes a provider instance (e.g. `github:github.com`, `github:github.mycorp.com`).
- GitLab instances:
  - Discovered via `glab` config (`~/.config/glab-cli/config.yml`).
  - Each host becomes a provider instance (e.g. `gitlab:gitlab.com`, `gitlab:gitlab.mycorp.com`).
- `gh-dash` config should not store tokens; it only controls which discovered instances are enabled.

Matching for `providers.include` / `providers.exclude`:
- Exact instance id: `gitlab:gitlab.mycorp.com`
- Wildcard by provider: `gitlab:*`, `github:*`
- Optional provider alias: `gitlab`, `github` (expanded to `provider:*`)

---

## Milestones

### Milestone 0 — Refactor foundations (no new functionality)

1. **Remote parsing**
   - Replace GitHub-only `GetRepoShortName` with a parser that supports:
     - HTTPS remotes (GitHub/GitLab)
     - SSH remotes (`git@host:group/repo.git`)
   - Emit a normalized `{host, projectPath}`.
2. **Remove `repository.Current()` dependency**
   - Smart filtering should use the parsed origin remote rather than `go-gh` repository discovery.
3. **Introduce domain models**
   - Create `internal/domain` and adapt row rendering to accept domain objects (initially still sourced from GitHub).

Exit criteria:
- App behavior unchanged for GitHub-only usage, except smart filtering now uses origin parsing.

### Milestone 1 — Unified lists across GitHub + GitLab + minimum GitLab actions

Core:
1. **Provider registry**
   - Load provider instances from `gh` + `glab` configs.
   - Add include/exclude filters from `gh-dash` config.
2. **DSL parser + AST**
   - Implement parsing, normalization, and `me/@me` expansion.
3. **Readers**
   - GitHub reader: translate DSL → GitHub search query, fetch list items.
   - GitLab reader: translate DSL → GitLab API requests per host.
4. **Unified section fetch**
   - For PRs and issues sections, fetch from all enabled providers, concat results.
5. **Global “group by provider” keybinding**
   - Toggle between concatenated and grouped render modes.

Minimum GitLab actions (must-have in this milestone):
- MR: comment, approve, close/reopen, merge, assign/unassign, labels.
- Issue: comment, close/reopen, assign/unassign, labels.

Implementation approach for actions:
- Prefer invoking `glab` commands where they exist and are stable.
- Fill gaps via GitLab API calls using tokens from `glab` config.

Repo view (branch → MR lookup):
- For GitLab-origin repos: find open MR by `source_branch=<branch>` in the origin project (branch-name matching only).

Exit criteria:
- PR/MR and issue lists show items from GitHub + all configured GitLab instances in one UI.
- Core actions work on GitLab items (minimum set above).
- Repo view can open/find MR for a local branch when origin is GitLab.

### Milestone 2 — Details + parity expansion (fast follow)

1. PR/MR side panes:
   - Comments/activity rendering for GitLab.
   - Basic “files changed” and status/checks where feasible (capability-driven).
2. Additional actions parity:
   - Review request equivalents (if supported), ready-for-review/draft transitions (GitHub has it; GitLab has draft/WIP semantics).
3. Performance improvements:
   - Provider-level caching and rate-limit handling.

Exit criteria:
- Most existing GitHub features have equivalents or graceful degradation for GitLab.

### Milestone 3 — UX polish + correctness hardening

- Better provider indicators and filtering UX.
- More robust DSL error messages and “unsupported predicate” reporting.
- Optional stable sorting rules and pagination strategy per provider.
- Documentation updates + migration guide for DSL.

---

## Domain Identity & Keying (required for correctness)

We must define stable identifiers to avoid collisions between providers and between GitLab IID vs internal IDs.

Proposed identifiers:
- `ProviderID`: `"<provider>:<host>"` (e.g. `github:github.com`, `gitlab:gitlab.mycorp.com`)
- `ProjectPath`: path as seen in URLs:
  - GitHub: `owner/repo`
  - GitLab: `group/subgroup/repo`
- `WorkItemKind`: `{pull_request, issue}`
- `WorkItemNumber`:
  - GitHub: PR number / issue number
  - GitLab: IID (user-facing per-project number)
- `WorkItemKey`: `{ProviderID, ProjectPath, WorkItemKind, WorkItemNumber}`

Policy:
- All caches, selection state, updates, and action routing key off `WorkItemKey`.
- Do not use GitLab global IDs as the primary UI key.

---

## Capabilities Model (how we keep one UI without forcing parity)

Add a `Capabilities` struct per provider instance that drives:
- Which columns are renderable (review status, CI status, additions/deletions, etc.).
- Which actions are enabled (and shown in help).
- Which detail panes can show enriched data (files/checks/reviews/activity).

The UI should not assume GitHub-only fields exist; it should render optional fields only when supported and present.

---

## Actions Mapping (v1 inventory)

This list should become the definitive “built-in actions” contract, implemented provider-by-provider.

PR/MR actions:
- Open in browser
- Comment
- Approve
- Close / Reopen
- Merge
- Assign / Unassign
- Label add/remove
- Mark ready / convert draft (provider-dependent; likely milestone 2)
- Update branch (GitHub-specific; evaluate GitLab equivalent later)

Issue actions:
- Open in browser
- Comment
- Close / Reopen
- Assign / Unassign
- Label add/remove

Implementation policy:
- Prefer provider CLI (`gh`, `glab`) when it supports the action on that host.
- Use provider API when CLI coverage is missing, using tokens from the CLI’s config source of truth.
- If an action is unsupported for a provider, surface a clear error and disable it in help for that provider/item.

---

## Pagination & Ordering (multi-provider)

Milestone 1: concat per provider instance.

Rules:
- Each provider fetch runs independently with its own cursor/page info.
- “Next page” should page within a provider deterministically; avoid duplicates/skips within a provider.
- If one provider has no more pages, others can continue paging.

Grouped mode:
- “Next page” pages the currently focused provider group.

Concatenated mode:
- “Next page” can page the provider that owns the currently selected row (simplest and predictable).

---

## Performance & Resiliency

Concurrency:
- Fetch providers in parallel with a configurable concurrency limit (avoid spiky rate limits).

Caching:
- Cache `me/@me` resolution per provider instance.
- Cache GitLab project path → project ID lookups.
- Optional short-lived list result caching per provider+DSL+limit for refresh.

Failure isolation:
- One provider failing (auth, rate limit, unsupported predicate) must not blank out other providers’ results.
- Surface provider-scoped errors in the UI.

Retries:
- Add backoff for transient HTTP failures (429/5xx) for read operations.
- Do not automatically retry destructive actions.

---

## Testing Plan

DSL:
- Unit tests for parser (precedence, quoting, list syntax, all three time syntaxes).
- Unit tests for normalization (`me/@me`, `last()` rewrite, provider shorthand expansion).

Translators:
- Golden tests: AST → provider query (GitHub qualifiers; GitLab params).
- “Unsupported predicate” tests per provider.

Providers:
- Fixture-driven tests using mocked HTTP servers and/or stubbed CLI runners.
- Contract tests to ensure required domain fields are populated for UI rendering.

TUI:
- Minimal regression tests for:
  - concatenated vs grouped rendering
  - provider toggle keybinding
  - action routing by `WorkItemKey`

---

## Migration Notes (user-facing)

- `filters:` become DSL only (no legacy GitHub search strings).
- Provide docs with examples mapping common prior filters to DSL:
  - `is:open author:@me` → `state = open and author = me`
  - `review-requested:@me` → `review_requested = me`
  - `repo:owner/repo` → `project = "owner/repo"`

The app should emit a clear parse error when `filters:` is not valid DSL and point to the docs.

---

## Custom Keybindings: Provider-Aware Template Vars

Extend keybinding template inputs to include provider context so user commands can work across GitHub/GitLab:
- `ProviderID`
- `ProviderHost`
- `ProviderName`
- `ProjectPath` (unified replacement for `RepoName`)
- `WorkItemKind` (`pr`/`issue`)
- `WorkItemNumber` (PR number or MR IID / issue number)
- `WorkItemURL`

This also enables users to route to `gh` vs `glab` in their own commands.

---

## Risks & Mitigations

- **DSL expressiveness mismatch** (GitHub search vs GitLab API filters)
  - Mitigation: explicitly define which predicates are supported per provider; error on unsupported combinations.
- **Multiple instances auth**
  - Mitigation: treat `glab` config as the source of truth; include/exclude in `gh-dash` config.
- **Action parity gaps**
  - Mitigation: implement actions via API when `glab` lacks a direct command; keep action layer provider-agnostic.
- **UI regressions during refactor**
  - Mitigation: migrate incrementally (Milestone 0) with a GitHub provider that reuses the existing data access patterns.

---

## Implementation Checklist (per area)

- [ ] `internal/git`: remote URL parser → `{host, projectPath}`
- [ ] `internal/domain`: domain models + stable `WorkItemKey`
- [ ] `internal/dsl`: parser + AST + `me/@me` expansion
- [ ] `internal/providers`: registry + GitHub/GitLab instances
- [ ] `internal/providers/github`: list items + current user + actions via `gh`
- [ ] `internal/providers/gitlab`: list items + current user + actions via `glab`/API
- [ ] `internal/tui`: section fetch refactor to multi-provider concat/grouped rendering
- [ ] `internal/tui/keys`: global toggle for group-by-provider
