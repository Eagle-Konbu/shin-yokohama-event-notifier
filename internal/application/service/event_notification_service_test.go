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
	venueID event.VenueID
}

func NewMockEventFetcher(venueID event.VenueID) *MockEventFetcher {
	return &MockEventFetcher{venueID: venueID}
}

func (m *MockEventFetcher) FetchEvents(ctx context.Context) ([]event.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]event.Event), args.Error(1)
}

func (m *MockEventFetcher) VenueID() event.VenueID {
	return m.venueID
}

func setupSingleFetcherService() (*MockNotificationSender, *MockEventFetcher, *EventNotificationService, context.Context) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	return mockSender, mockFetcher, service, context.Background()
}

func setupThreeFetcherService() (*MockNotificationSender, *MockEventFetcher, *MockEventFetcher, *MockEventFetcher, *EventNotificationService, context.Context) {
	mockSender := new(MockNotificationSender)
	mockFetcher1 := NewMockEventFetcher(event.VenueIDYokohamaArena)
	mockFetcher2 := NewMockEventFetcher(event.VenueIDNissanStadium)
	mockFetcher3 := NewMockEventFetcher(event.VenueIDSkateCenter)
	service := NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher1, mockFetcher2, mockFetcher3})
	return mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, context.Background()
}

func TestNewEventNotificationService(t *testing.T) {
	_, _, service, _ := setupSingleFetcherService()

	require.NotNil(t, service)
	assert.NotNil(t, service.notificationSender)
	assert.Len(t, service.eventFetchers, 1)
}

func TestNotifyTodayEvents_NoEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService()

	mockFetcher1.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockFetcher2.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockFetcher3.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		sentNotification = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, "📅 新横浜 イベント情報", sentNotification.Title())
	assert.Equal(t, "本日のイベント情報をお知らせします。", sentNotification.Description())
	assert.Equal(t, notification.ColorGreen, sentNotification.Color())
	assert.Len(t, sentNotification.Fields(), 3)
	for _, field := range sentNotification.Fields() {
		assert.Equal(t, "本日の予定はありません", field.Value)
	}
}

func TestNotifyTodayEvents_OneVenueWithEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService()

	events := []event.Event{
		{
			Title: "テストイベント",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
		},
	}
	mockFetcher1.On("FetchEvents", mock.Anything).Return(events, nil)
	mockFetcher2.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockFetcher3.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		sentNotification = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, notification.ColorYellow, sentNotification.Color())

	arenaField := sentNotification.Fields()[0]
	assert.Equal(t, "🏟️ 横浜アリーナ", arenaField.Name)
	assert.Contains(t, arenaField.Value, "・**18:00〜** テストイベント")
}

func TestNotifyTodayEvents_TwoVenuesWithEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService()

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
	mockFetcher1.On("FetchEvents", mock.Anything).Return(arenaEvents, nil)
	mockFetcher2.On("FetchEvents", mock.Anything).Return(stadiumEvents, nil)
	mockFetcher3.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		sentNotification = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, notification.ColorRed, sentNotification.Color())
}

func TestNotifyTodayEvents_AllVenuesWithEvents(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService()

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
	mockFetcher1.On("FetchEvents", mock.Anything).Return(arenaEvents, nil)
	mockFetcher2.On("FetchEvents", mock.Anything).Return(stadiumEvents, nil)
	mockFetcher3.On("FetchEvents", mock.Anything).Return(skateEvents, nil)

	var sentNotification *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		sentNotification = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	assert.Equal(t, notification.ColorRed, sentNotification.Color())
}

func TestNotifyTodayEvents_MultipleEventsAtSameVenue(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService()

	events := []event.Event{
		{
			Title: "イベントB",
			Date:  time.Date(2026, 1, 28, 19, 0, 0, 0, time.Local),
		},
		{
			Title: "イベントA",
			Date:  time.Date(2026, 1, 28, 18, 0, 0, 0, time.Local),
		},
	}
	mockFetcher1.On("FetchEvents", mock.Anything).Return(events, nil)
	mockFetcher2.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockFetcher3.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		sentNotification = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)

	arenaField := sentNotification.Fields()[0]
	assert.Contains(t, arenaField.Value, "・**18:00〜** イベントA\n・**19:00〜** イベントB")
}

func TestNotifyTodayEvents_VenueOrder(t *testing.T) {
	mockSender, mockFetcher1, mockFetcher2, mockFetcher3, service, ctx := setupThreeFetcherService()

	mockFetcher1.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockFetcher2.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockFetcher3.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)

	var sentNotification *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		sentNotification = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.NotifyTodayEvents(ctx)

	require.NoError(t, err)
	require.NotNil(t, sentNotification)
	require.Len(t, sentNotification.Fields(), 3)
	assert.Equal(t, "🏟️ 横浜アリーナ", sentNotification.Fields()[0].Name)
	assert.Equal(t, "⚽ 日産スタジアム", sentNotification.Fields()[1].Name)
	assert.Equal(t, "⛸️ KOSÉ新横浜スケートセンター", sentNotification.Fields()[2].Name)
}

func TestNotifyTodayEvents_FetchError(t *testing.T) {
	mockSender, mockFetcher, service, ctx := setupSingleFetcherService()
	expectedErr := errors.New("fetch error")

	mockFetcher.On("FetchEvents", mock.Anything).Return(nil, expectedErr)

	var sentNotification *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		sentNotification = args.Get(1).(*notification.Notification)
	}).Return(nil)

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
	mockSender, mockFetcher, service, ctx := setupSingleFetcherService()
	fetchErr := errors.New("fetch error")
	sendErr := errors.New("send error")

	mockFetcher.On("FetchEvents", mock.Anything).Return(nil, fetchErr)
	mockSender.On("Send", ctx, mock.Anything).Return(sendErr)

	err := service.NotifyTodayEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch events")
	assert.Contains(t, err.Error(), "failed to send failure notification")
	assert.ErrorIs(t, err, fetchErr)
}

func TestNotifyTodayEvents_SendError(t *testing.T) {
	mockSender, mockFetcher, service, ctx := setupSingleFetcherService()
	expectedErr := errors.New("send error")

	mockFetcher.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockSender.On("Send", ctx, mock.Anything).Return(expectedErr)

	err := service.NotifyTodayEvents(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send notification")
	assert.ErrorIs(t, err, expectedErr)
}
