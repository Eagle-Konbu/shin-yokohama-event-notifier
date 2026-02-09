package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

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
			return errors.Join(
				fmt.Errorf("failed to fetch events: %w", err),
				fmt.Errorf("failed to send failure notification: %w", sendErr),
			)
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
		si := firstStartTime(events[i])
		sj := firstStartTime(events[j])
		if si == nil && sj == nil {
			return events[i].Title < events[j].Title
		}
		if si == nil {
			return false
		}
		if sj == nil {
			return true
		}
		return si.Before(*sj)
	})

	var lines []string
	for _, e := range events {
		lines = append(lines, formatEvent(e))
	}

	return strings.Join(lines, "\n")
}

func firstStartTime(e event.Event) *time.Time {
	if len(e.Schedules) > 0 {
		return e.Schedules[0].StartTime
	}
	return nil
}

func formatEvent(e event.Event) string {
	if len(e.Schedules) == 0 {
		return fmt.Sprintf("・%s", e.Title)
	}

	if len(e.Schedules) == 1 {
		return fmt.Sprintf("・**%s** %s", formatSchedule(e.Schedules[0]), e.Title)
	}

	var parts []string
	for i, slot := range e.Schedules {
		parts = append(parts, fmt.Sprintf("%s%s", circledNumber(i+1), formatSchedule(slot)))
	}
	return fmt.Sprintf("・**%s** %s", strings.Join(parts, " "), e.Title)
}

func formatSchedule(slot event.Schedule) string {
	switch {
	case slot.OpenTime != nil && slot.StartTime != nil:
		return fmt.Sprintf("%s開場 / %s開始", slot.OpenTime.Format("15:04"), slot.StartTime.Format("15:04"))
	case slot.StartTime != nil:
		return fmt.Sprintf("%s開始", slot.StartTime.Format("15:04"))
	case slot.OpenTime != nil:
		return fmt.Sprintf("%s開場", slot.OpenTime.Format("15:04"))
	default:
		return ""
	}
}

func circledNumber(n int) string {
	if n >= 1 && n <= 20 {
		return string(rune('\u2460' + n - 1))
	}
	return fmt.Sprintf("(%d)", n)
}
