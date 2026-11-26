# RSS Feed

A Go service that fetches, stores, and serves news articles from RSS feeds. Uses SQLite for persistence, an in-process Watermill broker for async event processing, S3-compatible object storage for article images, and Chroma + Ollama for semantic search.

## Architecture

Clean/Hexagonal Architecture with clear layer separation:

```
pkg/rss/           → Domain models (Feed, Article, Image, ImageType) — no external deps
pkg/event/         → Event types, topics and jobs for async messaging (Feed, Article, Job)
pkg/hash/          → Hash helper used to derive IDs from URLs
pkg/markdown/      → HTML → Markdown parser (used to build the text that gets embedded)
usecase/           → Business logic with interface definitions
internal/handler  → HTTP request handlers (chi router)
presenter/         → Request/response DTOs
internal/database  → GORM + SQLite persistence (gorm.go) and Chroma vector store (chroma.go)
internal/feed      → RSS feed fetcher (gofeed) — emits event.Feed
internal/ml        → Ollama embedding client
internal/queue     → Watermill in-memory broker (watermill.go, used in production) and RabbitMQ client (rabbitmq.go, currently not wired)
internal/storage   → S3-compatible object storage (upload + presign)
cmd/api/           → Single binary: REST API server plus every background consumer, run as goroutines
e2e/               → End-to-end tests over the HTTP router with an in-memory broker
```

There is no separate worker process: `cmd/api/main.go` starts the HTTP server and launches `HandleFeed`, `ConsumeImage`, `Article` and `UpdateFeeds` as goroutines sharing one `queue.WatermillBroker` (Watermill `gochannel`, in-memory, `Persistent: true`). Messages do not survive a restart.

## Async Processing Flow

```
POST /feeds ─► SaveFeed.Execute()
                   └─► fetch → CreateFeed (SQLite) → publish "rss.feed.found"

UpdateFeeds.Execute(ctx, UPDATE_INTERVAL)   (runs once at startup, then every interval, default 30m)
    └─► for each stored feed:
            └─► fetcher.Fetch() → skip if not updated since last fetch
            └─► publish "rss.feed.found"
            └─► feed.Touch() → UpdateFeed (SQLite)
    └─► also consumes "rss.feed.jobs": POST /feeds/update publishes a job
        that forces an immediate refresh and resets the timer

HandleFeed.Execute(ctx)        (consumes "rss.feed.found")
    └─► for each article: SaveArticle (SQLite) → publish "rss.feed.article.found"

"rss.feed.article.found" has two consumers, each receiving every message:
    Article.Execute(ctx)       → CreateArticle in Chroma (Ollama embedding of the article's Markdown)
    ConsumeImage.Execute(ctx)  → if the article has an image:
                                     HTTP GET original image URL
                                     upload to S3 at images/original/{article_id}{ext}
                                     SaveImage (SQLite)
```

## Endpoints

| Method | Path                        | Description                                |
|--------|-----------------------------|--------------------------------------------|
| `GET`  | `/articles/today`           | List articles published today              |
| `GET`  | `/articles?date=YYYY-MM-DD` | List articles from a specific date         |
| `GET`  | `/articles/search?q=term`   | Semantic search over articles (Chroma)     |
| `POST` | `/feeds`                    | Add a new RSS feed source                  |
| `POST` | `/feeds/update`             | Trigger an immediate feed re-fetch         |
| `GET`  | `/swagger/*`                | Swagger UI                                 |

Responses include presigned S3 URLs (1h TTL) nested under `images[].url`, generated at read time by `ReadArticles`.

## Use Cases

- `SaveFeed` — checks for duplicates (`rss.ErrFeedAlreadyExists`), fetches the feed, persists it to SQLite, then publishes `rss.feed.found`
- `HandleFeed` — consumes `rss.feed.found`; saves each article to SQLite and publishes `rss.feed.article.found` per article
- `Article` — consumes `rss.feed.article.found`; indexes each article in Chroma for semantic search
- `ConsumeImage` — consumes `rss.feed.article.found`; downloads the article's image, uploads it to S3, saves the image record to SQLite
- `UpdateFeeds` — periodically (and on `rss.feed.jobs`) re-fetches all stored feeds, publishes `rss.feed.found`, then calls `feed.Touch()` and persists the updated entity
- `ReadArticles` — queries articles by date and presigns image URLs
- `Find` — searches Chroma for matching article IDs, then loads the articles from SQLite in ranking order

Shared ports live in `usecase/usecase.go`: `Publisher`, `Subscriber`, `Broker` (both) and `FeedFetcher`. Every other interface is declared next to the use case that consumes it.

## Database

SQLite file at the path given by `DB_PATH`. Auto-migrated on startup (`database.NewGorm`). IDs are SHA256 hashes of URLs (`pkg/hash`).

Three tables: `feeds`, `articles`, `images` (`images` has a FK to `articles`, stores `path`, `type`).

Chroma holds a separate `articles` collection with embeddings, used only by `Find`.

## Infrastructure

