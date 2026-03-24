# RSS Feed

Aggregates articles from RSS feeds, stores them in SQLite, and serves them through a REST API. Article images are downloaded asynchronously and re-hosted from S3-compatible object storage.

## Requirements

- Go 1.21+
- RabbitMQ
- S3-compatible object storage (AWS S3, MinIO, RustFS, etc.)

## Configuration

All configuration is done via environment variables. Copy `.env.example` to `.env` and adjust:

| Variable          | Description                    | Example                                  |
|-------------------|--------------------------------|------------------------------------------|
| `DB_PATH`         | SQLite file path               | `rss.sqlite`                             |
| `S3_ENDPOINT`     | S3-compatible endpoint URL     | `https://s3.example.com`                 |
| `S3_REGION`       | S3 region                      | `us-east-1`                              |
| `AWS_BUCKET`      | S3 bucket name                 | `rss-feed`                               |
| `AMQP_URL`        | RabbitMQ connection URL (TLS)  | `amqps://user:pass@amqp.example.com`     |
| `PORT`            | HTTP port for the API server   | `8080` (default)                         |
| `UPDATE_INTERVAL` | Feed refresh interval (worker) | `30m` (default), accepts any Go duration |

## Running

The project has two binaries. You'll typically want both running at the same time.

**API server** — serves the REST API:
```bash
make start
```

**Worker** — processes RabbitMQ events and periodically re-fetches all registered feeds:
```bash
make worker
```

To build all binaries to `./bin/`:
```bash
make build
```

## API

### Add a feed

```
POST /feeds
Content-Type: application/json

{ "url": "https://example.com/rss" }
```

Response `201 Created`:
```json
{
  "id": "abc123",
  "name": "Example Blog",
  "url": "https://example.com/rss"
}
```

Adding a feed registers it and immediately triggers an async fetch. Articles will appear once the worker processes the event.

### List articles

```
GET /articles?date=2025-10-27
GET /articles/today
```

Response `200 OK`:
```json
{
  "articles": [
    {
      "title": "Article title",
      "url": "https://example.com/article",
      "published_at": "2025-10-27T14:00:00-03:00",
      "feed": {
        "id": "abc123",
        "name": "Example Blog",
        "url": "https://example.com/rss"
      },
      "images": [
        {
          "url": "https://s3.example.com/rss-feed/images/original/abc123.jpg",
          "type": "original"
        }
      ]
    }
  ]
}
```

The `date` parameter must follow the `YYYY-MM-DD` format. `published_at` is an RFC3339 timestamp.

### Swagger UI

Interactive docs available at `/swagger/index.html` once the server is running. To regenerate after changing annotations:

```bash
make docs
```

## How it works

When a feed is added, the API publishes an event to RabbitMQ. The worker picks it up, saves the articles to SQLite, and publishes a separate event for each article image. A second worker consumer downloads the image, uploads it to S3, and records its path in the database. Image URLs in API responses point directly to S3.

The worker also runs a background ticker (default every 30 minutes, overridable via `UPDATE_INTERVAL`) that re-triggers this pipeline for every feed already in the database.

## RabbitMQ setup

The service expects a topic exchange named `rss` with three queues bound to it:

| Queue                            | Routing key                    |
|----------------------------------|--------------------------------|
| `rss.feed.article.found`         | `feed.article.found`           |
| `rss.feed.article.image.found`   | `feed.article.image.found`     |
| `rss.feed.jobs`                  | `feed.jobs`                    |
