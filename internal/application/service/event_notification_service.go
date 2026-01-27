package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type EventNotificationService struct {
	notificationSender ports.NotificationSender
	eventFetchers      []ports.EventFetcher
}

func NewEventNotificationService(sender ports.NotificationSender, fetchers []ports.EventFetcher) *EventNotificationService {
	return &EventNotificationService{
		notificationSender: sender,
		eventFetchers:      fetchers,
	}
}

var venueEmojis = map[event.Venue]string{
	event.VenueYokohamaArena: "🏟️",
	event.VenueNissanStadium: "⚽",
	event.VenueSkateCenter:   "⛸️",
}

var venueOrder = []event.Venue{
	event.VenueYokohamaArena,
	event.VenueNissanStadium,
	event.VenueSkateCenter,
}

func (s *EventNotificationService) NotifyTodayEvents(ctx context.Context) error {
	eventsByVenue, err := s.fetchAllEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	notif := s.buildNotification(eventsByVenue)

	if err := s.notificationSender.Send(ctx, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *EventNotificationService) fetchAllEvents(ctx context.Context) (map[event.Venue][]event.Event, error) {
	var mu sync.Mutex
	eventsByVenue := make(map[event.Venue][]event.Event)

	eg, ctx := errgroup.WithContext(ctx)
	for _, fetcher := range s.eventFetchers {
		eg.Go(func() error {
			events, err := fetcher.FetchEvents(ctx)
			if err != nil {
				return err
			}

			mu.Lock()
			for _, e := range events {
				eventsByVenue[e.Venue] = append(eventsByVenue[e.Venue], e)
			}
			mu.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return eventsByVenue, nil
}

func (s *EventNotificationService) buildNotification(eventsByVenue map[event.Venue][]event.Event) *notification.Notification {
	color := s.determineColor(eventsByVenue)
	notif := notification.NewNotification(
		"📅 新横浜 イベント情報",
		"本日のイベント情報をお知らせします。",
		color,
	)

	for _, venue := range venueOrder {
		fieldName := fmt.Sprintf("%s %s", venueEmojis[venue], string(venue))
		fieldValue := s.formatVenueEvents(eventsByVenue[venue])
		notif.AddField(fieldName, fieldValue, false)
	}

	return notif
}

func (s *EventNotificationService) determineColor(eventsByVenue map[event.Venue][]event.Event) notification.Color {
	venuesWithEvents := 0
	for _, events := range eventsByVenue {
		if len(events) > 0 {
			venuesWithEvents++
		}
	}

	switch venuesWithEvents {
	case 0:
		return notification.ColorBlue
	case 1:
		return notification.ColorYellow
	default:
		return notification.ColorRed
	}
}

func (s *EventNotificationService) formatVenueEvents(events []event.Event) string {
	if len(events) == 0 {
		return "本日の予定はありません"
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Date.Before(events[j].Date)
	})

	var lines []string
	for _, e := range events {
		line := fmt.Sprintf("・**%s〜** %s", e.Date.Format("15:04"), e.Title)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
