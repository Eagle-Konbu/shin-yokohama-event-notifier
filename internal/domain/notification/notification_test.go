package notification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotification(t *testing.T) {
	title := "Test Title"
	description := "Test Description"
	color := ColorBlue

	notif := NewNotification(title, description, color)

	require.NotNil(t, notif)
	assert.Equal(t, title, notif.Title())
	assert.Equal(t, description, notif.Description())
	assert.Equal(t, color, notif.Color())
	assert.NotZero(t, notif.Timestamp())
	assert.Empty(t, notif.Fields())
	assert.WithinDuration(t, time.Now(), notif.Timestamp(), time.Second)
}

func TestNotification_AddField(t *testing.T) {
	notif := NewNotification("Title", "Description", ColorBlue)

	t.Run("add single field", func(t *testing.T) {
		notif.AddField("Field1", "Value1", true)
		fields := notif.Fields()
		require.Len(t, fields, 1)
		assert.Equal(t, "Field1", fields[0].Name)
		assert.Equal(t, "Value1", fields[0].Value)
		assert.True(t, fields[0].Inline)
	})

	t.Run("add multiple fields in order", func(t *testing.T) {
		notif := NewNotification("Title", "Description", ColorGreen)
		notif.AddField("First", "Value1", true)
		notif.AddField("Second", "Value2", false)
		notif.AddField("Third", "Value3", true)

		fields := notif.Fields()
		require.Len(t, fields, 3)
		assert.Equal(t, "First", fields[0].Name)
		assert.Equal(t, "Second", fields[1].Name)
		assert.Equal(t, "Third", fields[2].Name)
		assert.True(t, fields[0].Inline)
		assert.False(t, fields[1].Inline)
		assert.True(t, fields[2].Inline)
	})
}

func TestNotification_Getters(t *testing.T) {
	title := "Test Title"
	description := "Test Description"
	color := ColorPurple
	notif := NewNotification(title, description, color)
	notif.AddField("Field1", "Value1", true)

	t.Run("returns correct title", func(t *testing.T) {
		assert.Equal(t, title, notif.Title())
	})

	t.Run("returns correct description", func(t *testing.T) {
		assert.Equal(t, description, notif.Description())
	})

	t.Run("returns correct color", func(t *testing.T) {
		assert.Equal(t, color, notif.Color())
	})

	t.Run("returns correct timestamp", func(t *testing.T) {
		assert.WithinDuration(t, time.Now(), notif.Timestamp(), time.Second)
	})

	t.Run("returns correct fields", func(t *testing.T) {
		fields := notif.Fields()
		require.Len(t, fields, 1)
		assert.Equal(t, "Field1", fields[0].Name)
	})
}

func TestColor_PredefinedValues(t *testing.T) {
	tests := []struct {
		name     string
		color    Color
		expected int
	}{
		{"ColorBlue", ColorBlue, 3447003},
		{"ColorGreen", ColorGreen, 3066993},
		{"ColorYellow", ColorYellow, 16776960},
		{"ColorRed", ColorRed, 15158332},
		{"ColorPurple", ColorPurple, 10181046},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, int(tt.color))
		})
	}
}

func TestField(t *testing.T) {
	field := Field{
		Name:   "Test Field",
		Value:  "Test Value",
		Inline: true,
	}

	assert.Equal(t, "Test Field", field.Name)
	assert.Equal(t, "Test Value", field.Value)
	assert.True(t, field.Inline)
}
