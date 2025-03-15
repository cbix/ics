package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

func init() {
	RegisterAdapter("meffis", meffisUrl, "Shared creative space in Aachen", meffisAdapter)
}

type meffisTextItem struct {
	Rendered string `json:"rendered"`
}

type meffisMeta struct {
	StartDate string `json:"_piecal_start_date"`
	EndDate   string `json:"_piecal_end_date"`
	Color     string `json:"_piecal_color"`
	IsEvent   bool   `json:"_piecal_is_event"`
	IsAllDay  bool   `json:"_piecal_is_allday"`
}

type meffisEvent struct {
	Id       int64          `json:"id"`
	Link     string         `json:"link"`
	Meta     meffisMeta     `json:"meta"`
	Modified string         `json:"modified"`
	Title    meffisTextItem `json:"title"`
	Content  meffisTextItem `json:"content"`
}

const (
	meffisUrl         = "https://www.meffis.org/"
	meffisSrc         = "https://www.meffis.org/wp-json/wp/v2/posts?per_page=100"
	meffisLocalFormat = "2006-01-02T15:04:05"
)

var (
	meffisTZ, _     = time.LoadLocation("Europe/Berlin")
	meffisStripTags = regexp.MustCompile(`<.*?>`)
	meffisJunk      = "[add-to-calendar"
)

func (t meffisTextItem) PlainText() string {
	str := t.Rendered
	if junk := strings.Index(str, meffisJunk); junk != -1 {
		str = str[:junk]
	}
	return html.UnescapeString(meffisStripTags.ReplaceAllString(str, ""))
}

func (e meffisEvent) ModifiedAt() time.Time {
	t, _ := time.ParseInLocation(meffisLocalFormat, e.Modified, meffisTZ)
	return t
}

func (e meffisEvent) StartTime() time.Time {
	t, _ := time.ParseInLocation(meffisLocalFormat, e.Meta.StartDate, meffisTZ)
	return t
}

func (e meffisEvent) EndTime() time.Time {
	t, _ := time.ParseInLocation(meffisLocalFormat, e.Meta.EndDate, meffisTZ)
	return t
}

func meffisAdapter(w io.Writer) error {
	now := time.Now()
	res, err := http.Get(meffisSrc)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var events []meffisEvent
	err = json.NewDecoder(res.Body).Decode(&events)
	if err != nil {
		return err
	}

	cal := ics.NewCalendarFor("github.com/cbix/ics")
	cal.SetName("meffi.s Aachen")
	cal.SetColor("#29603e")
	cal.SetUrl(meffisUrl)
	cal.SetTimezoneId("Europe/Berlin")
	lastModified := time.Time{}
	for _, ev := range events {
		if !ev.Meta.IsEvent {
			continue
		}
		event := cal.AddEvent(fmt.Sprintf("event-%d@meffis.org", ev.Id))
		event.SetDtStampTime(now)
		event.SetSummary(ev.Title.PlainText())
		event.SetDescription(ev.Content.PlainText())
		if ev.Meta.Color != "" {
			event.SetColor(ev.Meta.Color)
		}
		event.SetLocation("meffi.s Aachen")
		if ev.Meta.IsAllDay {
			event.SetAllDayStartAt(ev.StartTime())
			event.SetAllDayEndAt(ev.EndTime())
		} else {
			if ev.EndTime().After(ev.StartTime()) {
				event.SetEndAt(ev.EndTime())
				event.SetStartAt(ev.StartTime())
			}
		}
		event.SetURL(ev.Link)
		modifiedAt := ev.ModifiedAt()
		event.SetCreatedTime(modifiedAt)
		event.SetModifiedAt(modifiedAt)
		if modifiedAt.After(lastModified) {
			lastModified = modifiedAt
		}
	}
	if !lastModified.IsZero() {
		cal.SetLastModified(lastModified)
	}
	return cal.SerializeTo(w)
}
