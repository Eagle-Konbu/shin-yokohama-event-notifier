package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) Send(ctx context.Context, notif *notification.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

type MockEventFetcher struct {
	mock.Mock
}

func (m *MockEventFetcher) FetchEvents(ctx context.Context) ([]event.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]event.Event), args.Error(1)
}

func TestNewEventNotificationService(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})

	require.NotNil(t, service)
	assert.NotNil(t, service.notificationSender)
	assert.Len(t, service.eventFetchers, 1)
}

func TestNotifyTodayEvents_NoEvents(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()

	mockFetcher.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)
	assert.Equal(t, "📅 新横浜 イベント情報", capturedNotif.Title())
	assert.Equal(t, "本日のイベント情報をお知らせします。", capturedNotif.Description())
	assert.Equal(t, notification.ColorGreen, capturedNotif.Color())
	assert.Len(t, capturedNotif.Fields(), 3)
	for _, field := range capturedNotif.Fields() {
		assert.Equal(t, "本日の予定はありません", field.Value)
	}
}

func TestNotifyTodayEvents_OneVenueWithEvents(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()

	events := []event.Event{
		{
			Title: "テストイベント",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
			Venue: event.VenueIDYokohamaArena,
		},
	}
	mockFetcher.On("FetchEvents", mock.Anything).Return(events, nil)

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)
	assert.Equal(t, notification.ColorYellow, capturedNotif.Color())

	arenaField := capturedNotif.Fields()[0]
	assert.Equal(t, "🏟️ 横浜アリーナ", arenaField.Name)
	assert.Contains(t, arenaField.Value, "・**18:00〜** テストイベント")
}

func TestNotifyTodayEvents_TwoVenuesWithEvents(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()

	events := []event.Event{
		{
			Title: "横浜アリーナイベント",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
			Venue: event.VenueIDYokohamaArena,
		},
		{
			Title: "日産スタジアムイベント",
			Date:  time.Date(2026, 1, 28, 14, 0, 0, 0, time.Local),
			Venue: event.VenueIDNissanStadium,
		},
	}
	mockFetcher.On("FetchEvents", mock.Anything).Return(events, nil)

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)
	assert.Equal(t, notification.ColorRed, capturedNotif.Color())
}

func TestNotifyTodayEvents_AllVenuesWithEvents(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()

	events := []event.Event{
		{
			Title: "横浜アリーナイベント",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
			Venue: event.VenueIDYokohamaArena,
		},
		{
			Title: "日産スタジアムイベント",
			Date:  time.Date(2026, 1, 28, 14, 0, 0, 0, time.Local),
			Venue: event.VenueIDNissanStadium,
		},
		{
			Title: "スケートセンターイベント",
			Date:  time.Date(2026, 1, 28, 10, 0, 0, 0, time.Local),
			Venue: event.VenueIDSkateCenter,
		},
	}
	mockFetcher.On("FetchEvents", mock.Anything).Return(events, nil)

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)
	assert.Equal(t, notification.ColorRed, capturedNotif.Color())
}

func TestNotifyTodayEvents_MultipleEventsAtSameVenue(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()

	events := []event.Event{
		{
			Title: "イベントB",
			Date:  time.Date(2026, 1, 28, 19, 0, 0, 0, time.Local),
			Venue: event.VenueIDYokohamaArena,
		},
		{
			Title: "イベントA",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
			Venue: event.VenueIDYokohamaArena,
		},
	}
	mockFetcher.On("FetchEvents", mock.Anything).Return(events, nil)

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)

	arenaField := capturedNotif.Fields()[0]
	assert.Contains(t, arenaField.Value, "・**18:00〜** イベントA\n・**19:00〜** イベントB")
}

func TestNotifyTodayEvents_VenueOrder(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()

	mockFetcher.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)
	require.Len(t, capturedNotif.Fields(), 3)
	assert.Equal(t, "🏟️ 横浜アリーナ", capturedNotif.Fields()[0].Name)
	assert.Equal(t, "⚽ 日産スタジアム", capturedNotif.Fields()[1].Name)
	assert.Equal(t, "⛸️ KOSÉ新横浜スケートセンター", capturedNotif.Fields()[2].Name)
}

func TestNotifyTodayEvents_FetchError(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()
	expectedErr := errors.New("fetch error")

	mockFetcher.On("FetchEvents", mock.Anything).Return(nil, expectedErr)

	err := service.NotifyTodayEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch events")
	assert.ErrorIs(t, err, expectedErr)
}

func TestNotifyTodayEvents_SendError(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := new(MockEventFetcher)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	ctx := context.Background()
	expectedErr := errors.New("send error")

	mockFetcher.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockSender.On("Send", ctx, mock.Anything).Return(expectedErr)

	err := service.NotifyTodayEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send notification")
	assert.ErrorIs(t, err, expectedErr)
}
