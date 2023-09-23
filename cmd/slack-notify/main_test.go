package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/api/calendar/v3"
)

func TestCreateSlackMessage(t *testing.T) {
	tempDir := t.TempDir()
	f, err := os.Create(filepath.Join(tempDir, "template.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString("{{.Msg}}\n\n{{range .Events -}}- {{.Summary}}\n{{end}}")

	var tests = []struct {
		name         string
		events       []*calendar.Event
		msg          string
		alt          string
		templateFile string
		want         string
		err          error
	}{
		{
			name: "normal",
			events: []*calendar.Event{
				{Summary: "Summary1"},
				{Summary: "Summary2"},
			},
			msg:          "Test Message",
			alt:          "Alternative Message",
			templateFile: "",
			want:         "Test Message\n\n• Summary1\n• Summary2\n",
			err:          nil,
		},
		{
			name:         "empty events",
			events:       []*calendar.Event{},
			msg:          "Test Message",
			alt:          "Alternative Message",
			templateFile: "",
			want:         "Alternative Message",
			err:          nil,
		},
		{
			name: "templateFile",
			events: []*calendar.Event{
				{Summary: "Summary1"},
				{Summary: "Summary2"},
			},
			msg:          "Test Message",
			alt:          "Alternative Message",
			templateFile: filepath.Join(tempDir, "template.txt"),
			want:         "Test Message\n\n- Summary1\n- Summary2\n",
			err:          nil,
		},
	}

	for _, tt := range tests {
		got, err := createSlackMessage(tt.events, tt.msg, tt.alt, tt.templateFile)
		if got != tt.want {
			t.Errorf("Want '%s', but got '%s'", tt.want, got)
		}
		if !errors.Is(err, tt.err) {
			t.Errorf("Want '%s', but got '%s'", tt.err, err)
		}
	}
}
