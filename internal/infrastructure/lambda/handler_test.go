package lambda

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
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

func (m *MockEventFetcher) FetchEvents(ctx context.Context, date time.Time) ([]event.Event, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]event.Event), args.Error(1)
}

func (m *MockEventFetcher) VenueID() event.VenueID {
	return m.venueID
}

func TestNewDailyHandler(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.eventService)
}

func TestDailyHandler_HandleRequest_Success(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	ctx := context.Background()

	mockFetcher.On("FetchEvents", mock.Anything, mock.Anything).Return([]event.Event{}, nil)
	mockSender.On("Send", ctx, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	mockSender.AssertExpectations(t)
	mockFetcher.AssertExpectations(t)
}

func TestDailyHandler_HandleRequest_ServiceError(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	ctx := context.Background()
	expectedErr := errors.New("fetch error")

	mockFetcher.On("FetchEvents", mock.Anything, mock.Anything).Return(nil, expectedErr)
	mockSender.On("Send", ctx, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to notify today events")
	assert.ErrorIs(t, err, expectedErr)
	mockFetcher.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestDailyHandler_HandleRequest_ContextPropagation(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")

	var capturedCtx context.Context
	mockFetcher.On("FetchEvents", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedCtx = args.Get(0).(context.Context)
	}).Return([]event.Event{}, nil)
	mockSender.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}

func TestNewWeeklyHandler(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.eventService)
}

func TestWeeklyHandler_HandleRequest_Success(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	ctx := context.Background()

	// NotifyWeeklyEvents fetches events for 7 days
	mockFetcher.On("FetchEvents", mock.Anything, mock.Anything).Return([]event.Event{}, nil).Times(7)
	mockSender.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	mockSender.AssertExpectations(t)
	mockFetcher.AssertExpectations(t)
}

func TestWeeklyHandler_HandleRequest_ServiceError(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	ctx := context.Background()
	expectedErr := errors.New("fetch error")

	mockFetcher.On("FetchEvents", mock.Anything, mock.Anything).Return(nil, expectedErr)
	mockSender.On("Send", ctx, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to notify weekly events")
	assert.ErrorIs(t, err, expectedErr)
	mockFetcher.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestWeeklyHandler_HandleRequest_ContextPropagation(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")

	var capturedCtx context.Context
	mockFetcher.On("FetchEvents", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedCtx = args.Get(0).(context.Context)
	}).Return([]event.Event{}, nil).Times(7)
	mockSender.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}
