# SOP: documenting the `internal/` packages

Applies to every non-test `.go` file under `internal/` (`database`, `database/model`, `feed`, `ml`, `queue`, `storage`). Loaded by the `document` skill.

These packages are adapters: they are where concrete technologies live. This SOP **overrides** the "never name the technology" common rule.

## Hard rules

1. **Do name the technology** the adapter wraps (SQLite/GORM, Chroma, Ollama, S3, RabbitMQ, Watermill, gofeed, ...). Say what it is and how the adapter talks to it.
2. **Never say where the adapter is used.** No mention of use cases, handlers, endpoints, `cmd/api`, which consumer calls it, or which feature depends on it. Do not write "used by Find", "called by the API", "wired in main", "not currently wired", etc. An adapter documents itself, not its callers.
   - Satisfying a port is fine to state generically ("implements the article search port" is borderline; prefer describing behavior: "finds article IDs by semantic similarity").
3. Comment-only changes; never touch `*_test.go`.

## Special case: `internal/handler/`

Handler doc comments are **generated API documentation**: the `// Name godoc` blocks carry `swag` annotations (`@Summary`, `@Param`, `@Success`, `@Router`, ...) that produce the Swagger docs (`make docs`).

- Keep the existing `// Name godoc` + annotation block format exactly; never reword, reorder, or delete annotation lines unless the endpoint's behavior changed.
- Document a handler by completing its annotations (`@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`) so they match the code's real status codes and request/response types.
- Annotations describe the HTTP contract, so route paths and methods belong here; the "no caller references" rule still applies to everything else (no mention of use cases or `cmd/api`).
- Plain Go comments (constructor, types) follow the normal rules below.
- Verify with `make docs` (requires `swag`) that generation still succeeds.

## What to document

- **Package comment** (`// Package x ...`): one per package, on one file only (the main file).
- **Every type** (inside the grouped `type ( ... )` block, comment on each spec), exported or not when it carries meaning (e.g. request/response DTOs: say which wire format they mirror).
- **Constructors**: what is built and what each parameter configures (endpoint, region, bucket, model, collection...).
- **Methods/functions**: behavior, return semantics (notably `(value, found bool, err)`), and the notable technology details a maintainer needs:
  - persistence: queries, eager-loaded relations, upsert (`Save`) vs insert (`Create`), auto-migration, date matching rules;
  - vector store: stored document, metadata fields, result count, similarity threshold and ordering assumptions;
  - HTTP clients: endpoint, request/response shape, status handling, sentinel errors;
  - object storage: key layout, path-style addressing, presign TTL;
  - messaging: topic/exchange/queue mapping, ack/nack/requeue translation, buffering, persistence, close semantics;
  - model mapping: which fields the `ToDomain` / `New...FromDomain` pair carries and which it drops or derives (IDs, timestamps, JSON-encoded columns, preloaded relations).
- **Sentinel errors and constants**: when the condition occurs; why a constant has its value.
- Fix or keep existing comments; do not leave stale ones.

## Style

- Go doc convention: starts with the identifier name, full sentences, third person, present tense.
- Read the function before describing it; state only what the code does (e.g. if a method ignores an argument, don't claim it uses it).
- Keep it short: one line for trivial methods, a few lines where there is non-obvious behavior.

## Verification

```bash
gofmt -l internal/
go vet ./internal/...
go test ./internal/...
# no caller/usage references in added comments:
git diff -U0 | grep '^+.*//' | grep -iE 'usecase|use case|handler|endpoint|cmd/api|used by|called by|wired|consumer of|main\.go'   # no output
```
