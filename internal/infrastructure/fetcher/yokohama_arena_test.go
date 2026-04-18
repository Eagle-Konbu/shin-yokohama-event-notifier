package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

func TestNewYokohamaArenaFetcher(t *testing.T) {
	scraper := NewYokohamaArenaFetcher()

	require.NotNil(t, scraper)
	arenaScraper, ok := scraper.(*YokohamaArenaFetcher)
	require.True(t, ok)
	assert.Equal(t, "https://www.yokohama-arena.co.jp", arenaScraper.baseURL)
}

func TestYokohamaArenaFetcher_VenueID(t *testing.T) {
	scraper := NewYokohamaArenaFetcher()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDYokohamaArena, vid)
}

func TestYokohamaArenaFetcher_FetchEvents_SingleEventSingleTime(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "テストイベント",
		"ev_open": ["16:00"],
		"ev_start": ["17:00"],
		"path": "/event/detail/test"
	}]`, todayStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "テストイベント", events[0].Title)
	assert.Equal(t, today.Day(), events[0].Date.Day())
	require.Len(t, events[0].Schedules, 1)
	require.NotNil(t, events[0].Schedules[0].StartTime)
	assert.Equal(t, 17, events[0].Schedules[0].StartTime.Hour())
	assert.Equal(t, 0, events[0].Schedules[0].StartTime.Minute())
	require.NotNil(t, events[0].Schedules[0].OpenTime)
	assert.Equal(t, 16, events[0].Schedules[0].OpenTime.Hour())
	assert.Equal(t, 0, events[0].Schedules[0].OpenTime.Minute())
}

func TestYokohamaArenaFetcher_FetchEvents_SingleEventMultipleTimes(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "複数公演イベント",
		"ev_open": ["\u24ea11:30", "\u24ea16:30"],
		"ev_start": ["\u24ea12:30", "\u24ea17:30"],
		"path": "/event/detail/multi"
	}]`, todayStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "複数公演イベント", events[0].Title)

	require.Len(t, events[0].Schedules, 2)

	require.NotNil(t, events[0].Schedules[0].StartTime)
	assert.Equal(t, 12, events[0].Schedules[0].StartTime.Hour())
	assert.Equal(t, 30, events[0].Schedules[0].StartTime.Minute())
	require.NotNil(t, events[0].Schedules[0].OpenTime)
	assert.Equal(t, 11, events[0].Schedules[0].OpenTime.Hour())
	assert.Equal(t, 30, events[0].Schedules[0].OpenTime.Minute())

	require.NotNil(t, events[0].Schedules[1].StartTime)
	assert.Equal(t, 17, events[0].Schedules[1].StartTime.Hour())
	assert.Equal(t, 30, events[0].Schedules[1].StartTime.Minute())
	require.NotNil(t, events[0].Schedules[1].OpenTime)
	assert.Equal(t, 16, events[0].Schedules[1].OpenTime.Hour())
	assert.Equal(t, 30, events[0].Schedules[1].OpenTime.Minute())
}

func TestYokohamaArenaFetcher_FetchEvents_NoEventsToday(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	tomorrow := time.Now().In(jst).AddDate(0, 0, 1)
	tomorrowStr := tomorrow.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "明日のイベント",
		"ev_open": ["16:00"],
		"ev_start": ["17:00"],
		"path": "/event/detail/tomorrow"
	}]`, tomorrowStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now().In(jst), time.Now().In(jst))

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestYokohamaArenaFetcher_FetchEvents_EmptyPath(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "リンクなしイベント",
		"ev_open": [],
		"ev_start": [],
		"path": ""
	}]`, todayStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestYokohamaArenaFetcher_FetchEvents_EmptyResponse(t *testing.T) {
	server := createYokohamaArenaMockServer(`[]`)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now(), time.Now())

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestYokohamaArenaFetcher_FetchEvents_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), time.Now(), time.Now())

	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestYokohamaArenaFetcher_FetchEvents_ContextCancellation(t *testing.T) {
	server := createYokohamaArenaMockServer(`[]`)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events, err := scraper.FetchEvents(ctx, time.Now(), time.Now())

	require.Error(t, err)
	assert.Nil(t, events)
}

func TestYokohamaArenaFetcher_FetchEvents_NoStartTime(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "開場のみイベント",
		"ev_open": ["10:00"],
		"ev_start": [],
		"path": "/event/detail/open-only"
	}]`, todayStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "開場のみイベント", events[0].Title)
	require.Len(t, events[0].Schedules, 1)
	assert.Nil(t, events[0].Schedules[0].StartTime)
	require.NotNil(t, events[0].Schedules[0].OpenTime)
	assert.Equal(t, 10, events[0].Schedules[0].OpenTime.Hour())
}

func TestYokohamaArenaFetcher_FetchEvents_NoTimes(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "時間未定イベント",
		"ev_open": [],
		"ev_start": [],
		"path": "/event/detail/no-time"
	}]`, todayStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "時間未定イベント", events[0].Title)
	assert.Empty(t, events[0].Schedules)
}

