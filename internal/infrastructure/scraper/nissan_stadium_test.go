package scraper

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

func TestNewNissanStadiumScraper(t *testing.T) {
	scraper := NewNissanStadiumScraper()

	require.NotNil(t, scraper)
	nissanScraper, ok := scraper.(*NissanStadiumScraper)
	require.True(t, ok)
	assert.Equal(t, "https://www.nissan-stadium.jp", nissanScraper.baseURL)
}

func TestNissanStadiumScraper_VenueID(t *testing.T) {
	scraper := NewNissanStadiumScraper()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDNissanStadium, vid)
}

func TestNissanStadiumScraper_FetchEvents_Success_SingleEvent(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	testTime := time.Date(2026, 1, 28, 0, 0, 0, 0, jst)

	calendarHTML := createMockCalendarHTML(28, "サッカー練習試合", "691aa8fccc37e", "日産スタジアム")
	detailHTML := createMockDetailHTML("サッカー練習試合", "2026年1月28日", "14:00", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "サッカー練習試合", events[0].Title)
	assert.Equal(t, 28, events[0].Date.Day())
	assert.Equal(t, 14, events[0].Date.Hour())
	assert.Equal(t, 0, events[0].Date.Minute())
	assert.True(t, events[0].Date.After(testTime) || events[0].Date.Equal(testTime))
}

func TestNissanStadiumScraper_FetchEvents_Success_MultipleEvents(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div>%d 火</div>
		<a href="detail.php?id=event1">イベント1</a> 日産スタジアム<br>
		<a href="detail.php?id=event2">イベント2</a> 日産スタジアム<br>
		</body></html>
	`, currentDay)

	detailHTML1 := createMockDetailHTML("イベント1", fmt.Sprintf("2026年1月%d日", currentDay), "10:00", "日産スタジアム")
	detailHTML2 := createMockDetailHTML("イベント2", fmt.Sprintf("2026年1月%d日", currentDay), "15:00", "日産スタジアム")

	server := createMockServerMultiDetail(calendarHTML, map[string]string{
		"event1": detailHTML1,
		"event2": detailHTML2,
	})
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.NoError(t, err)
	require.Len(t, events, 2)
}

func TestNissanStadiumScraper_FetchEvents_NoEventsToday(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	otherDay := (today.Day() + 1) % 31
	if otherDay == 0 {
		otherDay = 1
	}

	calendarHTML := createMockCalendarHTML(otherDay, "サッカー練習試合", "691aa8fccc37e", "日産スタジアム")
	detailHTML := createMockDetailHTML("サッカー練習試合", "2026年1月15日", "14:00", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestNissanStadiumScraper_FetchEvents_FiltersByVenue(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div>%d 火</div>
		<a href="detail.php?id=event1">イベント1</a> 日産スタジアム<br>
		<a href="detail.php?id=event2">イベント2</a> 小机競技場<br>
		<a href="detail.php?id=event3">イベント3</a> しんよこフットボールパーク<br>
		</body></html>
	`, currentDay)

	detailHTML := createMockDetailHTML("イベント1", fmt.Sprintf("2026年1月%d日", currentDay), "14:00", "日産スタジアム")

	server := createMockServerMultiDetail(calendarHTML, map[string]string{
		"event1": detailHTML,
	})
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "イベント1", events[0].Title)
}

func TestNissanStadiumScraper_FetchEvents_CalendarFetchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "failed to fetch calendar candidates")
}

