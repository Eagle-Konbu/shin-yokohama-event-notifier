package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type YokohamaArenaScraper struct {
	baseURL string
}

func NewYokohamaArenaScraper() ports.EventFetcher {
	return &YokohamaArenaScraper{
		baseURL: "https://www.yokohama-arena.co.jp",
	}
}

type yokohamaArenaEvent struct {
	Date1   string   `json:"date1"`
	Title   string   `json:"title"`
	Path    string   `json:"path"`
	EvOpen  []string `json:"ev_open"`
	EvStart []string `json:"ev_start"`
}

func (s *YokohamaArenaScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")
	yearMonth := today.Format("200601")

	slog.Info("fetching yokohama arena events", "date", todayStr)

	apiURL := fmt.Sprintf("%s/event/%s?_format=json", s.baseURL, yearMonth)

	rawEvents, err := s.fetchJSON(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch yokohama arena events: %w", err)
	}

	var events []event.Event
	for _, raw := range rawEvents {
		if raw.Path == "" || raw.Date1 != todayStr {
			continue
		}
		expanded := s.expandEvents(raw, today)
		events = append(events, expanded...)
	}

	slog.Info("fetched yokohama arena events", "count", len(events))

	return events, nil
}

func (s *YokohamaArenaScraper) fetchJSON(ctx context.Context, apiURL string) ([]yokohamaArenaEvent, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rawEvents []yokohamaArenaEvent
	if err := json.NewDecoder(resp.Body).Decode(&rawEvents); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return rawEvents, nil
}

func (s *YokohamaArenaScraper) expandEvents(raw yokohamaArenaEvent, today time.Time) []event.Event {
	jst := time.FixedZone("JST", 9*60*60)
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, jst)

	n := len(raw.EvStart)
	if len(raw.EvOpen) > n {
		n = len(raw.EvOpen)
	}

	if n == 0 {
		return []event.Event{{Title: raw.Title, Date: date}}
	}

	var events []event.Event
	for i := range n {
		evt := event.Event{Title: raw.Title, Date: date}

		if i < len(raw.EvStart) {
			if t, err := parseArenaTime(raw.EvStart[i], date); err == nil {
				evt.StartTime = &t
			} else {
				slog.Error("failed to parse start time", "time", raw.EvStart[i], "err", err)
			}
		}

		if i < len(raw.EvOpen) {
			if t, err := parseArenaTime(raw.EvOpen[i], date); err == nil {
				evt.OpenTime = &t
			} else {
				slog.Error("failed to parse open time", "time", raw.EvOpen[i], "err", err)
			}
		}

		events = append(events, evt)
	}

	return events
}

func parseArenaTime(s string, baseDate time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	s = stripCircledNumberPrefix(s)
	s = strings.ReplaceAll(s, "：", ":")

	jst := time.FixedZone("JST", 9*60*60)
	t, err := time.ParseInLocation("15:04", s, jst)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time '%s': %w", s, err)
	}

	return time.Date(
		baseDate.Year(), baseDate.Month(), baseDate.Day(),
		t.Hour(), t.Minute(), 0, 0, jst,
	), nil
}

func stripCircledNumberPrefix(s string) string {
	if len(s) == 0 {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == '\u24EA' || (r >= '\u2460' && r <= '\u2473') {
		return s[size:]
	}
	return s
}

func (s *YokohamaArenaScraper) VenueID() event.VenueID {
	return event.VenueIDYokohamaArena
}