func TestYokohamaArenaFetcher_FetchEvents_MixedTypeFieldsIgnored(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "テスト",
		"artist": false,
		"ev_open": ["16:00"],
		"ev_start": ["17:00"],
		"ev_end": [],
		"path": "/event/detail/test",
		"ticket_url_ja": false,
		"ticket_url_en": false
	}]`, todayStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "テスト", events[0].Title)
}

func TestYokohamaArenaFetcher_FetchEvents_FullwidthColon(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	todayStr := today.Format("2006-01-02")

	jsonResp := fmt.Sprintf(`[{
		"date1": "%s",
		"title": "全角コロンイベント",
		"ev_open": ["15\uff1a00"],
		"ev_start": ["16\uff1a00"],
		"path": "/event/detail/fullwidth"
	}]`, todayStr)

	server := createYokohamaArenaMockServer(jsonResp)
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Len(t, events[0].Schedules, 1)
	require.NotNil(t, events[0].Schedules[0].StartTime)
	assert.Equal(t, 16, events[0].Schedules[0].StartTime.Hour())
	assert.Equal(t, 0, events[0].Schedules[0].StartTime.Minute())
	require.NotNil(t, events[0].Schedules[0].OpenTime)
	assert.Equal(t, 15, events[0].Schedules[0].OpenTime.Hour())
}

func TestYokohamaArenaFetcher_FetchEvents_DateRange_SameMonth(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	from := time.Date(2026, 4, 20, 0, 0, 0, 0, jst)
	to := time.Date(2026, 4, 26, 0, 0, 0, 0, jst)

	jsonResp := `[
		{"date1": "2026-04-19", "title": "範囲前イベント", "ev_open": [], "ev_start": ["10:00"], "path": "/event/detail/before"},
		{"date1": "2026-04-20", "title": "初日イベント", "ev_open": [], "ev_start": ["11:00"], "path": "/event/detail/first"},
		{"date1": "2026-04-23", "title": "中間イベント", "ev_open": [], "ev_start": ["14:00"], "path": "/event/detail/mid"},
		{"date1": "2026-04-26", "title": "最終日イベント", "ev_open": [], "ev_start": ["18:00"], "path": "/event/detail/last"},
		{"date1": "2026-04-27", "title": "範囲後イベント", "ev_open": [], "ev_start": ["19:00"], "path": "/event/detail/after"}
	]`

	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck
		io.WriteString(w, jsonResp)
	}))
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
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

func TestYokohamaArenaFetcher_FetchEvents_DateRange_CrossMonth(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	from := time.Date(2026, 4, 27, 0, 0, 0, 0, jst)
	to := time.Date(2026, 5, 3, 0, 0, 0, 0, jst)

	aprilJSON := `[
		{"date1": "2026-04-26", "title": "範囲前イベント", "path": "/e/1", "ev_open": [], "ev_start": []},
		{"date1": "2026-04-27", "title": "4月末イベント", "path": "/e/2", "ev_open": [], "ev_start": ["18:00"]},
		{"date1": "2026-04-30", "title": "4月最終日イベント", "path": "/e/3", "ev_open": [], "ev_start": ["19:00"]}
	]`
	mayJSON := `[
		{"date1": "2026-05-01", "title": "5月初日イベント", "path": "/e/4", "ev_open": [], "ev_start": ["10:00"]},
		{"date1": "2026-05-03", "title": "5月3日イベント", "path": "/e/5", "ev_open": [], "ev_start": ["14:00"]},
		{"date1": "2026-05-04", "title": "範囲後イベント", "path": "/e/6", "ev_open": [], "ev_start": ["15:00"]}
	]`

	var requestedPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "202604") {
			//nolint:errcheck
			io.WriteString(w, aprilJSON)
		} else if strings.Contains(r.URL.Path, "202605") {
			//nolint:errcheck
			io.WriteString(w, mayJSON)
		} else {
			//nolint:errcheck
			io.WriteString(w, "[]")
		}
	}))
	defer server.Close()

	scraper := &YokohamaArenaFetcher{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background(), from, to)

	require.NoError(t, err)
	require.Len(t, events, 4)
	assert.Equal(t, "4月末イベント", events[0].Title)
	assert.Equal(t, time.April, events[0].Date.Month())
	assert.Equal(t, "4月最終日イベント", events[1].Title)
	assert.Equal(t, "5月初日イベント", events[2].Title)
	assert.Equal(t, time.May, events[2].Date.Month())
	assert.Equal(t, "5月3日イベント", events[3].Title)

	assert.Len(t, requestedPaths, 2)
}

func TestDistinctYearMonths(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)

	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		expected []string
	}{
		{
			name:     "same month",
			from:     time.Date(2026, 4, 20, 0, 0, 0, 0, jst),
			to:       time.Date(2026, 4, 26, 0, 0, 0, 0, jst),
			expected: []string{"202604"},
		},
		{
			name:     "cross month",
			from:     time.Date(2026, 4, 27, 0, 0, 0, 0, jst),
			to:       time.Date(2026, 5, 3, 0, 0, 0, 0, jst),
			expected: []string{"202604", "202605"},
		},
		{
			name:     "single day",
			from:     time.Date(2026, 4, 19, 0, 0, 0, 0, jst),
			to:       time.Date(2026, 4, 19, 0, 0, 0, 0, jst),
			expected: []string{"202604"},
		},
		{
			name:     "cross year",
			from:     time.Date(2026, 12, 28, 0, 0, 0, 0, jst),
			to:       time.Date(2027, 1, 3, 0, 0, 0, 0, jst),
			expected: []string{"202612", "202701"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distinctYearMonths(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripCircledNumberPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "circled zero prefix",
			input:    "⓪12:30",
			expected: "12:30",
		},
		{
			name:     "circled one prefix",
			input:    "①17:30",
			expected: "17:30",
		},
		{
			name:     "circled twenty prefix",
			input:    "⑳09:00",
			expected: "09:00",
		},
		{
			name:     "no prefix",
			input:    "12:30",
			expected: "12:30",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripCircledNumberPrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func createYokohamaArenaMockServer(jsonResponse string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck
		io.WriteString(w, jsonResponse)
	}))
}