func TestNissanStadiumScraper_FetchEvents_ContextCancellation(t *testing.T) {
	calendarHTML := createMockCalendarHTML(28, "イベント", "event1", "日産スタジアム")
	detailHTML := createMockDetailHTML("イベント", "2026年1月28日", "14:00", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events, err := scraper.FetchEvents(ctx)

	require.Error(t, err)
	assert.Nil(t, events)
}

func TestNissanStadiumScraper_FetchEvents_MissingTime_DefaultsToZero(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := createMockCalendarHTML(currentDay, "終日イベント", "event1", "日産スタジアム")
	detailHTML := createMockDetailHTML("終日イベント", fmt.Sprintf("2026年1月%d日", currentDay), "", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, 0, events[0].Date.Hour())
	assert.Equal(t, 0, events[0].Date.Minute())
}

func TestNissanStadiumScraper_FetchEvents_PartialFailure(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div>%d 火</div>
		<a href="detail.php?id=event1">イベント1</a> 日産スタジアム<br>
		<a href="detail.php?id=event2">イベント2</a> 日産スタジアム<br>
		</body></html>
	`, currentDay)

	detailHTML := createMockDetailHTML("イベント1", fmt.Sprintf("2026年1月%d日", currentDay), "14:00", "日産スタジアム")

	server := createMockServerMultiDetail(calendarHTML, map[string]string{
		"event1": detailHTML,
	})
	defer server.Close()

	scraper := &NissanStadiumScraper{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "イベント1", events[0].Title)
}

func TestExtractEventID(t *testing.T) {
	tests := []struct {
		name     string
		href     string
		expected string
	}{
		{
			name:     "valid event ID",
			href:     "detail.php?id=691aa8fccc37e",
			expected: "691aa8fccc37e",
		},
		{
			name:     "valid event ID with prefix",
			href:     "detail.php?id691aa8fccc37e",
			expected: "691aa8fccc37e",
		},
		{
			name:     "no id parameter",
			href:     "detail.php",
			expected: "",
		},
		{
			name:     "empty href",
			href:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractEventID(tt.href)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseJapaneseDateTime(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)

	tests := []struct {
		name        string
		dateStr     string
		timeStr     string
		expectedDay int
		expectedHr  int
		expectedMin int
		wantErr     bool
	}{
		{
			name:        "valid date and time",
			dateStr:     "2026年1月28日",
			timeStr:     "14:00",
			expectedDay: 28,
			expectedHr:  14,
			expectedMin: 0,
			wantErr:     false,
		},
		{
			name:        "valid date with single digit month and day",
			dateStr:     "2026年1月5日",
			timeStr:     "09:30",
			expectedDay: 5,
			expectedHr:  9,
			expectedMin: 30,
			wantErr:     false,
		},
		{
			name:        "empty time defaults to 00:00",
			dateStr:     "2026年1月28日",
			timeStr:     "00:00",
			expectedDay: 28,
			expectedHr:  0,
			expectedMin: 0,
			wantErr:     false,
		},
		{
			name:        "empty date uses today",
			dateStr:     "",
			timeStr:     "14:00",
			expectedDay: today.Day(),
			expectedHr:  14,
			expectedMin: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJapaneseDateTime(tt.dateStr, tt.timeStr, today)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDay, result.Day())
				assert.Equal(t, tt.expectedHr, result.Hour())
				assert.Equal(t, tt.expectedMin, result.Minute())
			}
		})
	}
}

func createMockCalendarHTML(day int, eventTitle, eventID, venue string) string {
	return fmt.Sprintf(`
		<html>
		<body>
		<div>%d 火</div>
		<a href="detail.php?id=%s">%s</a> %s
		</body>
		</html>
	`, day, eventID, eventTitle, venue)
}

func createMockDetailHTML(title, date, time, venue string) string {
	timeRow := ""
	if time != "" {
		timeRow = fmt.Sprintf("<tr><th>開始</th><td>%s</td></tr>", time)
	}

	return fmt.Sprintf(`
		<html>
		<body>
		<table>
			<tr><th>行事名</th><td>%s</td></tr>
			<tr><th>期日</th><td>%s</td></tr>
			%s
			<tr><th>対象施設</th><td>%s</td></tr>
		</table>
		</body>
		</html>
	`, title, date, timeRow, venue)
}

func createMockServer(calendarHTML, detailHTML string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/calendar/") && !strings.Contains(r.URL.Path, "detail") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			//nolint:errcheck
			io.WriteString(w, calendarHTML)
		} else if strings.Contains(r.URL.Path, "detail.php") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			//nolint:errcheck
			io.WriteString(w, detailHTML)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func createMockServerMultiDetail(calendarHTML string, detailHTMLMap map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/calendar/") && !strings.Contains(r.URL.Path, "detail") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			//nolint:errcheck
			io.WriteString(w, calendarHTML)
		} else if strings.Contains(r.URL.Path, "detail.php") {
			eventID := r.URL.Query().Get("id")
			if detailHTML, ok := detailHTMLMap[eventID]; ok {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				//nolint:errcheck
				io.WriteString(w, detailHTML)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
