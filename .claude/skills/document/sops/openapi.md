# SOP: OpenAPI (Swagger) documentation

Applies to the HTTP contract, which is generated with `swag` from Go comments (`make docs` / `go generate ./...`, output in the git-ignored `docs/`):

| Where | What it produces |
|---|---|
| `internal/handler/*.go` | one `// Name godoc` annotation block per handler → one operation |
| `presenter/*.go` | struct field tags and trailing comments → request/response schemas |
| `cmd/api/main.go` | the API-level block (`@title`, `@version`, `@description`, `@host`, `@BasePath`) |

Loaded by the `document` skill. Plain Go doc rules for the same files are in `internal.md`; this SOP governs the annotations. These comments are **generated output**: they must match the code exactly, and the format below is the project standard (derived from the existing handlers).

## Operation block (handlers)

Placed directly above the handler constructor. The first line is `// <FuncName> godoc`, then an empty `//` line, then a tab-indented annotation block (gofmt formats it this way; keep the tab-aligned columns):

```go
// AddFeed godoc
//
//	@Summary		Add a new RSS feed
//	@Description	Fetches the RSS feed at the given URL and persists it along with its articles
//	@Tags			feeds
//	@Accept			json
//	@Produce		json
//	@Param			request	body		presenter.AddFeedRequest	true	"Feed URL"
//	@Success		201		{object}	presenter.FeedResponse
//	@Failure		400		{string}	string	"invalid request body"
//	@Failure		500		{string}	string	"internal server error"
//	@Router			/feeds [post]
```

### Annotation order (always)
`@Summary` → `@Description` → `@Tags` → `@Accept` → `@Produce` → `@Param`(s) → `@Success` → `@Failure`(s, ascending status) → `@Router`.

### Rules per annotation
- **`@Summary`**: short imperative phrase, sentence case, no period ("List articles by date").
- **`@Description`**: one sentence stating observable behavior from the client's point of view, no trailing period. Describe the HTTP effect only: no use case names, no internal components, no technology names, no processes that don't exist (this app has no separate worker).
- **`@Tags`**: the resource, lowercase plural: `feeds`, `articles`. New resource → new tag; never invent per-endpoint tags.
- **`@Accept json`**: only when the handler reads a body. **`@Produce json`**: only when it writes a body. Omit both for bodiless responses (e.g. `202`).
- **`@Param`**: `name  in  type  required  "description"`. `in` is `body`, `query`, `path` or `header`.
  - body → type is the presenter request struct (`presenter.AddFeedRequest`).
  - query/path → Go primitive type; add `Format(...)` and `example(...)` when the value has a fixed pattern (`Format(date) example(2025-10-27)`); description states the pattern ("Date in YYYY-MM-DD format").
  - `required` must reflect what the handler enforces (a handler that returns 400 for a missing param ⇒ `true`).
- **`@Success`**: the exact status the handler sets (`render.Status` / `WriteHeader`; `200` is the default when unset). Body responses: `{object} presenter.<Response>` (use the presenter type, never the domain type or a model). Bodiless: status only (`@Success 202`).
- **`@Failure`**: one line for **every** `http.Error` status the handler can return, in ascending order. Errors are plain text, so use `{string} string "<message>"` with the message **copied verbatim** from the handler (`"missing q param"`, `"feed already exists"`). For passthrough errors (`err.Error()`) use the generic message `"internal server error"` with `500`, `"invalid request body"` for decode `400`s.
- **`@Router`**: path exactly as registered in the router (`cmd/api/main.go` route setup), `[method]` lowercase. Paths are relative to `@BasePath`. Several handlers must not share a `(path, method)`.
- Do not add annotations the project does not use (`@Security`, `@Header`, `@ID`, `@Deprecated`) unless the code needs them; when one is needed, add it here to the standard in the same change.

## Schemas (presenter)

- Every field has a `json:"snake_case"` tag.
- Add `example:"..."` to fields whose shape isn't obvious (IDs, URLs, names, timestamps); use `format:"date-time"` plus an RFC 3339 example for times.
- Field description goes in a trailing `// ...` comment on the same line (sentence fragment, no period): `Name string \`json:"name" example:"Some Feed"\` // Name of the feed`.
- Keep types in the grouped `type ( ... )` block and build responses through `New...` functions (project conventions), so the schema name is what `@Success` / `@Param` reference.

## API-level block (`cmd/api/main.go`)

Single block above `package main`. Update `@version` only on a deliberate API version change, and keep `@description` to one sentence about what the API offers. Never add per-endpoint info here.

## Checklist when adding or changing an endpoint

1. Read the handler: every status code, every `http.Error` message, body decode/encode, query/path params.
2. Write the block in the order above; messages copied verbatim; types are presenter types.
3. Confirm `@Router` matches the chi route registration.
4. Add `example`/`format` tags to any new presenter fields.
5. Run `gofmt -l internal/handler presenter` (it normalizes the tab alignment) and `go vet ./...`.
6. Regenerate with `make docs` (needs `swag` installed) and confirm the operation and its schemas appear in `docs/swagger.json`. If handlers under `internal/` are missing from the output, `swag` needs `--parseInternal`.
7. Handler tests still pass: `go test ./internal/handler/ ./e2e/`.
