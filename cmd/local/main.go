package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/briandowns/spinner"

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
	venueMap := make(map[event.VenueID]*event.Venue)
	for _, v := range venues {
		venueMap[v.ID] = v
	}

	var hasError bool

	for _, fetcher := range fetchers {
		venue := venueMap[fetcher.VenueID()]

		s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
		s.Prefix = fmt.Sprintf("Fetching %s", venue.DisplayName)
		s.Start()
		events, err := fetcher.FetchEvents(ctx)
		s.Stop()

		if err != nil {
			fmt.Printf("[%s]\n", venue.DisplayName)
			fmt.Printf("  error: %v\n\n", err)
			hasError = true
			continue
		}

		venue.Events = events
		printVenue(venue)
	}

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

		fmt.Println("Notification sent to Discord")
	}

	if hasError {
		os.Exit(1)
	}
}

func printVenue(venue *event.Venue) {
	fmt.Printf("[%s]\n", venue.DisplayName)

	if len(venue.Events) == 0 {
		fmt.Println("  (none)")
		fmt.Println()
		return
	}

	sort.Slice(venue.Events, func(i, j int) bool {
		return venue.Events[i].Date.Before(venue.Events[j].Date)
	})

	for _, e := range venue.Events {
		fmt.Printf("  %s %s\n", e.Date.Format("15:04"), e.Title)
	}
	fmt.Println()
}
