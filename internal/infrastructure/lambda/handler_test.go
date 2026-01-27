package lambda

import (
	"context"
	"errors"
	"testing"

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

func TestNewHandler(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewHandler(svc)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.eventService)
}

func TestHandler_HandleRequest_Success(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewHandler(svc)

	ctx := context.Background()

	mockFetcher.On("FetchEvents", mock.Anything).Return([]event.Event{}, nil)
	mockSender.On("Send", ctx, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	mockSender.AssertExpectations(t)
	mockFetcher.AssertExpectations(t)
}

func TestHandler_HandleRequest_ServiceError(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewHandler(svc)

	ctx := context.Background()
	expectedErr := errors.New("fetch error")

	mockFetcher.On("FetchEvents", mock.Anything).Return(nil, expectedErr)
	mockSender.On("Send", ctx, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to notify today events")
	assert.ErrorIs(t, err, expectedErr)
	mockFetcher.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestHandler_HandleRequest_ContextPropagation(t *testing.T) {
	mockSender := new(MockNotificationSender)
	mockFetcher := NewMockEventFetcher(event.VenueIDYokohamaArena)
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewHandler(svc)

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")

	var capturedCtx context.Context
	mockFetcher.On("FetchEvents", mock.Anything).Run(func(args mock.Arguments) {
		capturedCtx = args.Get(0).(context.Context)
	}).Return([]event.Event{}, nil)
	mockSender.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}
