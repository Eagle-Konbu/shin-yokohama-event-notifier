package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/discord"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/scraper"
)

func main() {
	sendFlag := flag.Bool("send", false, "Send notification to Discord (requires DISCORD_WEBHOOK_URL)")
	flag.Parse()

	ctx := context.Background()

	fetchers := []ports.EventFetcher{
		scraper.NewYokohamaArenaScraper(),
		scraper.NewNissanStadiumScraper(),
		scraper.NewSkateCenterScraper(),
	}

	venues := event.NewAllVenues()
	if err := fetchAllEvents(ctx, fetchers, venues); err != nil {
		log.Fatalf("Failed to fetch events: %v", err)
	}

	printEvents(venues)

	if *sendFlag {
		webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
		if webhookURL == "" {
			log.Fatal("DISCORD_WEBHOOK_URL environment variable is required when using --send")
		}

		discordSender := discord.NewWebhookAdapter(webhookURL)
		eventService := service.NewEventNotificationService(discordSender, fetchers)

		if err := eventService.NotifyTodayEvents(ctx); err != nil {
			log.Fatalf("Failed to send notification: %v", err)
		}

		fmt.Println("\n✅ Notification sent to Discord")
	}
}

func fetchAllEvents(ctx context.Context, fetchers []ports.EventFetcher, venues []*event.Venue) error {
	venueMap := make(map[event.VenueID]*event.Venue)
	for _, v := range venues {
		venueMap[v.ID] = v
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, fetcher := range fetchers {
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

	return eg.Wait()
}

func printEvents(venues []*event.Venue) {
	fmt.Println("========================================")
	fmt.Println("📅 新横浜 イベント情報")
	fmt.Println("========================================")

	for _, venue := range venues {
		fmt.Printf("\n%s %s\n", venue.Emoji, venue.DisplayName)
		fmt.Println(strings.Repeat("-", 40))

		if len(venue.Events) == 0 {
			fmt.Println("  本日の予定はありません")
			continue
		}

		sort.Slice(venue.Events, func(i, j int) bool {
			return venue.Events[i].Date.Before(venue.Events[j].Date)
		})

		for _, e := range venue.Events {
			fmt.Printf("  %s〜 %s\n", e.Date.Format("15:04"), e.Title)
		}
	}

	fmt.Println()
}
