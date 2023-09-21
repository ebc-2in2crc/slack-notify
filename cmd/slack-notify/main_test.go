package main

import (
	"testing"

	"google.golang.org/api/calendar/v3"
)

func TestCreateSlackMessage(t *testing.T) {
	var tests = []struct {
		name   string
		events []*calendar.Event
		msg    string
		alt    string
		want   string
	}{
		{
			name: "normal",
			events: []*calendar.Event{
				{Summary: "Summary1"},
				{Summary: "Summary2"},
			},
			msg:  "Test Message",
			alt:  "Alternative Message",
			want: "Test Message\n\n• Summary1\n• Summary2",
		},
		{
			name:   "empty events",
			events: []*calendar.Event{},
			msg:    "Test Message",
			alt:    "Alternative Message",
			want:   "Alternative Message",
		},
	}

	for _, tt := range tests {
		got := createSlackMessage(tt.events, tt.msg, tt.alt)
		if got != tt.want {
			t.Errorf("Want '%s', but got '%s'", tt.want, got)
		}
	}
}
