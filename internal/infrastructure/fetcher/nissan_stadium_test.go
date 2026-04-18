package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

func TestNewNissanStadiumFetcher(t *testing.T) {
	scraper := NewNissanStadiumFetcher()

	require.NotNil(t, scraper)
	nissanScraper, ok := scraper.(*NissanStadiumFetcher)
	require.True(t, ok)
	assert.Equal(t, "https://www.nissan-stadium.jp", nissanScraper.baseURL)
}

func TestNissanStadiumFetcher_VenueID(t *testing.T) {
	scraper := NewNissanStadiumFetcher()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDNissanStadium, vid)
}

func TestNissanStadiumFetcher_FetchEvents_Success_SingleEvent(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := createMockCalendarHTML(currentDay, "サッカー練習試合", "691aa8fccc37e", "日産スタジアム")
	detailHTML := createMockDetailHTML("サッカー練習試合", fmt.Sprintf("2026年1月%d日", currentDay), "14時", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "サッカー練習試合", events[0].Title)
	assert.Equal(t, currentDay, events[0].Date.Day())
	require.Len(t, events[0].Schedules, 1)
	require.NotNil(t, events[0].Schedules[0].StartTime)
	assert.Equal(t, 14, events[0].Schedules[0].StartTime.Hour())
	assert.Equal(t, 0, events[0].Schedules[0].StartTime.Minute())
}

func TestNissanStadiumFetcher_FetchEvents_Success_MultipleEvents(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>%d</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event1">イベント1</a></td>
						</tr>
						<tr>
							<th></th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event2">イベント2</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>
	`, currentDay)

	detailHTML1 := createMockDetailHTML("イベント1", fmt.Sprintf("2026年1月%d日", currentDay), "10時", "日産スタジアム")
	detailHTML2 := createMockDetailHTML("イベント2", fmt.Sprintf("2026年1月%d日", currentDay), "15時", "日産スタジアム")

	server := createMockServerMultiDetail(calendarHTML, map[string]string{
		"event1": detailHTML1,
		"event2": detailHTML2,
	})
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	require.Len(t, events, 2)
}

func TestNissanStadiumFetcher_FetchEvents_NoEventsToday(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	otherDay := today.AddDate(0, 0, 1).Day()

	calendarHTML := createMockCalendarHTML(otherDay, "サッカー練習試合", "691aa8fccc37e", "日産スタジアム")
	detailHTML := createMockDetailHTML("サッカー練習試合", "2026年1月15日", "14時", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestNissanStadiumFetcher_FetchEvents_FiltersByVenue(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>%d</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event1">イベント1</a></td>
						</tr>
						<tr>
							<th></th>
							<td>火</td>
							<td><a href="#">小机競技場</a><a href="detail.php?id=event2">イベント2</a></td>
						</tr>
						<tr>
							<th></th>
							<td>火</td>
							<td><a href="#">フットボールパーク</a><a href="detail.php?id=event3">イベント3</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>
	`, currentDay)

	detailHTML := createMockDetailHTML("イベント1", fmt.Sprintf("2026年1月%d日", currentDay), "14時", "日産スタジアム")

	server := createMockServerMultiDetail(calendarHTML, map[string]string{
		"event1": detailHTML,
	})
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "イベント1", events[0].Title)
}

func TestNissanStadiumFetcher_FetchEvents_CalendarFetchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, time.Now(), time.Now())

	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "failed to fetch event candidates")
}

