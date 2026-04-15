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
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, jst)

	venues := event.NewAllVenues()

	if err := s.fetchAllEvents(ctx, venues, today); err != nil {
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

	notif := s.buildDailyNotification(venues)

	if err := s.notificationSender.Send(ctx, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *EventNotificationService) NotifyWeeklyEvents(ctx context.Context) error {
	jst := time.FixedZone("JST", 9*60*60)
	today := time.Now().In(jst)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, jst)

	venues := event.NewAllVenues()

	for i := range 7 {
		date := today.AddDate(0, 0, i)
		if err := s.fetchAllEvents(ctx, venues, date); err != nil {
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
	}

	notif := s.buildWeeklyNotification(venues, today)

	if err := s.notificationSender.Send(ctx, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *EventNotificationService) fetchAllEvents(ctx context.Context, venues []*event.Venue, date time.Time) error {
	venueMap := make(map[event.VenueID]*event.Venue)
	for _, v := range venues {
		venueMap[v.ID] = v
	}

	type fetchResult struct {
		venueID event.VenueID
		events  []event.Event
	}
	results := make([]fetchResult, len(s.eventFetchers))

	eg, ctx := errgroup.WithContext(ctx)
	for i, fetcher := range s.eventFetchers {
		eg.Go(func() error {
			events, err := fetcher.FetchEvents(ctx, date)
			if err != nil {
				return err
			}
			results[i] = fetchResult{venueID: fetcher.VenueID(), events: events}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	for _, r := range results {
		if venue, ok := venueMap[r.venueID]; ok {
			venue.Events = append(venue.Events, r.events...)
		}
	}

	return nil
}

func (s *EventNotificationService) buildDailyNotification(venues []*event.Venue) *notification.Notification {
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

func (s *EventNotificationService) buildWeeklyNotification(venues []*event.Venue, startDate time.Time) *notification.Notification {
	color := s.determineColor(venues)
	endDate := startDate.AddDate(0, 0, 6)
	description := fmt.Sprintf(
		"%s 〜 %s のイベント情報をお知らせします。",
		startDate.Format("1/2"),
		endDate.Format("1/2"),
	)
	notif := notification.NewNotification(
		"📅 新横浜 週間イベント情報",
		description,
		color,
	)

	for _, venue := range venues {
		fieldName := fmt.Sprintf("%s %s", venue.Emoji, venue.DisplayName)
		fieldValue := s.formatVenueWeeklyEvents(venue.Events)
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

func (s *EventNotificationService) formatVenueWeeklyEvents(events []event.Event) string {
	if len(events) == 0 {
		return "今週の予定はありません"
	}

	// Sort by date, then by start time within each date, then by title
	sort.Slice(events, func(i, j int) bool {
		iy, im, id := events[i].Date.Date()
		jy, jm, jd := events[j].Date.Date()
		if iy != jy || im != jm || id != jd {
			di := time.Date(iy, im, id, 0, 0, 0, 0, events[i].Date.Location())
			dj := time.Date(jy, jm, jd, 0, 0, 0, 0, events[j].Date.Location())
			return di.Before(dj)
		}
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
		if !si.Equal(*sj) {
			return si.Before(*sj)
		}
		return events[i].Title < events[j].Title
	})

	// Group events by date (compare year/month/day in the event's timezone)
	var lines []string
	var curY, curD int
	var curM time.Month
	for _, e := range events {
		ey, em, ed := e.Date.Date()
		if ey != curY || em != curM || ed != curD {
			curY, curM, curD = ey, em, ed
			lines = append(lines, fmt.Sprintf("**%s**", formatDateLabel(e.Date)))
		}
		lines = append(lines, formatEvent(e))
	}

	return strings.Join(lines, "\n")
}

var weekdayJP = [7]string{"日", "月", "火", "水", "木", "金", "土"}

func formatDateLabel(t time.Time) string {
	return fmt.Sprintf("%d/%d(%s)", t.Month(), t.Day(), weekdayJP[t.Weekday()])
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
