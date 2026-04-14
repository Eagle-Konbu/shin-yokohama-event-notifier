package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type SkateCenterFetcher struct {
	baseURL string
}

func NewSkateCenterFetcher() ports.EventFetcher {
	return &SkateCenterFetcher{
		baseURL: "https://ticketjam.jp",
	}
}

type jsonLDEvent struct {
	Type      string         `json:"@type"`
	Name      string         `json:"name"`
	StartDate string         `json:"startDate"`
	Location  jsonLDLocation `json:"location"`
}

type jsonLDLocation struct {
	Name string `json:"name"`
}

func (s *SkateCenterFetcher) FetchEvents(ctx context.Context, date time.Time) ([]event.Event, error) {
	jst := time.FixedZone("JST", 9*60*60)
	target := date.In(jst)
	targetStr := target.Format("2006-01-02")

	slog.Info("fetching skate center events", "date", targetStr)

	htmlContent, err := s.fetchHTML(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch skate center events: %w", err)
	}

	rawEvents, err := extractJSONLDEvents(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON-LD events: %w", err)
	}

	var events []event.Event
	for _, raw := range rawEvents {
		t, err := time.Parse(time.RFC3339, raw.StartDate)
		if err != nil {
			slog.Error("failed to parse startDate", "startDate", raw.StartDate, "err", err)
			continue
		}
		if t.In(jst).Format("2006-01-02") != targetStr {
			continue
		}
		events = append(events, buildSkateCenterEvent(raw, target))
	}

	slog.Info("fetched skate center events", "count", len(events))

	return events, nil
}

func (s *SkateCenterFetcher) fetchHTML(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/venues/3442", s.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func extractJSONLDEvents(htmlContent string) ([]jsonLDEvent, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var events []jsonLDEvent
	var parseErrs []error
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					if n.FirstChild != nil {
						var raw json.RawMessage
						text := n.FirstChild.Data
						if err := json.Unmarshal([]byte(text), &raw); err != nil {
							slog.Error("failed to unmarshal JSON-LD", "err", err)
							parseErrs = append(parseErrs, err)
							break
						}

						// Try single event
						var single jsonLDEvent
						if err := json.Unmarshal(raw, &single); err == nil && single.Type == "Event" {
							events = append(events, single)
							break
						}

						// Try array of events
						var arr []jsonLDEvent
						if err := json.Unmarshal(raw, &arr); err == nil {
							for _, e := range arr {
								if e.Type == "Event" {
									events = append(events, e)
								}
							}
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	if len(events) == 0 && len(parseErrs) > 0 {
		return nil, fmt.Errorf("failed to parse %d JSON-LD block(s), no events extracted: %w", len(parseErrs), errors.Join(parseErrs...))
	}

	return events, nil
}

func buildSkateCenterEvent(raw jsonLDEvent, today time.Time) event.Event {
	jst := time.FixedZone("JST", 9*60*60)
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, jst)

	evt := event.Event{
		Title: raw.Name,
		Date:  date,
	}

	t, err := time.Parse(time.RFC3339, raw.StartDate)
	if err == nil {
		startTime := t.In(jst)
		schedule := event.Schedule{
			StartTime: &startTime,
		}
		evt.Schedules = append(evt.Schedules, schedule)
	}

	return evt
}

func (s *SkateCenterFetcher) VenueID() event.VenueID {
	return event.VenueIDSkateCenter
}