func TestNissanStadiumFetcher_FetchEvents_ContextCancellation(t *testing.T) {
	calendarHTML := createMockCalendarHTML(28, "イベント", "event1", "日産スタジアム")
	detailHTML := createMockDetailHTML("イベント", "2026年1月28日", "14時", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events, err := scraper.FetchEvents(ctx, time.Now(), time.Now())

	require.Error(t, err)
	assert.Nil(t, events)
}

func TestNissanStadiumFetcher_FetchEvents_MissingTime_DefaultsToZero(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := createMockCalendarHTML(currentDay, "終日イベント", "event1", "日産スタジアム")
	detailHTML := createMockDetailHTML("終日イベント", fmt.Sprintf("2026年1月%d日", currentDay), "", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Empty(t, events[0].Schedules)
}

func TestNissanStadiumFetcher_FetchEvents_PartialFailure(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>%d</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event1">イベント1</a></td>
						</tr>
						<tr>
							<th></th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event2">イベント2</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>
	`, currentDay)

	detailHTML := createMockDetailHTML("イベント1", fmt.Sprintf("2026年1月%d日", currentDay), "14時", "日産スタジアム")

	server := createMockServerMultiDetail(calendarHTML, map[string]string{
		"event1": detailHTML,
	})
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "イベント1", events[0].Title)
}

func TestNissanStadiumFetcher_FetchEvents_EmptyIDOrTitle(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>%d</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=">イベント1</a></td>
						</tr>
						<tr>
							<th></th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event2"></a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>
	`, currentDay)

	server := createMockServer(calendarHTML, "")
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestNissanStadiumFetcher_FetchEvents_TitleFallback(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := createMockCalendarHTML(currentDay, "カレンダータイトル", "event1", "日産スタジアム")
	detailHTML := createMockDetailHTMLWithoutTitle(fmt.Sprintf("2026年1月%d日", currentDay), "14時", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "カレンダータイトル", events[0].Title)
}

func TestNissanStadiumFetcher_FetchEvents_AllDetailsFailed(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := fmt.Sprintf(`
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>%d</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event1">イベント1</a></td>
						</tr>
						<tr>
							<th></th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=event2">イベント2</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>
	`, currentDay)

	detailHTML1 := createMockDetailHTMLWithInvalidDate("イベント1", "日産スタジアム")
	detailHTML2 := createMockDetailHTMLWithInvalidDate("イベント2", "日産スタジアム")

	server := createMockServerMultiDetail(calendarHTML, map[string]string{
		"event1": detailHTML1,
		"event2": detailHTML2,
	})
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.Error(t, err)
	assert.Nil(t, events)
	assert.Contains(t, err.Error(), "failed to fetch event details")
}

func TestNissanStadiumFetcher_FetchEvents_InvalidURL(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	calendarHTML := createMockCalendarHTML(currentDay, "イベント", "event1", "日産スタジアム")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/calendar/") && !strings.Contains(r.URL.Path, "detail") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			//nolint:errcheck
			io.WriteString(w, calendarHTML)
		} else {
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, err := hj.Hijack()
				if err == nil {
					conn.Close()
				}
			}
		}
	}))
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	require.Error(t, err)
	assert.Nil(t, events)
}

func TestNissanStadiumFetcher_FetchEvents_InvalidTimeFormat_LogsError(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	currentDay := today.Day()

	// Create an event with an invalid time format that will fail parsing
	calendarHTML := createMockCalendarHTML(currentDay, "イベント", "event1", "日産スタジアム")
	detailHTML := createMockDetailHTML("イベント", fmt.Sprintf("2026年1月%d日", currentDay), "invalid time format", "日産スタジアム")

	server := createMockServer(calendarHTML, detailHTML)
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, today, today)

	// The event should still be returned, but without StartTime
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "イベント", events[0].Title)
	assert.Empty(t, events[0].Schedules, "Schedules should be empty when time parsing fails")
}

func TestNissanStadiumFetcher_FetchEvents_DateRange_SameMonth(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	from := time.Date(2026, 1, 20, 0, 0, 0, 0, jst)
	to := time.Date(2026, 1, 26, 0, 0, 0, 0, jst)

	calendarHTML := `
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>19</th>
							<td>月</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=before1">範囲前イベント</a></td>
						</tr>
						<tr>
							<th>20</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=day20">20日イベント</a></td>
						</tr>
						<tr>
							<th>23</th>
							<td>金</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=day23">23日イベント</a></td>
						</tr>
						<tr>
							<th>26</th>
							<td>月</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=day26">26日イベント</a></td>
						</tr>
						<tr>
							<th>27</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=after1">範囲後イベント</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>`

	detailHTMLMap := map[string]string{
		"day20": createMockDetailHTML("20日イベント", "2026年1月20日", "10時", "日産スタジアム"),
		"day23": createMockDetailHTML("23日イベント", "2026年1月23日", "14時", "日産スタジアム"),
		"day26": createMockDetailHTML("26日イベント", "2026年1月26日", "18時", "日産スタジアム"),
	}

	var calendarCount atomic.Int32
	var detailCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/calendar/") && !strings.Contains(r.URL.Path, "detail") {
			calendarCount.Add(1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			//nolint:errcheck
			io.WriteString(w, calendarHTML)
		} else if strings.Contains(r.URL.Path, "detail.php") {
			detailCount.Add(1)
			rawQuery := r.URL.RawQuery
			eventID := strings.TrimPrefix(rawQuery, "id")
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
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, from, to)

	require.NoError(t, err)
	require.Len(t, events, 3)
	assert.Equal(t, int32(1), calendarCount.Load())
	assert.Equal(t, int32(3), detailCount.Load())
}

