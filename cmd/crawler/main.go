package main

import (
	"context"
	"log"
	"net/http/cookiejar"
	"os"
	"strconv"
	"time"
	"ufc_stats_api/internal/config"
	"ufc_stats_api/internal/crawler"

	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// challengeProbeURL is any ufcstats page; solving its proof-of-work yields a
// cookie valid for both the apex and www hosts.
const challengeProbeURL = "http://ufcstats.com/statistics/events/completed?page=all"

// newCollector builds a fresh async collector. Each crawl phase needs its own so
// their overlapping OnHTML selectors don't fire across phases.
func newCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*ufcstats*",
		Parallelism: 5,
		RandomDelay: 500 * time.Millisecond,
	})

	// Solve the anti-bot proof-of-work once and share the resulting cookie jar
	// with colly, so every crawl request is waved through automatically.
	jar, _ := cookiejar.New(nil)
	if err := crawler.Bootstrap(jar, challengeProbeURL); err != nil {
		log.Printf("challenge bootstrap failed: %v", err)
	}
	c.SetCookieJar(jar)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})
	return c
}

func main() {
	godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	// Mode:
	//   go run ./cmd/crawler fighters       -> crawl all fighters
	//   go run ./cmd/crawler fights          -> crawl all events + fights + stats
	//   go run ./cmd/crawler fights 5        -> crawl only the 5 most recent events
	//   go run ./cmd/crawler all             -> fighters, then fights (full refresh)
	//   go run ./cmd/crawler all 5           -> fighters, then 5 most recent events
	mode := "fights"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	maxEvents := 0
	if len(os.Args) > 2 {
		maxEvents, _ = strconv.Atoi(os.Args[2])
	}
	switch mode {
	case "fighters":
		crawler.FighterCrawler(newCollector(), pool)
	case "fights":
		crawler.FightCrawler(newCollector(), pool, maxEvents)
	case "all":
		// Fighters first: fights look up fighter IDs and skip any that are missing.
		crawler.FighterCrawler(newCollector(), pool)
		crawler.FightCrawler(newCollector(), pool, maxEvents)
	default:
		log.Fatalf("unknown mode %q (use: fighters | fights [N] | all [N])", mode)
	}
}
