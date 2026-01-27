package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

func (s *EventNotificationService) NotifyTodayEvents(ctx context.Context) error {
	venues := event.NewAllVenues()

	if err := s.fetchAllEvents(ctx, venues); err != nil {
		failureNotif := notification.NewNotification(
			"❌ イベント取得エラー",
			"イベント情報の取得に失敗しました",
			notification.ColorRed,
		)
		if sendErr := s.notificationSender.Send(ctx, failureNotif); sendErr != nil {
			return fmt.Errorf("failed to fetch events: %w (failed to send failure notification: %v)", err, sendErr)
		}
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	notif := s.buildNotification(venues)

	if err := s.notificationSender.Send(ctx, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *EventNotificationService) fetchAllEvents(ctx context.Context, venues []*event.Venue) error {
	venueMap := make(map[event.VenueID]*event.Venue)
	for _, v := range venues {
		venueMap[v.ID] = v
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, fetcher := range s.eventFetchers {
		eg.Go(func() error {
			events, err := fetcher.FetchEvents(ctx)
			if err != nil {
				return err
			}

			if venue, ok := venueMap[fetcher.VenueID()]; ok {
				venue.Events = append(venue.Events, events...)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *EventNotificationService) buildNotification(venues []*event.Venue) *notification.Notification {
	color := s.determineColor(venues)
	notif := notification.NewNotification(
		"📅 新横浜 イベント情報",
		"本日のイベント情報をお知らせします。",
		color,
	)

	for _, venue := range venues {
		fieldName := fmt.Sprintf("%s %s", venue.Emoji, venue.DisplayName)
		fieldValue := s.formatVenueEvents(venue.Events)
		notif.AddField(fieldName, fieldValue, false)
	}

	return notif
}

func (s *EventNotificationService) determineColor(venues []*event.Venue) notification.Color {
	venuesWithEvents := 0
	for _, venue := range venues {
		if len(venue.Events) > 0 {
			venuesWithEvents++
		}
	}

	switch venuesWithEvents {
	case 0:
		return notification.ColorGreen
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