func TestNissanStadiumFetcher_FetchEvents_DateRange_CrossMonth(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	from := time.Date(2026, 1, 28, 0, 0, 0, 0, jst)
	to := time.Date(2026, 2, 3, 0, 0, 0, 0, jst)

	currentMonthCalendar := `
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>1</th>
							<td>木</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=jan1">1月1日イベント</a></td>
						</tr>
						<tr>
							<th>2</th>
							<td>金</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=jan2">1月2日イベント</a></td>
						</tr>
						<tr>
							<th>3</th>
							<td>土</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=jan3">1月3日イベント</a></td>
						</tr>
						<tr>
							<th>4</th>
							<td>日</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=jan4">1月4日イベント</a></td>
						</tr>
						<tr>
							<th>5</th>
							<td>月</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=jan5">1月5日イベント</a></td>
						</tr>
						<tr>
							<th>28</th>
							<td>水</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=jan28">1月28日イベント</a></td>
						</tr>
						<tr>
							<th>30</th>
							<td>金</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=jan30">1月30日イベント</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>`

	nextMonthCalendar := `
		<html><body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>1</th>
							<td>日</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=feb1">2月1日イベント</a></td>
						</tr>
						<tr>
							<th>2</th>
							<td>月</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=feb2">2月2日イベント</a></td>
						</tr>
						<tr>
							<th>3</th>
							<td>火</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=feb3">2月3日イベント</a></td>
						</tr>
						<tr>
							<th>4</th>
							<td>水</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=feb4">2月4日イベント</a></td>
						</tr>
						<tr>
							<th>5</th>
							<td>木</td>
							<td><a href="#">日産スタジアム</a><a href="detail.php?id=feb5">2月5日イベント</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body></html>`

	detailHTMLMap := map[string]string{
		"jan28": createMockDetailHTML("1月28日イベント", "2026年1月28日", "14時", "日産スタジアム"),
		"jan30": createMockDetailHTML("1月30日イベント", "2026年1月30日", "18時", "日産スタジアム"),
		"feb1":  createMockDetailHTML("2月1日イベント", "2026年2月1日", "10時", "日産スタジアム"),
		"feb2":  createMockDetailHTML("2月2日イベント", "2026年2月2日", "12時", "日産スタジアム"),
		"feb3":  createMockDetailHTML("2月3日イベント", "2026年2月3日", "15時", "日産スタジアム"),
	}

	var calendarCount atomic.Int32
	var detailCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/calendar/") && !strings.Contains(r.URL.Path, "detail") {
			calendarCount.Add(1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if r.URL.Query().Get("m") == "x" {
				//nolint:errcheck
				io.WriteString(w, nextMonthCalendar)
			} else {
				//nolint:errcheck
				io.WriteString(w, currentMonthCalendar)
			}
		} else if strings.Contains(r.URL.Path, "detail.php") {
			detailCount.Add(1)
			rawQuery := r.URL.RawQuery
			eventID := strings.TrimPrefix(rawQuery, "id")
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
	defer server.Close()

	scraper := &NissanStadiumFetcher{baseURL: server.URL}
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx, from, to)

	require.NoError(t, err)
	require.Len(t, events, 5)
	assert.Equal(t, int32(2), calendarCount.Load())
	assert.Equal(t, int32(5), detailCount.Load())
}

