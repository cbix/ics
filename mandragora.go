package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	ics "github.com/arran4/golang-ical"
)

func init() {
	RegisterAdapter("mandragora", mandragoraUrl, "Restaurant in Bochum with occasional live music", mandragoraAdapter)
}

const (
	mandragoraUrl   = "https://mandragora-bochum.eatbu.com/"
	mandragoraTZ    = "Europe/Berlin"
	mandragoraTitle = `h3.event-title`
)

var mandragoraLocation, _ = time.LoadLocation(mandragoraTZ)

type mandragoraEvent struct {
	Title       string
	Description string
	Image       string
	StartAt     time.Time
}

// extractEvents uses goquery CSS selectors to parse the events section.
func extractEvents(doc *goquery.Document) ([]mandragoraEvent, error) {
	var events []mandragoraEvent

	doc.Find(".mod-website-events .story-block-container").Each(func(_ int, s *goquery.Selection) {
		ev := mandragoraEvent{}

		// Title: <h3 class="event-title">...</h3>
		s.Find(mandragoraTitle).Each(func(_ int, title *goquery.Selection) {
			ev.Title = strings.TrimSpace(title.Text())
		})

		// Description: all <p> tags inside .text-block
		s.Find(".text-block p").Each(func(_ int, p *goquery.Selection) {
			text := strings.TrimSpace(p.Text())
			if text != "" {
				ev.Description += " " + text
			}
		})
		ev.Description = strings.TrimSpace(ev.Description)

		// Image: <img data-src="...">
		s.Find("img").Each(func(_ int, img *goquery.Selection) {
			if src, ok := img.Attr("data-src"); ok {
				ev.Image = src
			}
		})

		// Parse date from title: "BAND - DD.MM.YYYY - HH UHR"
		ev.StartAt = parseDateFromTitle(ev.Title)

		// Skip events without a parseable date
		if ev.Title != "" && !ev.StartAt.IsZero() {
			events = append(events, ev)
		}
	})

	return events, nil
}

// parseDateFromTitle extracts date+time from a title like "BAND - DD.MM.YYYY - HH UHR".
func parseDateFromTitle(title string) time.Time {
	// Find DD.MM.YYYY pattern (10 chars: dd.mm.yyyy)
	var day, month, year int
	var hour int

	for i := 0; i <= len(title)-10; i++ {
		part := title[i : i+10]
		if part[2] == '.' && part[5] == '.' {
			d, err1 := strconv.Atoi(part[0:2])
			m, err2 := strconv.Atoi(part[3:5])
			y, err3 := strconv.Atoi(part[6:10])
			if err1 == nil && err2 == nil && err3 == nil && y >= 2000 {
				day, month, year = d, m, y
				// Extract hour after the date
				dateEnd := i + 10
				remaining := strings.TrimSpace(title[dateEnd:])
				// " - HH UHR" or " - HH"
				if idx := strings.Index(remaining, "UHR"); idx > 0 {
					hourStr := strings.TrimSpace(remaining[:idx])
					// strip leading " - "
					hourStr = strings.TrimPrefix(hourStr, "-")
					hourStr = strings.TrimSpace(hourStr)
					fmt.Sscanf(hourStr, "%d", &hour)
				}
				break
			}
		}
	}

	if year == 0 {
		return time.Time{}
	}

	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, mandragoraLocation)
}

func mandragoraAdapter(w io.Writer) error {
	now := time.Now()

	htmlContent, err := fetchHTML(mandragoraUrl)
	if err != nil {
		return fmt.Errorf("fetching mandragora: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return fmt.Errorf("parsing HTML: %w", err)
	}

	events, err := extractEvents(doc)
	if err != nil {
		return fmt.Errorf("extracting events: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("no events found")
	}

	cal := ics.NewCalendarFor("github.com/cbix/ics")
	cal.SetName("Mandragora Bochum")
	cal.SetUrl(mandragoraUrl)
	cal.SetTimezoneId(mandragoraTZ)

	for i, ev := range events {
		event := cal.AddEvent(fmt.Sprintf("event-%d@mandragora-bochum.eatbu.com", i+1))
		event.SetDtStampTime(now)
		event.SetSummary(ev.Title)
		event.SetDescription(ev.Description)
		event.SetLocation("Mandragora Bochum")
		event.SetStartAt(ev.StartAt)
		event.SetEndAt(ev.StartAt.Add(3 * time.Hour))
		if ev.Image != "" {
			event.SetURL(ev.Image)
		}
	}

	return cal.SerializeTo(w)
}
