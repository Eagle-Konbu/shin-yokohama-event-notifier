package scraper

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

func TestNewYokohamaArenaScraper(t *testing.T) {
	scraper := NewYokohamaArenaScraper()

	require.NotNil(t, scraper)
	arenaScraper, ok := scraper.(*YokohamaArenaScraper)
	require.True(t, ok)
	assert.Equal(t, "https://www.yokohama-arena.co.jp", arenaScraper.baseURL)
}

func TestYokohamaArenaScraper_VenueID(t *testing.T) {
	scraper := NewYokohamaArenaScraper()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDYokohamaArena, vid)
}

func TestYokohamaArenaScraper_FetchEvents_SingleEventSingleTime(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "テストイベント", events[0].Title)
	assert.Equal(t, today.Day(), events[0].Date.Day())
	require.NotNil(t, events[0].StartTime)
	assert.Equal(t, 17, events[0].StartTime.Hour())
	assert.Equal(t, 0, events[0].StartTime.Minute())
	require.NotNil(t, events[0].OpenTime)
	assert.Equal(t, 16, events[0].OpenTime.Hour())
	assert.Equal(t, 0, events[0].OpenTime.Minute())
}

func TestYokohamaArenaScraper_FetchEvents_SingleEventMultipleTimes(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	require.Len(t, events, 2)

	assert.Equal(t, "複数公演イベント", events[0].Title)
	require.NotNil(t, events[0].StartTime)
	assert.Equal(t, 12, events[0].StartTime.Hour())
	assert.Equal(t, 30, events[0].StartTime.Minute())
	require.NotNil(t, events[0].OpenTime)
	assert.Equal(t, 11, events[0].OpenTime.Hour())
	assert.Equal(t, 30, events[0].OpenTime.Minute())

	assert.Equal(t, "複数公演イベント", events[1].Title)
	require.NotNil(t, events[1].StartTime)
	assert.Equal(t, 17, events[1].StartTime.Hour())
	assert.Equal(t, 30, events[1].StartTime.Minute())
	require.NotNil(t, events[1].OpenTime)
	assert.Equal(t, 16, events[1].OpenTime.Hour())
	assert.Equal(t, 30, events[1].OpenTime.Minute())
}

func TestYokohamaArenaScraper_FetchEvents_NoEventsToday(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestYokohamaArenaScraper_FetchEvents_EmptyPath(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestYokohamaArenaScraper_FetchEvents_EmptyResponse(t *testing.T) {
	server := createYokohamaArenaMockServer(`[]`)
	defer server.Close()

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestYokohamaArenaScraper_FetchEvents_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestYokohamaArenaScraper_FetchEvents_ContextCancellation(t *testing.T) {
	server := createYokohamaArenaMockServer(`[]`)
	defer server.Close()

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events, err := scraper.FetchEvents(ctx)

	require.Error(t, err)
	assert.Nil(t, events)
}

func TestYokohamaArenaScraper_FetchEvents_NoStartTime(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "開場のみイベント", events[0].Title)
	assert.Nil(t, events[0].StartTime)
	require.NotNil(t, events[0].OpenTime)
	assert.Equal(t, 10, events[0].OpenTime.Hour())
}

func TestYokohamaArenaScraper_FetchEvents_NoTimes(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "時間未定イベント", events[0].Title)
	assert.Nil(t, events[0].StartTime)
	assert.Nil(t, events[0].OpenTime)
}

func TestYokohamaArenaScraper_FetchEvents_MixedTypeFieldsIgnored(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "テスト", events[0].Title)
}

func TestYokohamaArenaScraper_FetchEvents_FullwidthColon(t *testing.T) {
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

	scraper := &YokohamaArenaScraper{baseURL: server.URL}
	events, err := scraper.FetchEvents(context.Background())

	require.NoError(t, err)
	require.Len(t, events, 1)
	require.NotNil(t, events[0].StartTime)
	assert.Equal(t, 16, events[0].StartTime.Hour())
	assert.Equal(t, 0, events[0].StartTime.Minute())
	require.NotNil(t, events[0].OpenTime)
	assert.Equal(t, 15, events[0].OpenTime.Hour())
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
