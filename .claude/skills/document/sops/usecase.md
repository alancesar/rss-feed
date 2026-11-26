# SOP: documenting the `usecase/` package

Applies to files under `usecase/`. Loaded by the `document` skill.

Comments are documentation for the *behavior and intent* of the business logic. They must stay valid if the underlying infrastructure is swapped.

## Hard rule: no technology names

Never name the tool, framework, or protocol behind an interface. Describe what it does.

| Don't write | Write |
|---|---|
| publish to Watermill / RabbitMQ | publish to the broker / announce the event |
| create the embeddings using Chroma/Ollama | vectorize the article text and store it in the search index |
| upload to S3 | store the file |
| save to SQLite / GORM | persist / record |
| presigned S3 URL | temporary access URL |

Neutral vocabulary: broker, topic, delivery, store, repository, search index, file storage, vectorize, semantic search, acknowledge, reject (with/without redelivery).

Also avoid logging libraries and library types in comments. Existing log strings are out of scope unless asked.

## Scope

- Files: every non-test `.go` file in `usecase/`. Never edit `*_test.go` for documentation.
- Comment-only changes. Do not alter logic, signatures, or log messages.

## What to document

1. **Package** — one `// Package usecase ...` comment, only in `usecase.go`.
2. **Every type, inside the grouped `type ( ... )` block** (project convention: never a standalone `type X`). Put the comment directly above each spec.
   - Port interfaces: the capability the use case needs, not who implements it.
   - Interface methods: one comment each; state non-obvious returns (e.g. the `bool` means "found", ordering of results).
   - Use case structs: one sentence on its role (consumer, command, query).
3. **Constructors (`New...`)**: `NewX creates an X use case that <role of each dependency>.`
4. **`Execute`**: the contract, in this order when applicable:
   - what it does (flow);
   - delivery handling for consumers: bad payload → rejected **without** redelivery; processing failure → rejected **for** redelivery; empty/irrelevant → acknowledged;
   - failure policy (aborts vs. logs and continues vs. skips an item);
   - termination: returns `ctx.Err()` on cancel, `nil` when the subscription closes, and what it closes on exit;
   - side effects worth knowing (e.g. mutates returned slice, fixed TTL).
5. **Unexported helpers** with real logic (e.g. `saveImageToStorage`, `updateFeeds`): document the steps and error conditions.

## Style

- Go doc convention: the comment starts with the identifier name (`// Execute consumes ...`).
- English, third person, present tense, full sentences ending in a period.
- Quote event names in plain words: `"feed found"`, `"article found"` — not topic constants or dotted topic strings.
- State facts verifiable in the code; read the function before describing it. Do not document intent the code does not implement.
- Keep it short: 1–4 lines for `Execute`, 1 line for most others.

## Procedure

1. Read the file fully, including its interfaces and every branch of `Execute` (ack/nack paths, early returns, `defer`).
2. Write comments following the sections above.
3. Add a comment for any newly added exported identifier in the same change that introduces it.

## Verification checklist

```bash
gofmt -l usecase/                      # no output
go vet ./usecase/
go test ./usecase/
git diff -U0 | grep '^+.*//' | grep -iE 'chroma|ollama|s3|sqlite|watermill|rabbit|gorm|zerolog|amqp'   # no output
go doc ./usecase                       # every exported symbol has a description
```

If the grep prints anything, rewrite that comment in neutral terms.
