package discord

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
)

func TestMapNotificationToEmbed_BasicFields(t *testing.T) {
	notif := notification.NewNotification(
		"Test Title",
		"Test Description",
		notification.ColorBlue,
	)

	embed := mapNotificationToEmbed(notif)

	assert.Equal(t, "Test Title", embed.Title)
	assert.Equal(t, "Test Description", embed.Description)
	assert.Equal(t, int(notification.ColorBlue), embed.Color)
	assert.NotEmpty(t, embed.Timestamp)
	assert.Empty(t, embed.Fields)
}

func TestMapNotificationToEmbed_WithFields(t *testing.T) {
	notif := notification.NewNotification("Title", "Description", notification.ColorGreen)
	notif.AddField("Field1", "Value1", true)
	notif.AddField("Field2", "Value2", false)
	notif.AddField("Field3", "Value3", true)

	embed := mapNotificationToEmbed(notif)

	require.Len(t, embed.Fields, 3)
	assert.Equal(t, "Field1", embed.Fields[0].Name)
	assert.Equal(t, "Value1", embed.Fields[0].Value)
	assert.True(t, embed.Fields[0].Inline)

	assert.Equal(t, "Field2", embed.Fields[1].Name)
	assert.Equal(t, "Value2", embed.Fields[1].Value)
	assert.False(t, embed.Fields[1].Inline)

	assert.Equal(t, "Field3", embed.Fields[2].Name)
	assert.Equal(t, "Value3", embed.Fields[2].Value)
	assert.True(t, embed.Fields[2].Inline)
}

func TestMapNotificationToEmbed_ColorConversion(t *testing.T) {
	tests := []struct {
		name        string
		color       notification.Color
		expectedInt int
	}{
		{"ColorBlue", notification.ColorBlue, 3447003},
		{"ColorGreen", notification.ColorGreen, 3066993},
		{"ColorYellow", notification.ColorYellow, 16776960},
		{"ColorRed", notification.ColorRed, 15158332},
		{"ColorPurple", notification.ColorPurple, 10181046},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notif := notification.NewNotification("Title", "Description", tt.color)
			embed := mapNotificationToEmbed(notif)
			assert.Equal(t, tt.expectedInt, embed.Color)
		})
	}
}

func TestMapNotificationToEmbed_TimestampFormat(t *testing.T) {
	notif := notification.NewNotification("Title", "Description", notification.ColorBlue)

	embed := mapNotificationToEmbed(notif)

	parsedTime, err := time.Parse(time.RFC3339, embed.Timestamp)
	require.NoError(t, err, "timestamp should be in RFC3339 format")

	assert.WithinDuration(t, time.Now().UTC(), parsedTime, 2*time.Second)
}

func TestMapNotificationToEmbed_EmptyFields(t *testing.T) {
	notif := notification.NewNotification("Title", "Description", notification.ColorRed)

	embed := mapNotificationToEmbed(notif)

	assert.NotNil(t, embed.Fields)
	assert.Empty(t, embed.Fields)
}

func TestMapNotificationToEmbed_FieldOrderingPreserved(t *testing.T) {
	notif := notification.NewNotification("Title", "Description", notification.ColorPurple)
	notif.AddField("First", "1", false)
	notif.AddField("Second", "2", false)
	notif.AddField("Third", "3", false)
	notif.AddField("Fourth", "4", false)

	embed := mapNotificationToEmbed(notif)

	require.Len(t, embed.Fields, 4)
	assert.Equal(t, "First", embed.Fields[0].Name)
	assert.Equal(t, "Second", embed.Fields[1].Name)
	assert.Equal(t, "Third", embed.Fields[2].Name)
	assert.Equal(t, "Fourth", embed.Fields[3].Name)
}

func TestEmbed_JSONMarshaling(t *testing.T) {
	embed := Embed{
		Title:       "Test Title",
		Description: "Test Description",
		Color:       3447003,
		Timestamp:   "2024-01-01T00:00:00.000Z",
		Fields: []EmbedField{
			{Name: "Field1", Value: "Value1", Inline: true},
			{Name: "Field2", Value: "Value2", Inline: false},
		},
	}

	jsonData, err := json.Marshal(embed)
	require.NoError(t, err)

	var unmarshaled Embed
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, embed.Title, unmarshaled.Title)
	assert.Equal(t, embed.Description, unmarshaled.Description)
	assert.Equal(t, embed.Color, unmarshaled.Color)
	assert.Equal(t, embed.Timestamp, unmarshaled.Timestamp)
	assert.Len(t, unmarshaled.Fields, 2)
}

func TestWebhookPayload_JSONMarshaling(t *testing.T) {
	payload := WebhookPayload{
		Content:  "Test Content",
		Username: "TestBot",
		Embeds: []Embed{
			{
				Title:       "Embed Title",
				Description: "Embed Description",
				Color:       3447003,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	var unmarshaled WebhookPayload
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, payload.Content, unmarshaled.Content)
	assert.Equal(t, payload.Username, unmarshaled.Username)
	assert.Len(t, unmarshaled.Embeds, 1)
	assert.Equal(t, "Embed Title", unmarshaled.Embeds[0].Title)
}

func TestWebhookPayload_OmitEmptyFields(t *testing.T) {
	payload := WebhookPayload{
		Embeds: []Embed{
			{Title: "Title Only"},
		},
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	jsonStr := string(jsonData)
	assert.NotContains(t, jsonStr, "content")
	assert.NotContains(t, jsonStr, "username")
	assert.Contains(t, jsonStr, "embeds")
}
