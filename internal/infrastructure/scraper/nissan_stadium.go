package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type NissanStadiumScraper struct {
	baseURL string
}

func NewNissanStadiumScraper() ports.EventFetcher {
	return &NissanStadiumScraper{
		baseURL: "https://www.nissan-stadium.jp",
	}
}

type eventCandidate struct {
	id    string
	title string
	date  int
}

func (s *NissanStadiumScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)

	slog.Info("fetching nissan stadium events", "date", today.Format("2006-01-02"))

	candidates, err := s.retrieveEventCandidates(ctx, today)
	if err != nil {
		slog.Error("failed to fetch event candidates", "error", err)
		return nil, fmt.Errorf("failed to fetch event candidates: %w", err)
	}

	if len(candidates) == 0 {
		slog.Info("no event candidates found for today")
		return []event.Event{}, nil
	}

	slog.Info("found event candidates", "count", len(candidates))

	events, err := s.fetchEventDetails(ctx, candidates, today)
	if err != nil {
		slog.Error("failed to fetch event details", "error", err)
		return nil, fmt.Errorf("failed to fetch event details: %w", err)
	}

	slog.Info("fetched nissan stadium events", "count", len(events))

	return events, nil
}

func (s *NissanStadiumScraper) retrieveEventCandidates(ctx context.Context, today time.Time) ([]eventCandidate, error) {
	c := colly.NewCollector()
	c.SetRequestTimeout(10 * time.Second)

	var candidates []eventCandidate
	var currentDate int
	targetDay := today.Day()

	c.OnHTML("#areacontents01 > div:nth-child(2) > table > tbody tr", func(row *colly.HTMLElement) {
		if candidate, ok := s.parseCalendarRow(row, &currentDate, targetDay); ok {
			candidates = append(candidates, candidate)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		slog.Error("error during calendar scraping", "status", r.StatusCode, "error", err)
	})

	var visitErr error
	c.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			visitErr = ctx.Err()
			r.Abort()
		default:
		}
	})

	calendarURL := s.baseURL + "/calendar/"
	slog.Debug("visiting calendar page", "url", calendarURL)

	if err := c.Visit(calendarURL); err != nil {
		return nil, fmt.Errorf("failed to visit calendar page: %w", err)
	}

	if visitErr != nil {
		return nil, visitErr
	}

	slog.Debug("calendar scraping completed", "candidates", len(candidates))

	return candidates, nil
}

func (s *NissanStadiumScraper) parseCalendarRow(row *colly.HTMLElement, currentDate *int, targetDay int) (eventCandidate, bool) {
	if dateStr := row.ChildText("th:nth-child(1)"); dateStr != "" {
		var date int
		if _, err := fmt.Sscanf(dateStr, "%d", &date); err == nil {
			*currentDate = date
		}
	}

	if *currentDate != targetDay {
		return eventCandidate{}, false
	}

	title := strings.TrimSpace(row.ChildText("td:nth-child(3) > a:nth-child(2)"))
	href := row.ChildAttr("td:nth-child(3) > a:nth-child(2)", "href")
	id := extractEventID(href)

	slog.Debug("processing row", "date", *currentDate, "title", title, "href", href, "id", id)

	if id == "" || title == "" {
		return eventCandidate{}, false
	}

	slog.Debug("found event candidate", "id", id, "title", title, "date", *currentDate)

	return eventCandidate{id: id, title: title, date: *currentDate}, true
}

