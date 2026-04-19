package lambda

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports/mock_ports"
)

func TestNewDailyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.eventService)
}

func TestDailyHandler_HandleRequest_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handler.HandleRequest(context.Background())

	require.NoError(t, err)
}

func TestDailyHandler_HandleRequest_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	expectedErr := errors.New("fetch error")

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handler.HandleRequest(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to notify today events")
	assert.ErrorIs(t, err, expectedErr)
}

func TestDailyHandler_HandleRequest_ContextPropagation(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewDailyHandler(svc)

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")

	var capturedCtx context.Context
	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, from, to interface{}) ([]event.Event, error) {
		capturedCtx = ctx
		return []event.Event{}, nil
	})
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}

func TestNewWeeklyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.eventService)
}

func TestWeeklyHandler_HandleRequest_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return([]event.Event{}, nil)
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handler.HandleRequest(context.Background())

	require.NoError(t, err)
}

func TestWeeklyHandler_HandleRequest_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	expectedErr := errors.New("fetch error")

	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handler.HandleRequest(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to notify weekly events")
	assert.ErrorIs(t, err, expectedErr)
}

func TestWeeklyHandler_HandleRequest_ContextPropagation(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSender := mock_ports.NewMockNotificationSender(ctrl)
	mockFetcher := mock_ports.NewMockEventFetcher(ctrl)
	mockFetcher.EXPECT().VenueID().Return(event.VenueIDYokohamaArena).AnyTimes()
	svc := service.NewEventNotificationService(mockSender, []ports.EventFetcher{mockFetcher})
	handler := NewWeeklyHandler(svc)

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")

	var capturedCtx context.Context
	mockFetcher.EXPECT().FetchEvents(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, from, to interface{}) ([]event.Event, error) {
		capturedCtx = ctx
		return []event.Event{}, nil
	})
	mockSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handler.HandleRequest(ctx)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}
