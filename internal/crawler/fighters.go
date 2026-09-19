package crawler

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"ufc_stats_api/internal/models"
	"ufc_stats_api/internal/storage"

	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func FighterCrawler(c *colly.Collector, pool *pgxpool.Pool) {

	c.OnHTML("tr.b-statistics__table-row td:first-child a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link != "" {
			c.Visit(link)
		}
	})

	c.OnHTML("h2.b-content__title", func(e *colly.HTMLElement) {
		fighter := parseFighterDetail(e)
		fighter.URL = e.Request.URL.String()
		if err := storage.InsertFighter(fighter, pool); err != nil {
			log.Printf("failed to insert fighter %s (%s): %v", fighter.Name, fighter.URL, err)
		}
	})

	for _, letter := range "abcdefghijklmnopqrstuvwxyz" {
		link := fmt.Sprintf("http://www.ufcstats.com/statistics/fighters?char=%c&page=all", letter)
		c.Visit(link)
		link = fmt.Sprintf("Queued %s", link)
		fmt.Println(link)
	}
	c.Wait()
}
func parseFighterDetail(e *colly.HTMLElement) models.Fighter {
	var fighter models.Fighter

	name := strings.TrimSpace(e.ChildText("span.b-content__title-highlight"))
	fighter.Name = name
	record := strings.TrimSpace(e.ChildText("span.b-content__title-record"))
	record = strings.TrimPrefix(record, "Record: ")
	parts := strings.Split(record, "-")
	if len(parts) == 3 {
		fighter.Wins, _ = strconv.Atoi(parts[0])
		fighter.Losses, _ = strconv.Atoi(parts[1])
		fighter.Draws, _ = strconv.Atoi(parts[2])
	}
	nickname := strings.TrimSpace(e.DOM.Parent().Find("p.b-content__Nickname").Text())
	if nickname != "" {
		fighter.Nickname = &nickname
	}
	statsBox := e.DOM.Parent().Find("div.b-list__info-box.b-list__info-box_style_small-width")

	height := strings.TrimSpace(statsBox.Find("li:nth-child(1)").Text())
	fields := strings.Fields(height)
	if len(fields) > 1 {
		heightVal := strings.TrimSpace(strings.Join(fields[1:], " "))
		if heightVal != "--" {
			fighter.Height = &heightVal
		}
	}
	weight := strings.TrimSpace(statsBox.Find("li:nth-child(2)").Text())
	fields = strings.Fields(weight)
	if len(fields) >= 1 {
		weightVal := fields[1]
		if weightVal != "" && weightVal != "--" {
			fighter.WeightClass = &weightVal
		}
	}
	reach := strings.TrimSpace(statsBox.Find("li:nth-child(3)").Text())
	fields = strings.Fields(reach)
	if len(fields) > 1 {
		reachVal := fields[len(fields)-1]
		reachVal = strings.TrimSuffix(reachVal, "\"")
		fighter.ReachIn, _ = strconv.Atoi(reachVal)
	}

	stance := statsBox.Find("li:nth-child(4)").Text()
	fields = strings.Fields(stance)
	if len(fields) > 1 {
		stanceVal := fields[len(fields)-1]
		fighter.Stance = &stanceVal
	}

	dob := strings.TrimSpace(statsBox.Find("li:nth-child(5)").Text())
	fields = strings.Fields(dob)
	if len(fields) > 1 {
		if fields[len(fields)-1] != "--" {
			fields = strings.Fields(dob)
			dobVal := fields[len(fields)-3:]
			dob = strings.Join(dobVal, " ")
			t, err := time.Parse("Jan 02, 2006", dob)
			if err == nil {
				fighter.DOB = &t
			}
		}
	}

	fightstats := e.DOM.Parent().Find("div.b-list__info-box-left div.b-list__info-box-left ul.b-list__box-list")

	slpm := strings.Fields(fightstats.Find("li:nth-child(1)").Text())
	if len(slpm) > 1 {
		fighter.SLPM, _ = strconv.ParseFloat(slpm[len(slpm)-1], 64)
	}

	strAcc := strings.Fields(fightstats.Find("li:nth-child(2)").Text())
	if len(strAcc) > 1 {
		strAccVal := strings.TrimSuffix(strAcc[len(strAcc)-1], "%")
		fighter.StrAcc, _ = strconv.Atoi(strAccVal)
	}

	sapm := strings.Fields(fightstats.Find("li:nth-child(3)").Text())
	if len(sapm) > 1 {
		fighter.SAPM, _ = strconv.ParseFloat(sapm[len(sapm)-1], 64)
	}

	strDef := strings.Fields(fightstats.Find("li:nth-child(4)").Text())
	if len(strDef) > 1 {
		strDefVal := strings.TrimSuffix(strDef[len(strDef)-1], "%")
		fighter.StrDef, _ = strconv.Atoi(strDefVal)
	}

	substats := e.DOM.Parent().Find("div.b-list__info-box-right.b-list__info-box_style-margin-right ul.b-list__box-list")
	tdAvg := strings.Fields(substats.Find("li:nth-child(2)").Text())
	if len(tdAvg) > 1 {
		fighter.TdAvg, _ = strconv.ParseFloat(tdAvg[len(tdAvg)-1], 64)
	}
	tdAcc := strings.Fields(substats.Find("li:nth-child(3)").Text())
	if len(tdAcc) > 1 {
		tdAccVal := strings.TrimSuffix(tdAcc[len(tdAcc)-1], "%")
		fighter.TdAcc, _ = strconv.Atoi(tdAccVal)
	}

	tdDef := strings.Fields(substats.Find("li:nth-child(4)").Text())
	if len(tdDef) > 1 {
		tdDefVal := strings.TrimSuffix(tdDef[len(tdDef)-1], "%")
		fighter.TdDef, _ = strconv.Atoi(tdDefVal)
	}

	subAvg := strings.Fields(substats.Find("li:nth-child(5)").Text())
	if len(subAvg) > 1 {
		fighter.SubAvg, _ = strconv.ParseFloat(subAvg[len(subAvg)-1], 64)
	}
	return fighter
}
