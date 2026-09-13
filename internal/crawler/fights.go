package crawler

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"ufc_stats_api/internal/models"
	"ufc_stats_api/internal/storage"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func FightCrawler(c *colly.Collector, pool *pgxpool.Pool, maxEvents int) {
	eventsSeen := 0
	// --- events list page: insert the event, then visit its page ---
	c.OnHTML("tr.b-statistics__table-row a.b-link", func(e *colly.HTMLElement) {
		if maxEvents > 0 && eventsSeen >= maxEvents {
			return
		}
		eventsSeen++
		var event models.Event
		event.EventName = strings.TrimSpace(e.Text)
		eventDate := strings.TrimSpace(e.DOM.Parent().Find("span.b-statistics__date").Text())
		date, err := time.Parse("January 02, 2006", eventDate)
		if err != nil {
			log.Println("error parsing date from", e.Text)
			return
		}
		event.Date = date
		event.Location = strings.TrimSpace(e.DOM.Closest("tr").Find("td").Eq(1).Text())
		event.URL = e.Attr("href")
		if err := storage.InsertEvent(pool, &event); err != nil {
			log.Println(err)
			return
		}
		log.Println("event:", event.EventName)
		c.Visit(event.URL)
	})

	// --- event page: each fight row carries its detail URL on data-link ---
	c.OnHTML("tr.js-fight-details-click", func(f *colly.HTMLElement) {
		if link := f.Attr("data-link"); link != "" {
			c.Visit(link)
		}
	})

	// --- fight-details page: full fight record + per-round stats ---
	c.OnHTML("html", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Request.URL.Path, "/fight-details/") {
			return
		}
		fight, ok := parseFightRecord(e, pool)
		if !ok {
			return
		}
		if err := storage.InsertFight(pool, &fight); err != nil {
			log.Printf("insert fight %s: %v", fight.URL, err)
			return
		}
		// fight.ID is populated by InsertFight (RETURNING id).
		stats := parseFightStats(e, pool, fight.ID)
		for i := range stats {
			if err := storage.InsertFightStats(pool, &stats[i]); err != nil {
				log.Printf("insert stats fight %d fighter %d round %d: %v",
					stats[i].FightID, stats[i].FighterID, stats[i].RoundNumber, err)
			}
		}
		log.Printf("fight %s -> %d stat rows", fight.URL, len(stats))
	})

	c.Visit("http://ufcstats.com/statistics/events/completed?page=all")
	fmt.Println("crawl complete")
}

// parseFightRecord builds a complete Fight from the fight-details page. Returns
func parseFightRecord(e *colly.HTMLElement, pool *pgxpool.Pool) (models.Fight, bool) {
	var fight models.Fight
	fight.URL = e.Request.URL.String()

	eventID, err := storage.GetEventIDByURL(pool, e.ChildAttr("h2.b-content__title a", "href"))
	if err != nil {
		log.Printf("fight %s: event lookup failed: %v", fight.URL, err)
		return fight, false
	}
	fight.EventID = eventID

	persons := e.DOM.Find(".b-fight-details__person")
	if persons.Length() < 2 {
		return fight, false
	}
	href1, _ := persons.Eq(0).Find("a.b-fight-details__person-link").Attr("href")
	href2, _ := persons.Eq(1).Find("a.b-fight-details__person-link").Attr("href")
	fight.Fighter1ID, err = storage.GetFighterIDByURL(pool, href1)
	if err != nil {
		log.Printf("fight %s: fighter1 lookup failed: %v", fight.URL, err)
		return fight, false
	}
	fight.Fighter2ID, err = storage.GetFighterIDByURL(pool, href2)
	if err != nil {
		log.Printf("fight %s: fighter2 lookup failed: %v", fight.URL, err)
		return fight, false
	}

	// Winner is flagged by a "W" status; draws/no-contests leave winner_id null.
	switch "W" {
	case strings.TrimSpace(persons.Eq(0).Find(".b-fight-details__person-status").Text()):
		fight.WinnerID = &fight.Fighter1ID
	case strings.TrimSpace(persons.Eq(1).Find(".b-fight-details__person-status").Text()):
		fight.WinnerID = &fight.Fighter2ID
	}

	titleText := strings.TrimSpace(e.DOM.Find("i.b-fight-details__fight-title").Text())
	fight.IsTitle = strings.Contains(titleText, "Title")
	fight.WeightClass = cleanWeightClass(titleText)

	// Method / Round / Time / Time format / Referee / Details each sit in an item
	// whose label <i> we strip to get the value.
	e.DOM.Find("i.b-fight-details__label").Each(func(_ int, lab *goquery.Selection) {
		label := strings.TrimSpace(lab.Text())
		value := strings.TrimSpace(strings.TrimPrefix(collapse(lab.Parent().Text()), label))
		switch label {
		case "Method:":
			fight.Method = value
		case "Round:":
			fight.Round, _ = strconv.Atoi(value)
		case "Time:":
			fight.Time = value
		case "Time format:":
			fight.TimeFormat = value
		case "Referee:":
			fight.Referee = value
		case "Details:":
			if value != "" && value != "--" {
				fight.Details = &value
			}
		}
	})

	return fight, true
}

