# UFC Stats API

A Go-based scraper and REST API for historical UFC fight, fighter, and event data. A [Colly](https://github.com/gocolly/colly)-based crawler pulls raw data from ufcstats.com and normalizes it into PostgreSQL, and a separate HTTP service exposes it for querying.

The database currently holds **~8,800 fights across 784 events**, fully linked to fighter records.

## How it works

The project is split into two binaries that share a common data layer:

- **`cmd/crawler`** — crawls UFC event, fight, and fighter pages with Colly, parses the HTML, and upserts normalized records into PostgreSQL.
- **`cmd/api`** — a lightweight HTTP server (stdlib `net/http`) that reads from the same database and serves the data as JSON.

Both connect to Postgres through [pgx](https://github.com/jackc/pgx) using connection pooling via `pgxpool`. All writes are upserts (`ON CONFLICT (url) DO UPDATE`), so the crawler is safe to re-run without creating duplicates.

## Data model

Data is normalized across four entities:

- **Fighters** — name, nickname, physical stats (height, reach, stance, DOB), and career striking/grappling averages (SLpM, strike accuracy/defense, takedown average/accuracy/defense, submission average)
- **Events** — event name, date, location
- **Fights** — links two fighters to an event, with winner, method, round, time, referee, and title-fight flag
- **Fight Stats** *(planned)* — per-round breakdown of strikes and takedowns by target (head/body/leg) and position (distance/clinch/ground)

Fights are linked to fighters by resolving each fighter's page URL to an internal ID (matched on the URL's trailing hash, so it survives `www`/scheme differences between pages).

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/fighters` | All fighters in the database |
| `GET` | `/fighters/search?name=<name>` | Fighters whose name matches (case-insensitive, partial); returns a JSON array |
| `GET` | `/events` | All events |
| `GET` | `/events/{id}/fights` | All fights on a given event |
| `GET` | `/fighters/{id}/fights` | A fighter's complete fight history (either corner) |

Responses are JSON. List endpoints return `[]` (not `null`) when there are no matches. A non-numeric `{id}` returns `400`; a missing `name` parameter returns `400`; database errors return `500`.

### Example

```bash
curl http://localhost:8080/fighters/17951/fights   # Jim Miller's 47 fights
curl "http://localhost:8080/fighters/search?name=silva"
curl http://localhost:8080/events/1135/fights
```

## Getting started

### Prerequisites

- Go 1.26+
- A PostgreSQL database (the project runs against [Supabase](https://supabase.com))

### Setup

1. Clone the repo and copy the env template:
   ```bash
   git clone https://github.com/bilalmaq01/ufc_stats_api.git
   cd ufc_stats_api
   cp .env.example .env
   ```
2. Configure `.env`:
   - **`DATABASE_URL`** — your Postgres connection string (required).
   - **`PORT`** — API server port (optional, defaults to `8080`).
   - **`COOKIE`** — optional cookie header sent by the crawler.

   > **Supabase note:** use the **connection pooler** string (`...pooler.supabase.com`, Session mode), not the direct `db.<ref>.supabase.co` host — the direct host is IPv6-only and unreachable on most networks.

3. Run the crawler to populate the database:
   ```bash
   go run ./cmd/crawler
   ```
4. Start the API server:
   ```bash
   go run ./cmd/api
   ```

## Project layout

```
cmd/
  api/        HTTP server entrypoint
  crawler/    scraper entrypoint
internal/
  config/     env-based configuration
  crawler/    Colly crawlers (fighters, events, fights)
  handlers/   HTTP handlers
  models/     domain structs
  storage/    pgx data-access layer
```

## Tech stack

Go · PostgreSQL · [pgx](https://github.com/jackc/pgx) · [Colly](https://github.com/gocolly/colly) · [godotenv](https://github.com/joho/godotenv)

## Roadmap

- [x] Fighter crawler + endpoints
- [x] Event crawler + endpoints
- [x] Fight crawler (full historical backfill) + endpoints
- [x] Fight stats (per-round breakdown)
- [x] Head-to-head fight search (e.g. "Frank Mir vs Tim Sylvia")
- [x] Incremental weekly crawler (only new events + affected fighters)
- [x] Deploy to AWS
- [x] Yippie!
