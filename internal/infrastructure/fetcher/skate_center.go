package fetcher

import (
	"context"
	"encoding/json"
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

func (s *SkateCenterFetcher) FetchEvents(ctx context.Context) ([]event.Event, error) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	slog.Info("fetching skate center events", "date", todayStr)

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
		if t.In(jst).Format("2006-01-02") != todayStr {
			continue
		}
		events = append(events, buildSkateCenterEvent(raw, t, today))
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

	req.Header.Set("User-Agent", "shin-yokohama-event-notifier/1.0 (+https://github.com/Eagle-Konbu/shin-yokohama-event-notifier)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ja,en;q=0.8")

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
	var parseErr error
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if parseErr != nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					if n.FirstChild != nil {
						var raw json.RawMessage
						text := n.FirstChild.Data
						if err := json.Unmarshal([]byte(text), &raw); err != nil {
							parseErr = fmt.Errorf("failed to unmarshal JSON-LD: %w", err)
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

	if parseErr != nil {
		return nil, parseErr
	}

	return events, nil
}

func buildSkateCenterEvent(raw jsonLDEvent, startTime time.Time, today time.Time) event.Event {
	jst := time.FixedZone("JST", 9*60*60)
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, jst)
	startTimeJST := startTime.In(jst)

	return event.Event{
		Title: raw.Name,
		Date:  date,
		Schedules: []event.Schedule{
			{StartTime: &startTimeJST},
		},
	}
}

func (s *SkateCenterFetcher) VenueID() event.VenueID {
	return event.VenueIDSkateCenter
}