func (s *NissanStadiumScraper) fetchEventDetails(ctx context.Context, candidates []eventCandidate, today time.Time) ([]event.Event, error) {
	slog.Debug("fetching event details", "candidates", len(candidates))

	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(5)

	results := make([]event.Event, len(candidates))
	errs := make([]error, len(candidates))
	var mu sync.Mutex

	for i, candidate := range candidates {
		i, candidate := i, candidate
		eg.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			evt, err := s.fetchEventDetail(ctx, candidate, today)

			mu.Lock()
			if err != nil {
				errs[i] = err
			} else {
				results[i] = evt
			}
			mu.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	var events []event.Event
	var errorCount int
	for i := range results {
		if errs[i] != nil {
			errorCount++
			slog.Warn("failed to fetch event detail", "error", errs[i])
			continue
		}
		if !results[i].Date.IsZero() {
			events = append(events, results[i])
		}
	}

	if len(events) == 0 && errorCount > 0 {
		return nil, fmt.Errorf("all event detail fetches failed")
	}

	slog.Debug("event details fetched", "success", len(events), "errors", errorCount)

	return events, nil
}

func (s *NissanStadiumScraper) fetchEventDetail(ctx context.Context, candidate eventCandidate, today time.Time) (event.Event, error) {
	c := colly.NewCollector(
		colly.UserAgent("shin-yokohama-event-notifier/1.0"),
	)

	c.SetRequestTimeout(10 * time.Second)

	var evt event.Event
	var eventTitle, eventDate, eventTime, eventVenue string

	c.OnHTML("table", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			th := strings.TrimSpace(row.ChildText("th"))
			td := strings.TrimSpace(row.ChildText("td"))

			switch th {
			case "行事名":
				eventTitle = td
			case "期日":
				eventDate = td
			case "開始":
				eventTime = td
			case "対象施設":
				eventVenue = td
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		slog.Error("error during event detail scraping", "status", r.StatusCode, "error", err)
	})

	var visitErr error
	c.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			visitErr = ctx.Err()
			r.Abort()
		default:
		}
	})

	detailURL := fmt.Sprintf("%s/calendar/detail.php?id=%s", s.baseURL, candidate.id)
	slog.Debug("fetching event detail", "id", candidate.id, "url", detailURL)

	err := c.Visit(detailURL)
	if err != nil {
		return evt, fmt.Errorf("failed to visit detail page for event %s: %w", candidate.id, err)
	}

	if visitErr != nil {
		return evt, visitErr
	}

	if !strings.Contains(eventVenue, "日産スタジアム") {
		return evt, fmt.Errorf("event %s is not for Nissan Stadium: %s", candidate.id, eventVenue)
	}

	if eventTitle == "" {
		eventTitle = candidate.title
	}

	if eventTime == "" {
		eventTime = "00:00"
	}

	parsedDate, err := parseJapaneseDateTime(eventDate, eventTime, today)
	if err != nil {
		return evt, fmt.Errorf("failed to parse date/time for event %s: %w", candidate.id, err)
	}

	evt.Title = eventTitle
	evt.Date = parsedDate

	return evt, nil
}

func (s *NissanStadiumScraper) VenueID() event.VenueID {
	return event.VenueIDNissanStadium
}

func extractEventID(href string) string {
	parts := strings.Split(href, "id")
	if len(parts) < 2 {
		return ""
	}

	id := strings.TrimPrefix(parts[1], "=")
	id = strings.TrimSpace(id)

	return id
}

func parseJapaneseDateTime(dateStr, timeStr string, today time.Time) (time.Time, error) {
	if dateStr == "" {
		dateStr = fmt.Sprintf("%d年%d月%d日", today.Year(), today.Month(), today.Day())
	}

	dateStr = strings.ReplaceAll(dateStr, "年", "-")
	dateStr = strings.ReplaceAll(dateStr, "月", "-")
	dateStr = strings.ReplaceAll(dateStr, "日", "")
	dateStr = strings.TrimSpace(dateStr)

	datetimeStr := dateStr + " " + timeStr

	jst := time.FixedZone("JST", 9*60*60)

	layouts := []string{
		"2006-1-2 15:04",
		"2006-01-02 15:04",
		"2006-1-2 15:4",
		"2006-01-02 15:4",
	}

	var parsedTime time.Time
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, datetimeStr, jst)
		if err == nil {
			parsedTime = t
			break
		}
		lastErr = err
	}

	if parsedTime.IsZero() {
		return time.Time{}, fmt.Errorf("failed to parse datetime '%s': %w", datetimeStr, lastErr)
	}

	return parsedTime, nil
}
