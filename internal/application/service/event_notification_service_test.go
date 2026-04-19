package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports/mock_ports"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func setupSingleFetcherService(t *testing.T) (*mock_ports.MockNotificationSender, *mock_ports.MockEventFetcher, *EventNotificationService, context.Context) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	return mockSender, mockFetcher, service, context.Background()
}

func setupThreeFetcherService(t *testing.T) (*mock_ports.MockNotificationSender, *mock_ports.MockEventFetcher, *mock_ports.MockEventFetcher, *mock_ports.MockEventFetcher, *EventNotificationService, context.Context) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher1 := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher2 := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher3 := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher1.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	mockFetcher2.EXPECT().VenueID().Return(event.VenueIDNissanStadium).AnyTimes()
	mockFetcher3.EXPECT().VenueID().Return(event.VenueIDSkateCenter).AnyTimes()
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher1, mockFetcher2, mockFetcher3})
	return mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, context.Background()
}

func TestNewEventNotificationService(t *testing.T) {
	_, _, service, _ := setupSingleFetcherService(t)

	require.NotNil(t, service)
	assert.NotNil(t, service.notificationSender)
	assert.Len(t, service.eventFetchers, 1)
}

func TestNotifyTodayEvents_NoEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, "📅 新横浜 イベント情報", sentNotification.Title())
	assert.Equal(t, "本日の開催イベントはありません", sentNotification.Description())
	assert.Equal(t, notification.ColorGreen, sentNotification.Color())
	assert.Len(t, sentNotification.Fields(), 3)
	for _, field := range sentNotification.Fields() {
		assert.Equal(t, "本日の予定はありません", field.Value)
	}
}

func TestNotifyTodayEvents_OneVenueWithEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "テストイベント",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
			Schedules: []event.Schedule{
				{StartTime: timePtr(time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local))},
			},
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, "本日のイベント数: 1件", sentNotification.Description())
	assert.Equal(t, notification.ColorYellow, sentNotification.Color())

	arenaField := sentNotification.Fields()[0]
	assert.Equal(t, "🏟️ 横浜アリーナ", arenaField.Name)
	assert.Contains(t, arenaField.Value, "・**18:00開始** テストイベント")
}