func TestBuildTargetDays(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)

	tests := []struct {
		from     time.Time
		to       time.Time
		expected map[int]bool
		name     string
	}{
		{
			name:     "single day",
			from:     time.Date(2026, 1, 20, 0, 0, 0, 0, jst),
			to:       time.Date(2026, 1, 20, 0, 0, 0, 0, jst),
			expected: map[int]bool{20: true},
		},
		{
			name: "week range same month",
			from: time.Date(2026, 1, 20, 0, 0, 0, 0, jst),
			to:   time.Date(2026, 1, 26, 0, 0, 0, 0, jst),
			expected: map[int]bool{
				20: true, 21: true, 22: true, 23: true, 24: true, 25: true, 26: true,
			},
		},
		{
			name: "cross month",
			from: time.Date(2026, 1, 30, 0, 0, 0, 0, jst),
			to:   time.Date(2026, 2, 2, 0, 0, 0, 0, jst),
			expected: map[int]bool{
				30: true, 31: true, 1: true, 2: true,
			},
		},
		{
			name: "non-midnight times normalized",
			from: time.Date(2026, 1, 20, 15, 30, 0, 0, jst),
			to:   time.Date(2026, 1, 22, 8, 0, 0, 0, jst),
			expected: map[int]bool{
				20: true, 21: true, 22: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTargetDays(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNissanStadiumFetcher_FetchEvents_RangeExceedsTwoMonths(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, jst)
	to := time.Date(2026, 3, 1, 0, 0, 0, 0, jst)

	scraper := &NissanStadiumFetcher{baseURL: "http://localhost"}

	events, err := scraper.FetchEvents(context.Background(), from, to)

	require.Error(t, err)
	assert.Nil(t, events)
	assert.ErrorIs(t, err, errRangeExceedsLimit)
}

func TestDistinctMonthCount(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)

	tests := []struct {
		from     time.Time
		to       time.Time
		name     string
		expected int
	}{
		{
			name:     "same month",
			from:     time.Date(2026, 4, 1, 0, 0, 0, 0, jst),
			to:       time.Date(2026, 4, 30, 0, 0, 0, 0, jst),
			expected: 1,
		},
		{
			name:     "two months",
			from:     time.Date(2026, 4, 27, 0, 0, 0, 0, jst),
			to:       time.Date(2026, 5, 3, 0, 0, 0, 0, jst),
			expected: 2,
		},
		{
			name:     "three months",
			from:     time.Date(2026, 1, 1, 0, 0, 0, 0, jst),
			to:       time.Date(2026, 3, 1, 0, 0, 0, 0, jst),
			expected: 3,
		},
		{
			name:     "cross year",
			from:     time.Date(2026, 12, 1, 0, 0, 0, 0, jst),
			to:       time.Date(2027, 1, 1, 0, 0, 0, 0, jst),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distinctMonthCount(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
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
		{
			name:     "href without id substring",
			href:     "other.php?event=abc",
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

func createMockCalendarHTML(day int, eventTitle, eventID, venue string) string {
	return fmt.Sprintf(`
		<html>
		<body>
		<div id="areacontents01">
			<div></div>
			<div>
				<table>
					<tbody>
						<tr>
							<th>%d</th>
							<td>火</td>
							<td><a href="#">%s</a><a href="detail.php?id=%s">%s</a></td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
		</body>
		</html>
	`, day, venue, eventID, eventTitle)
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

func createMockDetailHTMLWithoutTitle(date, time, venue string) string {
	timeRow := ""
	if time != "" {
		timeRow = fmt.Sprintf("<tr><th>開始</th><td>%s</td></tr>", time)
	}

	return fmt.Sprintf(`
		<html>
		<body>
		<table>
			<tr><th>行事名</th><td></td></tr>
			<tr><th>期日</th><td>%s</td></tr>
			%s
			<tr><th>対象施設</th><td>%s</td></tr>
		</table>
		</body>
		</html>
	`, date, timeRow, venue)
}

func createMockDetailHTMLWithInvalidDate(title, venue string) string {
	return fmt.Sprintf(`
		<html>
		<body>
		<table>
			<tr><th>行事名</th><td>%s</td></tr>
			<tr><th>期日</th><td>invalid date</td></tr>
			<tr><th>開始</th><td>invalid time</td></tr>
			<tr><th>対象施設</th><td>%s</td></tr>
		</table>
		</body>
		</html>
	`, title, venue)
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
			rawQuery := r.URL.RawQuery
			eventID := strings.TrimPrefix(rawQuery, "id")
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