- **Message broker** — Watermill `gochannel` (in-process). Topics: `rss.feed.found`, `rss.feed.article.found`, `rss.feed.jobs` (`pkg/event/topic.go`). A RabbitMQ implementation exists in `internal/queue/rabbitmq.go` but is not wired in `cmd/api`
- **S3-compatible storage** — configured via `storage.NewS3(endpoint, region, bucket)`; images stored under `images/original/`
- **Presigned URLs** — generated on read with 1h TTL
- **Chroma** — vector store for semantic search; collection `articles`
- **Ollama** — embedding model provider

## Environment Variables

| Variable          | Purpose                                         |
|-------------------|-------------------------------------------------|
| `PORT`            | HTTP port (default `8080`)                      |
| `DB_PATH`         | SQLite file path                                |
| `S3_ENDPOINT`     | S3-compatible endpoint                          |
| `S3_REGION`       | S3 region                                       |
| `AWS_BUCKET`      | Bucket for images                               |
| `CHROMA_URL`      | Chroma base URL                                 |
| `OLLAMA_URL`      | Ollama base URL                                 |
| `OLLAMA_MODEL`    | Embedding model name                            |
| `UPDATE_INTERVAL` | Feed refresh interval, Go duration (default `30m`) |

## Key Dependencies

- `github.com/go-chi/chi/v5` — HTTP router
- `github.com/go-chi/cors` — CORS middleware (allows `http://localhost*`)
- `github.com/go-chi/render` — JSON rendering
- `github.com/mmcdole/gofeed` — RSS/Atom parser
- `github.com/ThreeDotsLabs/watermill` — in-memory pub/sub (`gochannel`)
- `github.com/rabbitmq/amqp091-go` — RabbitMQ client (not currently wired)
- `github.com/amikos-tech/chroma-go` — Chroma vector store client
- `github.com/JohannesKaufmann/html-to-markdown/v2` — HTML → Markdown conversion
- `github.com/aws/aws-sdk-go-v2` — S3 storage + presigning
- `github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager` — streaming uploads (v0.1.x, pre-stable)
- `github.com/rs/zerolog` — structured logging; the logger is carried in the context (`zerolog.Ctx(ctx)`)
- `github.com/swaggo/http-swagger` — Swagger UI
- `gorm.io/driver/sqlite` + `gorm.io/gorm` — ORM

## Common Commands

```bash
make docs          # regenerate Swagger docs (swag init) into ./docs — required before building, cmd/api imports rss-feed/docs
make start         # start the API server (and all consumers) via go run
go generate ./...  # same as make docs
go test ./...      # unit, integration (real SQLite) and e2e tests
```

`make build` and `make worker` still reference `./cmd/worker`, which no longer exists, so they fail.

## Type Declaration Convention

All struct and type declarations must use the grouped `type ( ... )` syntax, even for a single type:

```go
// correct
type (
    Foo struct {
        Bar string
    }
)

// incorrect
type Foo struct {
    Bar string
}
```

When multiple types belong to the same file, they should all be declared inside a single `type ( ... )` block at the top of the file.

## Goroutine and Defer Convention

When launching a goroutine or deferring a call to a function that returns an error, always wrap it in an anonymous function and explicitly discard the return value:

```go
// correct
go func() {
    _ = someUseCase.Execute(ctx)
}()

defer func() {
    _ = resource.Close()
}()

// incorrect — silently discards the return value without making it explicit
go someUseCase.Execute(ctx)
defer resource.Close()
```

This makes the intentional discard visible and avoids the implicit drop that Go allows but doesn't encourage.

## Design Notes

- Dependency injection via constructor functions
- `Publisher` interface in `usecase/usecase.go` is the shared abstraction for RabbitMQ publishing
- `Article.Images []Image` — an article can have multiple images (original, thumbnail types defined)
- Image paths are derived at upload time from article ID + file extension (not stored as a template)
- CORS allows all localhost origins regardless of port
- Do not put domain logic outside `pkg` entities — mutations like `Touch()` belong on the domain struct itself, not in the repository or use case layer

## Testing Convention

- Tests that interact with the database must use a real SQLite instance (via `newTestDB` in `usecase/integration_test.go`) — never mock the database
- Every new method must have a corresponding test; write the test alongside the implementation (TDD-like)
- Reuse the shared test helpers in `usecase/integration_test.go` when setting up common fixtures:
  - `createTestFeed(t, ctx, db)` — inserts a default feed (`feed-abc`, "Test Blog")
  - `saveTestArticle(t, ctx, db, publishedAt)` — inserts a default article (`article-1`, "First Post") under `feed-abc`

## Domain Binding Convention

Structs that are populated from a domain entity (types living in `pkg/`) must be constructed via a `New...` function:

```go
// correct
func NewFeedResponse(feed rss.Feed) FeedResponse { ... }
response := presenter.NewFeedResponse(feed)

// incorrect — do not fill presenter/model structs inline
response := presenter.FeedResponse{ID: feed.ID, Name: feed.Name, URL: feed.URL}
```

Structs that need to be converted back into a domain entity must expose a `ToDomain()` method:

```go
func (f Feed) ToDomain() rss.Feed { ... }
domain := model.ToDomain()
```

This applies to all layers that bridge to the domain: `presenter/`, `internal/database/model/`, and `pkg/event/`.