func TestNotifyTodayEvents_TwoVenuesWithEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	arenaEvents := []event.Event{
		{
			Title: "横浜アリーナイベント",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
		},
	}
	stadiumEvents := []event.Event{
		{
			Title: "日産スタジアムイベント",
			Date:  time.Date(2026, 1, 28, 14, 0, 0, 0, time.Local),
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(arenaEvents, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(stadiumEvents, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, "本日のイベント数: 2件", sentNotification.Description())
	assert.Equal(t, notification.ColorRed, sentNotification.Color())
}

func TestNotifyTodayEvents_AllVenuesWithEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	arenaEvents := []event.Event{
		{
			Title: "横浜アリーナイベント",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
		},
	}
	stadiumEvents := []event.Event{
		{
			Title: "日産スタジアムイベント",
			Date:  time.Date(2026, 1, 28, 14, 0, 0, 0, time.Local),
		},
	}
	skateEvents := []event.Event{
		{
			Title: "スケートセンターイベント",
			Date:  time.Date(2026, 1, 28, 10, 0, 0, 0, time.Local),
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(arenaEvents, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(stadiumEvents, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(skateEvents, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, "本日のイベント数: 3件", sentNotification.Description())
	assert.Equal(t, notification.ColorRed, sentNotification.Color())
}

func TestNotifyTodayEvents_MultipleEventsAtSameVenue(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "イベントB",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
			Schedules: []event.Schedule{
				{StartTime: timePtr(time.Date(2026, 1, 28, 19, 0, 0, 0, time.Local))},
			},
		},
		{
			Title: "イベントA",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
			Schedules: []event.Schedule{
				{StartTime: timePtr(time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local))},
			},
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Contains(t, arenaField.Value, "・**18:00開始** イベントA\n・**19:00開始** イベントB")
}

func TestNotifyTodayEvents_EventWithoutStartTime(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "時間未定イベント",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Equal(t, "・時間未定イベント", arenaField.Value)
	assert.NotContains(t, arenaField.Value, "開始")
}

func TestNotifyTodayEvents_MixedStartTimeEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "時間ありイベント",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
			Schedules: []event.Schedule{
				{StartTime: timePtr(time.Date(2026, 1, 28, 14, 0, 0, 0, time.Local))},
			},
		},
		{
			Title: "時間なしイベント",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Contains(t, arenaField.Value, "・時間なしイベント")
	assert.Contains(t, arenaField.Value, "・**14:00開始** 時間ありイベント")
}

func TestNotifyTodayEvents_EventWithOpenTime(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "開場時間のみイベント",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
			Schedules: []event.Schedule{
				{OpenTime: timePtr(time.Date(2026, 1, 28, 17, 0, 0, 0, time.Local))},
			},
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Equal(t, "・**17:00開場** 開場時間のみイベント", arenaField.Value)
}

func TestNotifyTodayEvents_EventWithBothOpenAndStartTime(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "開場開始両方イベント",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
			Schedules: []event.Schedule{
				{
					OpenTime:  timePtr(time.Date(2026, 1, 28, 17, 0, 0, 0, time.Local)),
					StartTime: timePtr(time.Date(2026, 1, 28, 18, 30, 0, 0, time.Local)),
				},
			},
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Equal(t, "・**17:00開場 / 18:30開始** 開場開始両方イベント", arenaField.Value)
}

func TestNotifyTodayEvents_EventWithMultipleSchedules(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "複数公演イベント",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
			Schedules: []event.Schedule{
				{
					OpenTime:  timePtr(time.Date(2026, 1, 28, 11, 30, 0, 0, time.Local)),
					StartTime: timePtr(time.Date(2026, 1, 28, 12, 30, 0, 0, time.Local)),
				},
				{
					OpenTime:  timePtr(time.Date(2026, 1, 28, 16, 30, 0, 0, time.Local)),
					StartTime: timePtr(time.Date(2026, 1, 28, 17, 30, 0, 0, time.Local)),
				},
			},
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Equal(t, "・**①11:30開場 / 12:30開始 ②16:30開場 / 17:30開始** 複数公演イベント", arenaField.Value)
}

func TestNotifyTodayEvents_VenueOrder(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	require.Len(t, sentNotification.Fields(), 3)
	assert.Equal(t, "🏟️ 横浜アリーナ", sentNotification.Fields()[0].Name)
	assert.Equal(t, "⚽ 日産スタジアム", sentNotification.Fields()[1].Name)
	assert.Equal(t, "⛸️ KOSÉ新横浜スケートセンター", sentNotification.Fields()[2].Name)
}

func TestNotifyTodayEvents_FetchError(t *testing.T) {
	mockSender, mockFetcher, service, ctx := setupSingleFetcherService(t)
	expectedErr := errors.New("fetch error")

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch events")
	assert.ErrorIs(t, err, expectedErr)

	require.NotNil(t, sentNotification)
	assert.Equal(t, "❌ イベント取得エラー", sentNotification.Title())
	assert.Equal(t, "イベント情報の取得に失敗しました", sentNotification.Description())
	assert.Equal(t, notification.ColorRed, sentNotification.Color())
}

func TestNotifyTodayEvents_FetchError_SendFailureNotificationFails(t *testing.T) {
	mockSender, mockFetcher, service, ctx := setupSingleFetcherService(t)
	fetchErr := errors.New("fetch error")
	sendErr := errors.New("send error")

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fetchErr)
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(sendErr)

	err := service.NotifyTodayEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch events")
	assert.Contains(t, err.Error(), "failed to send failure notification")
	assert.ErrorIs(t, err, fetchErr)
	assert.ErrorIs(t, err, sendErr)
}

func TestNotifyTodayEvents_SendError(t *testing.T) {
	mockSender, mockFetcher, service, ctx := setupSingleFetcherService(t)
	expectedErr := errors.New("send error")

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(expectedErr)

	err := service.NotifyTodayEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send notification")
	assert.ErrorIs(t, err, expectedErr)
}

func TestNotifyTodayEvents_BothNilStartTime_SortsByTitle(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	events := []event.Event{
		{
			Title: "イベントC",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
		},
		{
			Title: "イベントA",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
		},
		{
			Title: "イベントB",
			Date:  time.Date(2026, 1, 28, 0, 0, 0, 0, time.Local),
		},
	}
	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(events, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Contains(t, arenaField.Value, "・イベントA\n・イベントB\n・イベントC")
}

// Weekly notification tests

func TestNotifyWeeklyEvents_NoEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService(t)

	mockFetcher1.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher2.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockFetcher3.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyWeeklyEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, "📅 新横浜 週間イベント情報", sentNotification.Title())
	assert.Equal(t, notification.ColorGreen, sentNotification.Color())
	assert.Len(t, sentNotification.Fields(), 3)
	for _, field := range sentNotification.Fields() {
		assert.Equal(t, "今週の予定はありません", field.Value)
	}
}

func TestNotifyWeeklyEvents_FetchError(t *testing.T) {
	mockSender, mockFetcher, service, ctx := setupSingleFetcherService(t)
	expectedErr := errors.New("fetch error")

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	var sentNotification *notification.Notification
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, notif *notification.Notification) error {
		sentNotification = notif
		return nil
	})

	err := service.NotifyWeeklyEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch events")
	assert.ErrorIs(t, err, expectedErr)

	require.NotNil(t, sentNotification)
	assert.Equal(t, "❌ イベント取得エラー", sentNotification.Title())
}

