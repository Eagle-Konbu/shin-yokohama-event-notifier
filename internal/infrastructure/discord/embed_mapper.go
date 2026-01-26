package discord

import (
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
)

type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type WebhookPayload struct {
	Content  string  `json:"content,omitempty"`
	Username string  `json:"username,omitempty"`
	Embeds   []Embed `json:"embeds,omitempty"`
}

func mapNotificationToEmbed(notif *notification.Notification) Embed {
	embed := Embed{
		Title:       notif.Title(),
		Description: notif.Description(),
		Color:       int(notif.Color()),
		Timestamp:   notif.Timestamp().UTC().Format("2006-01-02T15:04:05.000Z"),
		Fields:      make([]EmbedField, 0, len(notif.Fields())),
	}

	for _, field := range notif.Fields() {
		embed.Fields = append(embed.Fields, EmbedField{
			Name:   field.Name,
			Value:  field.Value,
			Inline: field.Inline,
		})
	}

	return embed
}