func parseFightStats(e *colly.HTMLElement, pool *pgxpool.Pool, fightID int) []models.FightStats {
	tables := e.DOM.Find("table.js-fight-table")
	if tables.Length() < 2 {
		return nil
	}

	type key struct{ round, idx int }
	byKey := map[key]*models.FightStats{}
	var order []key
	fighterID := map[string]int{} // cache href -> id (only two fighters)

	resolve := func(href string) (int, bool) {
		if id, ok := fighterID[href]; ok {
			return id, true
		}
		id, err := storage.GetFighterIDByURL(pool, href)
		if err != nil {
			log.Printf("fight %d: stat fighter lookup failed for %s: %v", fightID, href, err)
			return 0, false
		}
		fighterID[href] = id
		return id, true
	}

	// Totals table: 0 Fighter, 1 KD, 2 Sig.str, 4 Total str, 5 Td, 7 Sub.att, 8 Rev, 9 Ctrl.
	eachStatRow(tables.Eq(0), func(round int, row *goquery.Selection) {
		tds := row.Find("td")
		if tds.Length() < 10 {
			return
		}
		links := tds.Eq(0).Find("a")
		for i := 0; i < 2; i++ {
			href, _ := links.Eq(i).Attr("href")
			id, ok := resolve(href)
			if !ok {
				continue
			}
			var st models.FightStats
			st.FightID = fightID
			st.FighterID = id
			st.RoundNumber = round
			st.Knockdowns = atoi(pText(tds.Eq(1), i))
			st.SigStrikesLanded, st.SigStrikesAttempted = parseXofY(pText(tds.Eq(2), i))
			st.TotalStrikesLanded, st.TotalStrikesAttempted = parseXofY(pText(tds.Eq(4), i))
			st.TotalTakedownsLanded, st.TotalTakedownsAttempted = parseXofY(pText(tds.Eq(5), i))
			st.SubAttempts = atoi(pText(tds.Eq(7), i))
			st.Reversals = atoi(pText(tds.Eq(8), i))
			st.ControlTime = parseControlTime(pText(tds.Eq(9), i))
			k := key{round, i}
			byKey[k] = &st
			order = append(order, k)
		}
	})

	// Significant Strikes table: 3 Head, 4 Body, 5 Leg, 6 Distance, 7 Clinch, 8 Ground.
	eachStatRow(tables.Eq(1), func(round int, row *goquery.Selection) {
		tds := row.Find("td")
		if tds.Length() < 9 {
			return
		}
		for i := 0; i < 2; i++ {
			st := byKey[key{round, i}]
			if st == nil {
				continue
			}
			st.SignificantStrikesHeadLanded, st.SignificantStrikesHeadAttempted = parseXofY(pText(tds.Eq(3), i))
			st.SignificantStrikesBodyLanded, st.SignificantStrikesBodyAttempted = parseXofY(pText(tds.Eq(4), i))
			st.SignificantStrikesLegLanded, st.SignificantStrikesLegAttempted = parseXofY(pText(tds.Eq(5), i))
			st.SignificantStrikesDistanceLanded, st.SignificantStrikesDistanceAttempted = parseXofY(pText(tds.Eq(6), i))
			st.SignificantStrikesClinchLanded, st.SignificantStrikesClinchAttempted = parseXofY(pText(tds.Eq(7), i))
			st.SignificantStrikesGroundLanded, st.SignificantStrikesGroundAttempted = parseXofY(pText(tds.Eq(8), i))
		}
	})

	out := make([]models.FightStats, 0, len(order))
	for _, k := range order {
		out = append(out, *byKey[k])
	}
	return out
}

// eachStatRow walks a per-round table's rows, tracking the current round from the
// "Round N" divider rows and calling fn for each data row (rows that have <td>s).
func eachStatRow(table *goquery.Selection, fn func(round int, row *goquery.Selection)) {
	round := 0
	table.Find("tr").Each(func(_ int, row *goquery.Selection) {
		if th := row.Find("th"); th.Length() > 0 {
			fields := strings.Fields(th.First().Text())
			if len(fields) == 2 && fields[0] == "Round" {
				round, _ = strconv.Atoi(fields[1])
			}
			return
		}
		if round > 0 {
			fn(round, row)
		}
	})
}

// pText returns the i-th <p> text inside a stat cell (p0 = fighter1, p1 = fighter2).
func pText(td *goquery.Selection, i int) string {
	return strings.TrimSpace(td.Find("p").Eq(i).Text())
}

// parseXofY splits a "landed of attempted" cell (e.g. "15 of 39") into two ints.
func parseXofY(s string) (int, int) {
	parts := strings.Split(s, " of ")
	if len(parts) != 2 {
		return 0, 0
	}
	return atoi(parts[0]), atoi(parts[1])
}

// parseControlTime turns "M:SS" control time into total seconds ("--" -> 0).
func parseControlTime(s string) int {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return 0
	}
	return atoi(parts[0])*60 + atoi(parts[1])
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// collapse squashes runs of whitespace (incl. newlines) into single spaces.
func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// cleanWeightClass pulls the weight class out of a fight-title tag such as
// "Welterweight Bout" or "UFC Interim Welterweight Title Bout".
func cleanWeightClass(title string) string {
	wc := collapse(title)
	for _, suf := range []string{"Bout", "Title", "Tournament", "Interim"} {
		wc = strings.TrimSpace(strings.TrimSuffix(wc, suf))
	}
	for _, pre := range []string{"UFC", "Interim", "Ultimate Fighter", "TUF", "Road to UFC"} {
		wc = strings.TrimSpace(strings.TrimPrefix(wc, pre))
	}
	return wc
}
