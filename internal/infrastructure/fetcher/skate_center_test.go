package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

func TestNewSkateCenterFetcher(t *testing.T) {
	scraper := NewSkateCenterFetcher()

	require.NotNil(t, scraper)
	skateScraper, ok := scraper.(*SkateCenterFetcher)
	require.True(t, ok)
	assert.Equal(t, "https://ticketjam.jp", skateScraper.baseURL)
}

func TestSkateCenterFetcher_VenueID(t *testing.T) {
	scraper := NewSkateCenterFetcher()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDSkateCenter, vid)
}

func TestSkateCenterFetcher_FetchEvents_SingleEvent(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	startDate := time.Date(today.Year(), today.Month(), today.Day(), 11, 0, 0, 0, jst)

	htmlResp := createSkateCenterHTML(fmt.Sprintf(`{
		"@type": "Event",
		"name": "テストイベント",
		"startDate": "%s",
		"location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}
	}`, startDate.Format(time.RFC3339)))

	server := createSkateCenterMockServer(htmlResp)
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "テストイベント", events[0].Title)
	assert.Equal(t, today.Day(), events[0].Date.Day())
	require.Len(t, events[0].Schedules, 1)
	require.NotNil(t, events[0].Schedules[0].StartTime)
	assert.Equal(t, 11, events[0].Schedules[0].StartTime.Hour())
	assert.Equal(t, 0, events[0].Schedules[0].StartTime.Minute())
}

func TestSkateCenterFetcher_FetchEvents_MultipleEvents(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	startDate1 := time.Date(today.Year(), today.Month(), today.Day(), 11, 0, 0, 0, jst)
	startDate2 := time.Date(today.Year(), today.Month(), today.Day(), 18, 30, 0, 0, jst)

	htmlResp := createSkateCenterHTMLMultiple(
		fmt.Sprintf(`{"@type": "Event", "name": "イベント1", "startDate": "%s", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`, startDate1.Format(time.RFC3339)),
		fmt.Sprintf(`{"@type": "Event", "name": "イベント2", "startDate": "%s", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`, startDate2.Format(time.RFC3339)),
	)

	server := createSkateCenterMockServer(htmlResp)
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "イベント1", events[0].Title)
	assert.Equal(t, "イベント2", events[1].Title)
	require.Len(t, events[1].Schedules, 1)
	require.NotNil(t, events[1].Schedules[0].StartTime)
	assert.Equal(t, 18, events[1].Schedules[0].StartTime.Hour())
	assert.Equal(t, 30, events[1].Schedules[0].StartTime.Minute())
}

func TestSkateCenterFetcher_FetchEvents_NoEventsToday(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	tomorrow := time.Now().In(jst).AddDate(0, 0, 1)
	startDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 11, 0, 0, 0, jst)

	htmlResp := createSkateCenterHTML(fmt.Sprintf(`{
		"@type": "Event",
		"name": "明日のイベント",
		"startDate": "%s",
		"location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}
	}`, startDate.Format(time.RFC3339)))

	server := createSkateCenterMockServer(htmlResp)
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now().In(jst), time.Now().In(jst))

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestSkateCenterFetcher_FetchEvents_EmptyPage(t *testing.T) {
	server := createSkateCenterMockServer(`<html><body></body></html>`)
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now(), time.Now())

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestSkateCenterFetcher_FetchEvents_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now(), time.Now())

	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestSkateCenterFetcher_FetchEvents_ContextCancellation(t *testing.T) {
	server := createSkateCenterMockServer(`<html><body></body></html>`)
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events, err := scraper.FetchEvents(ctx, time.Now(), time.Now())

	require.Error(t, err)
	assert.Nil(t, events)
}

func TestSkateCenterFetcher_FetchEvents_InvalidJSON(t *testing.T) {
	htmlResp := `<html><head><script type="application/ld+json">{invalid json}</script></head><body></body></html>`

	server := createSkateCenterMockServer(htmlResp)
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now(), time.Now())

	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "failed to extract JSON-LD events")
}

func TestSkateCenterFetcher_FetchEvents_NonEventType(t *testing.T) {
	htmlResp := createSkateCenterHTML(`{
		"@type": "Organization",
		"name": "KOSE新横浜スケートセンター"
	}`)

	server := createSkateCenterMockServer(htmlResp)
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now(), time.Now())

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestSkateCenterFetcher_FetchEvents_DateRange(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	from := time.Date(2026, 4, 20, 0, 0, 0, 0, jst)
	to := time.Date(2026, 4, 26, 0, 0, 0, 0, jst)

	htmlResp := createSkateCenterHTMLMultiple(
		`{"@type": "Event", "name": "範囲前イベント", "startDate": "2026-04-19T18:00:00+09:00", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`,
		`{"@type": "Event", "name": "初日イベント", "startDate": "2026-04-20T11:00:00+09:00", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`,
		`{"@type": "Event", "name": "中間イベント", "startDate": "2026-04-23T14:00:00+09:00", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`,
		`{"@type": "Event", "name": "最終日イベント", "startDate": "2026-04-26T18:00:00+09:00", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`,
		`{"@type": "Event", "name": "範囲後イベント", "startDate": "2026-04-27T10:00:00+09:00", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`,
	)

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//nolint:errcheck
		io.WriteString(w, htmlResp)
	}))
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), from, to)

	require.NoError(t, err)
	require.Len(t, events, 3)
	assert.Equal(t, "初日イベント", events[0].Title)
	assert.Equal(t, 20, events[0].Date.Day())
	assert.Equal(t, "中間イベント", events[1].Title)
	assert.Equal(t, 23, events[1].Date.Day())
	assert.Equal(t, "最終日イベント", events[2].Title)
	assert.Equal(t, 26, events[2].Date.Day())
	assert.Equal(t, 1, requestCount)
}

func TestSkateCenterFetcher_FetchEvents_DateRange_CrossMonth(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	from := time.Date(2026, 4, 28, 0, 0, 0, 0, jst)
	to := time.Date(2026, 5, 4, 0, 0, 0, 0, jst)

	htmlResp := createSkateCenterHTMLMultiple(
		`{"@type": "Event", "name": "4月イベント", "startDate": "2026-04-29T11:00:00+09:00", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`,
		`{"@type": "Event", "name": "5月イベント", "startDate": "2026-05-02T14:00:00+09:00", "location": {"@type": "Place", "name": "KOSE新横浜スケートセンター"}}`,
	)

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//nolint:errcheck
		io.WriteString(w, htmlResp)
	}))
	defer server.Close()

	scraper := &SkateCenterFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), from, to)

	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "4月イベント", events[0].Title)
	assert.Equal(t, time.April, events[0].Date.Month())
	assert.Equal(t, "5月イベント", events[1].Title)
	assert.Equal(t, time.May, events[1].Date.Month())
	assert.Equal(t, 1, requestCount)
}

func createSkateCenterMockServer(htmlResponse string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//nolint:errcheck
		io.WriteString(w, htmlResponse)
	}))
}

func createSkateCenterHTML(jsonLD string) string {
	return fmt.Sprintf(`<html><head><script type="application/ld+json">%s</script></head><body></body></html>`, jsonLD)
}

func createSkateCenterHTMLMultiple(jsonLDs ...string) string {
	var scripts string
	for _, j := range jsonLDs {
		scripts += fmt.Sprintf(`<script type="application/ld+json">%s</script>`, j)
	}
	return fmt.Sprintf(`<html><head>%s</head><body></body></html>`, scripts)
}