// Tests for weekly formatting logic (direct unit tests of formatVenueWeeklyEvents)

func TestFormatVenueWeeklyEvents_NoEvents(t *testing.T) {
	svc := &EventNotificationService{}
	result := svc.formatVenueWeeklyEvents(nil)
	assert.Equal(t, "今週の予定はありません", result)
}

func TestFormatVenueWeeklyEvents_GroupedByDate(t *testing.T) {
	svc := &EventNotificationService{}

	jst := time.FixedZone("JST", 9*60*60)
	day1 := time.Date(2026, 4, 6, 0, 0, 0, 0, jst) // Monday
	day3 := time.Date(2026, 4, 8, 0, 0, 0, 0, jst) // Wednesday

	events := []event.Event{
		{
			Title: "月曜イベント",
			Date:  day1,
			Schedules: []event.Schedule{
				{StartTime: timePtr(time.Date(2026, 4, 6, 18, 0, 0, 0, jst))},
			},
		},
		{
			Title: "水曜イベント",
			Date:  day3,
		},
	}

	result := svc.formatVenueWeeklyEvents(events)

	assert.Contains(t, result, "**4/6(月)**")
	assert.Contains(t, result, "・**18:00開始** 月曜イベント")
	assert.Contains(t, result, "**4/8(水)**")
	assert.Contains(t, result, "・水曜イベント")
	// Tuesday (4/7) has no events and should not appear
	assert.NotContains(t, result, "4/7")
	// Date header must appear before events of that date
	assert.Less(t, strings.Index(result, "**4/6(月)**"), strings.Index(result, "**4/8(水)**"))
}

func TestFormatVenueWeeklyEvents_SortedByDateThenTime(t *testing.T) {
	svc := &EventNotificationService{}

	jst := time.FixedZone("JST", 9*60*60)
	day1 := time.Date(2026, 4, 6, 0, 0, 0, 0, jst) // Monday
	day2 := time.Date(2026, 4, 7, 0, 0, 0, 0, jst) // Tuesday

	events := []event.Event{
		{
			Title: "夜イベント",
			Date:  day1,
			Schedules: []event.Schedule{
				{StartTime: timePtr(time.Date(2026, 4, 6, 19, 0, 0, 0, jst))},
			},
		},
		{
			Title: "翌日イベント",
			Date:  day2,
		},
		{
			Title: "昼イベント",
			Date:  day1,
			Schedules: []event.Schedule{
				{StartTime: timePtr(time.Date(2026, 4, 6, 14, 0, 0, 0, jst))},
			},
		},
	}

	result := svc.formatVenueWeeklyEvents(events)

	day1Pos := strings.Index(result, "**4/6(月)**")
	day2Pos := strings.Index(result, "**4/7(火)**")
	lunchPos := strings.Index(result, "14:00開始")
	eveningPos := strings.Index(result, "19:00開始")

	assert.GreaterOrEqual(t, day1Pos, 0)
	assert.GreaterOrEqual(t, day2Pos, 0)
	assert.Less(t, day1Pos, day2Pos)
	assert.Less(t, lunchPos, eveningPos)
}

func TestFormatVenueWeeklyEvents_WeekdayLabels(t *testing.T) {
	svc := &EventNotificationService{}

	jst := time.FixedZone("JST", 9*60*60)
	// 2026-04-06 is Monday (月)
	events := []event.Event{
		{Title: "月", Date: time.Date(2026, 4, 6, 0, 0, 0, 0, jst)},
		{Title: "火", Date: time.Date(2026, 4, 7, 0, 0, 0, 0, jst)},
		{Title: "水", Date: time.Date(2026, 4, 8, 0, 0, 0, 0, jst)},
		{Title: "木", Date: time.Date(2026, 4, 9, 0, 0, 0, 0, jst)},
		{Title: "金", Date: time.Date(2026, 4, 10, 0, 0, 0, 0, jst)},
		{Title: "土", Date: time.Date(2026, 4, 11, 0, 0, 0, 0, jst)},
		{Title: "日", Date: time.Date(2026, 4, 12, 0, 0, 0, 0, jst)},
	}

	result := svc.formatVenueWeeklyEvents(events)

	assert.Contains(t, result, "4/6(月)")
	assert.Contains(t, result, "4/7(火)")
	assert.Contains(t, result, "4/8(水)")
	assert.Contains(t, result, "4/9(木)")
	assert.Contains(t, result, "4/10(金)")
	assert.Contains(t, result, "4/11(土)")
	assert.Contains(t, result, "4/12(日)")
}
