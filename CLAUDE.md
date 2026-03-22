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
cmd/api/          → REST API server (port 3000)
cmd/update/       → CLI tool: reads all feed URLs from DB and triggers re-fetch
cmd/worker/       → Long-running worker: consumes RabbitMQ events and processes feeds/images
```

## Async Processing Flow

```
cmd/update
    └─► PublishFeed.Execute(url)
            └─► fetcher.Fetch() → event.Feed
            └─► RabbitMQ publish → "rss.feed.found"

cmd/worker (consumer 1: "rss.feed.found")
    └─► ConsumeFeed.Execute(event.Feed)
            └─► SaveFeed to SQLite
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

- `PublishFeed` — fetches RSS feed and publishes a `feed.found` event to RabbitMQ
- `ConsumeFeed` — persists feed+articles to DB, publishes `feed.article.image.found` per image
- `ConsumeImage` — downloads image, uploads to S3, saves image record to DB
- `UpdateFeeds` — iterates all stored feed URLs and calls `PublishFeed.Execute` for each
- `ReadArticles` — queries articles by date and presigns image URLs

## Database

SQLite file at `rss.sqlite`. Auto-migrated on startup. IDs are SHA256 hashes of URLs.

Three tables: `feeds`, `articles`, `images` (`images` has a FK to `articles`, stores `path`, `type`).

## Infrastructure

- **RabbitMQ** — exchange `rss`, queues `rss.feed.found` and `rss.feed.article.image.found`
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
make run           # run the API server via go run
make update        # run the feed sync CLI via go run
go generate ./...  # same as make docs — regenerate Swagger docs via go:generate directive
```

## Design Notes

- Dependency injection via constructor functions
- `Publisher` interface in `usecase/usecase.go` is the shared abstraction for RabbitMQ publishing
- `Article.Images []Image` — an article can have multiple images (original, thumbnail types defined)
- Image paths are derived at upload time from article ID + file extension (not stored as a template)
- CORS allows all localhost origins regardless of port
