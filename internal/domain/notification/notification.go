package notification

import (
	"time"
)

type Notification struct {
	timestamp   time.Time
	title       string
	description string
	fields      []Field
	color       Color
}

type Field struct {
	Name   string
	Value  string
	Inline bool
}

// Color represents the notification color (Discord uses int for colors)
type Color int

const (
	ColorGreen  Color = 3066993
	ColorYellow Color = 16776960
	ColorRed    Color = 15158332
)

func NewNotification(title, description string, color Color) *Notification {
	return &Notification{
		title:       title,
		description: description,
		color:       color,
		timestamp:   time.Now(),
		fields:      make([]Field, 0),
	}
}

func (n *Notification) AddField(name, value string, inline bool) {
	n.fields = append(n.fields, Field{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
}

func (n *Notification) Title() string        { return n.title }
func (n *Notification) Description() string  { return n.description }
func (n *Notification) Fields() []Field      { return n.fields }
func (n *Notification) Color() Color         { return n.color }
func (n *Notification) Timestamp() time.Time { return n.timestamp }
