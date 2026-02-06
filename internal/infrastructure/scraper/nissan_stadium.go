package scraper

import (
	"context"
	"errors"
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
	url   string
	title string
}

var errNotForNissanStadium = errors.New("event is not for Nissan Stadium")

func (s *NissanStadiumScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)

	slog.Info("fetching nissan stadium events", "date", today.Format("2006-01-02"))

	candidates, err := s.fetchEventCandidates(ctx, today)
	if err != nil {
		slog.Error("failed to fetch event candidates", "error", err)
		return nil, fmt.Errorf("failed to fetch event candidates: %w", err)
	}

	if len(candidates) == 0 {
		slog.Info("no event candidates found for today")
		return []event.Event{}, nil
	}

	slog.Info("found event candidates", "candidates", candidates)

	events, err := s.fetchEventDetails(ctx, candidates, today)
	if err != nil {
		slog.Error("failed to fetch event details", "error", err)
		return nil, fmt.Errorf("failed to fetch event details: %w", err)
	}

	slog.Info("fetched nissan stadium events", "count", len(events))

	return events, nil
}

func (s *NissanStadiumScraper) fetchEventCandidates(ctx context.Context, today time.Time) ([]eventCandidate, error) {
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
	detailURL := fmt.Sprintf("%s/calendar/detail.php?id%s", s.baseURL, id)

	slog.Debug("processing row", "date", *currentDate, "title", title, "href", href, "id", id)

	if id == "" || title == "" {
		return eventCandidate{}, false
	}

	slog.Debug("found event candidate", "id", id, "title", title)

	return eventCandidate{title: title, url: detailURL}, true
}

func (s *NissanStadiumScraper) fetchEventDetails(ctx context.Context, candidates []eventCandidate, today time.Time) ([]event.Event, error) {
	slog.Debug("fetching event details", "candidates", len(candidates))

	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(5)

	var results []event.Event
	var errorCount int
	var mu sync.Mutex

	for _, candidate := range candidates {
		candidate := candidate
		eg.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			evt, err := s.fetchEventDetail(ctx, candidate, today)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if errors.Is(err, errNotForNissanStadium) {
					slog.Info("skipping event not for Nissan Stadium", "url", candidate.url)
					return nil
				}
				errorCount++
				slog.Error("failed to fetch event detail", "error", err)
				return nil
			}

			if !evt.Date.IsZero() {
				results = append(results, evt)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if len(results) == 0 && errorCount > 0 {
		return nil, fmt.Errorf("all event detail fetches failed")
	}

	slog.Debug("event details fetched", "success", len(results), "errors", errorCount)

	return results, nil
}

type eventDetailFields struct {
	title string
	date  string
	time  string
	venue string
}

func (s *NissanStadiumScraper) fetchEventDetail(ctx context.Context, candidate eventCandidate, today time.Time) (event.Event, error) {
	c := colly.NewCollector()
	c.SetRequestTimeout(10 * time.Second)

	var fields eventDetailFields

	c.OnHTML("table tr", func(row *colly.HTMLElement) {
		s.parseDetailTableRow(row, &fields)
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

	slog.Debug("fetching event detail", "url", candidate.url)

	if err := c.Visit(candidate.url); err != nil {
		return event.Event{}, fmt.Errorf("failed to visit detail page for event %s: %w", candidate.url, err)
	}

	if visitErr != nil {
		return event.Event{}, visitErr
	}

	return s.buildEventFromFields(fields, candidate, today)
}

func (s *NissanStadiumScraper) parseDetailTableRow(row *colly.HTMLElement, fields *eventDetailFields) {
	th := strings.TrimSpace(row.ChildText("th"))
	td := strings.TrimSpace(row.ChildText("td"))

	switch th {
	case "行事名":
		fields.title = td
	case "期日":
		fields.date = td
	case "開始":
		fields.time = td
	case "対象施設":
		fields.venue = td
	}
}

func (s *NissanStadiumScraper) buildEventFromFields(fields eventDetailFields, candidate eventCandidate, today time.Time) (event.Event, error) {
	if !strings.Contains(fields.venue, "日産スタジアム") {
		return event.Event{}, errNotForNissanStadium
	}

	title := fields.title
	if title == "" {
		title = candidate.title
	}

	parsedDate, err := parseJapaneseDate(fields.date, today)
	if err != nil {
		return event.Event{}, fmt.Errorf("failed to parse date for event %s: %w", candidate.url, err)
	}

	var startTime *time.Time
	if fields.time != "" {
		t, err := parseJapaneseTime(fields.time, parsedDate)
		if err != nil {
			slog.Error("failed to parse event start time",
				"time", fields.time,
				"date", parsedDate,
				"url", candidate.url,
				"err", err,
			)
		} else {
			startTime = &t
		}
	}

	return event.Event{Title: title, Date: parsedDate, StartTime: startTime}, nil
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

func parseJapaneseDate(dateStr string, today time.Time) (time.Time, error) {
	if dateStr == "" {
		dateStr = fmt.Sprintf("%d年%d月%d日", today.Year(), today.Month(), today.Day())
	}

	jst := time.FixedZone("JST", 9*60*60)

	t, err := time.ParseInLocation("2006年1月2日", dateStr, jst)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date '%s': %w", dateStr, err)
	}

	return t, nil
}

func parseJapaneseTime(timeStr string, baseDate time.Time) (time.Time, error) {
	jst := time.FixedZone("JST", 9*60*60)

	layouts := []string{
		"15時04分",
		"15時4分",
		"15時",
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, timeStr, jst)
		if err == nil {
			return time.Date(
				baseDate.Year(), baseDate.Month(), baseDate.Day(),
				t.Hour(), t.Minute(), 0, 0, jst,
			), nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse time '%s'", timeStr)
}
