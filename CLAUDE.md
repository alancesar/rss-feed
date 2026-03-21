# RSS Summary

A Go API service that aggregates news articles from RSS feeds, backed by SQLite.

## Architecture

Clean/Hexagonal Architecture with clear layer separation:

```
pkg/rss/          → Domain models (Feed, Article, DateTime) — no external deps
usecase/          → Business logic with interface definitions
handler/          → HTTP request handlers (chi router)
presenter/        → Request/response DTOs
internal/database → GORM + SQLite persistence
internal/feed     → RSS feed fetcher (gofeed)
cmd/api/          → REST API server entrypoint (port 3000)
cmd/update/       → CLI tool to sync feeds from DB
```

## Endpoints

| Method | Path                        | Description                        |
|--------|-----------------------------|------------------------------------|
| `GET`  | `/articles/today`           | List articles published today      |
| `GET`  | `/articles?date=YYYY-MM-DD` | List articles from a specific date |
| `POST` | `/feeds`                    | Add a new RSS feed source          |

## CLI

`cmd/update` fetches and syncs all feeds stored in the database. Intended to be run as a cron job.

## Database

SQLite file at `rss.sqlite` (hardcoded). Auto-migrated on startup. IDs are SHA256 hashes of URLs for deduplication.

Two tables: `feeds` and `articles` (with `published_at` index for date queries).

## Key Dependencies

- `github.com/go-chi/chi/v5` — HTTP router
- `github.com/mmcdole/gofeed` — RSS/Atom parser
- `gorm.io/driver/sqlite` + `gorm.io/gorm` — ORM

## Use Cases

- `AddFeed` — fetches a feed by URL and persists it with articles
- `UpdateFeeds` — iterates all stored feed URLs and re-fetches them
- `ReadArticle` — queries articles by date

## Design Notes

- Dependency injection via constructor functions (`NewAddFeed`, `NewRead`, etc.)
- Interfaces defined in `usecase/usecase.go` (`FeedFetcher`, `FeedStore`, `ArticlesStore`)
- No environment config or external config files — ports and DB path are hardcoded
