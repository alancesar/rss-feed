# RSS Feed

A Go service that fetches, stores, and serves news articles from RSS feeds. Uses SQLite for persistence, RabbitMQ for async event processing, and S3-compatible object storage for article images.

## Architecture

Clean/Hexagonal Architecture with clear layer separation:

```
pkg/rss/          → Domain models (Feed, Article, Image, ImageType) — no external deps
pkg/event/        → Event types for async messaging (Feed, Article, Image)
usecase/          → Business logic with interface definitions
handler/          → HTTP request handlers (chi router)
presenter/        → Request/response DTOs
internal/database → GORM + SQLite persistence
internal/feed     → RSS feed fetcher (gofeed) — emits event.Feed
internal/queue    → RabbitMQ publisher and consumer
internal/storage  → S3-compatible object storage (upload + presign)
cmd/api/          → REST API server (default port 8080, overridable via PORT env)
cmd/update/       → CLI tool: reads all feed URLs from DB and triggers re-fetch
cmd/worker/       → Long-running worker: consumes RabbitMQ events and processes feeds/images
```

## Async Processing Flow

```
cmd/update
    └─► UpdateFeeds.Execute()
            └─► for each url:
                    └─► fetcher.Fetch() → event.Feed
                    └─► RabbitMQ publish → "rss.feed.article.found"

cmd/worker (consumer 1: "rss.feed.article.found")
    └─► ConsumeFeed.Execute(event.Feed)
            └─► for each article: SaveArticle to SQLite
            └─► for each article with image:
                    └─► RabbitMQ publish → "rss.feed.article.image.found"

cmd/worker (consumer 2: "rss.feed.article.image.found")
    └─► ConsumeImage.Execute(event.Image)
            └─► HTTP GET original image URL
            └─► Upload to S3 at images/original/{article_id}.{ext}
            └─► SaveImage to SQLite
```

## Endpoints

| Method | Path                        | Description                        |
|--------|-----------------------------|------------------------------------|
| `GET`  | `/articles/today`           | List articles published today      |
| `GET`  | `/articles?date=YYYY-MM-DD` | List articles from a specific date |
| `POST` | `/feeds`                    | Add a new RSS feed source          |

Responses include `image_url` as a presigned S3 URL (1h TTL), generated at read time by `ReadArticles`.

## Use Cases

- `SaveFeed` — checks for duplicates, persists feed to DB, then publishes a `feed.article.found` event to RabbitMQ
- `ConsumeFeed` — saves each article to DB, publishes `feed.article.image.found` per article with image
- `ConsumeImage` — downloads image, uploads to S3, saves image record to DB
- `UpdateFeeds` — retrieves all stored feeds, fetches each, publishes `feed.article.found` events, then calls `feed.Touch()` and persists the updated domain entity back to DB
- `ReadArticles` — queries articles by date and presigns image URLs

## Database

SQLite file at `rss.sqlite`. Auto-migrated on startup. IDs are SHA256 hashes of URLs.

Three tables: `feeds`, `articles`, `images` (`images` has a FK to `articles`, stores `path`, `type`).

## Infrastructure

- **RabbitMQ** — exchange `rss`, queues `rss.feed.article.found` and `rss.feed.article.image.found`
- **S3-compatible storage** — configured via `storage.NewS3(endpoint, region, bucket)`; images stored under `images/original/`
- **Presigned URLs** — generated on read with 1h TTL

## Key Dependencies

- `github.com/go-chi/chi/v5` — HTTP router
- `github.com/go-chi/cors` — CORS middleware (allows `http://localhost*`)
- `github.com/mmcdole/gofeed` — RSS/Atom parser
- `github.com/rabbitmq/amqp091-go` — RabbitMQ client
- `github.com/aws/aws-sdk-go-v2` — S3 storage + presigning
- `github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager` — streaming uploads (v0.1.x, pre-stable)
- `gorm.io/driver/sqlite` + `gorm.io/gorm` — ORM

## Common Commands

```bash
make docs          # regenerate Swagger docs (swag init)
make build         # compile binaries to ./bin/api, ./bin/update, ./bin/worker
make start         # start the API server via go run
make update        # run the feed sync CLI via go run
go generate ./...  # same as make docs — regenerate Swagger docs via go:generate directive
```

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
