package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"google.golang.org/api/calendar/v3"

	"golang.org/x/net/context"
)

var version = "0.0.1"

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)
}

type opt struct {
	alternativeMessage string
	calendarID         string
	credentials        string
	credentialsFile    string
	dryRun             bool
	eventFilterRegexp  string
	location           string
	slackAccessToken   string
	slackChannelID     string
	message            string
	targetDate         string
	timeout            time.Duration
	version            bool
	webhook            string
}

type client struct {
	fetcher *eventFetcher
	poster  *slackPoster
}

func main() {
	opt, err := parseFlag()
	if err != nil {
		logger.Fatalf("failed to parse flag: %v", err)
	}
	if opt.version {
		_, _ = fmt.Fprintf(os.Stderr, "slack-notify version %s\n", version)
		return
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), opt.timeout)
	defer cancelFunc()

	// client を作成する
	c, err := newClient(opt)
	if err != nil {
		logger.Fatalf("failed to create client: %v", err)
	}

	// イベントを取得する
	events, err := c.fetcher.fetch(ctx)
	if err != nil {
		logger.Fatalf("failed to fetch: %v", err)
	}

	// Slack に投稿するメッセージを作成する
	msg, err := createSlackMessage(events, opt.message, opt.alternativeMessage)
	if err != nil {
		logger.Fatalf("failed to create slack message: %v", err)
	}

	// Slack に投稿する
	if err := c.poster.post(ctx, msg); err != nil {
		logger.Fatalf("failed to post: %v", err)
	}
}

func parseFlag() (*opt, error) {
	alternativeMessage := flag.String("alternative-message", "", "Specify alternative message")
	calendarID := flag.String("calendar-id", "", "Specify Google Calendar ID")
	credentials := flag.String("credentials", "", "Specify credentials")
	credentialsFile := flag.String("credentials-file", "", "Specify credentials file")
	dryRun := flag.Bool("dry-run", false, "Specify dry-run")
	eventFilterRegexp := flag.String("event-filter-regexp", ".", "Specify event filter regexp")
	location := flag.String("location", "UTC", "Specify Location")
	message := flag.String("message", "", "Specify message")
	slackAccessToken := flag.String("slack-token", "", "Specify Slack Access Token")
	slackChannelID := flag.String("slack-channel-id", "", "Specify Slack Channel ID")
	targetDate := flag.String("target-date", "", "Specify targetDate date. e.g. 2020-01-01")
	timeoutOption := flag.Duration("timeout", 15*time.Minute, "Specify timeout")
	version := flag.Bool("v", false, "Show version")
	webhookOption := flag.String("webhook", "", "Specify Slack Webhook URL")
	flag.Parse()

	if *version {
		return &opt{version: *version}, nil
	}

	if *credentials == "" && *credentialsFile == "" {
		return nil, fmt.Errorf("credentials or credentials-file must be specified")
	}
	if *calendarID == "" {
		return nil, fmt.Errorf("calendar-id must be specified")
	}
	if *slackAccessToken == "" && *webhookOption == "" {
		return nil, fmt.Errorf("slack-token or webhook must be specified")
	}
	if *slackChannelID == "" && *webhookOption == "" {
		return nil, fmt.Errorf("slack-channel-id or webhook must be specified")
	}

	return &opt{
		alternativeMessage: *alternativeMessage,
		calendarID:         *calendarID,
		credentials:        *credentials,
		credentialsFile:    *credentialsFile,
		dryRun:             *dryRun,
		eventFilterRegexp:  *eventFilterRegexp,
		location:           *location,
		slackAccessToken:   *slackAccessToken,
		slackChannelID:     *slackChannelID,
		message:            *message,
		targetDate:         *targetDate,
		timeout:            *timeoutOption,
		version:            *version,
		webhook:            *webhookOption,
	}, nil
}

func newClient(opt *opt) (*client, error) {
	ef, err := newEventFetcher(&eventFetcherOpt{
		calendarID:        opt.calendarID,
		credentials:       opt.credentials,
		credentialsFile:   opt.credentialsFile,
		eventFilterRegexp: opt.eventFilterRegexp,
		location:          opt.location,
		targetDate:        opt.targetDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create eventFetcher: %w", err)
	}

	sp := newSlackPoster(&slackPosterOpt{
		dryRun:           opt.dryRun,
		slackAccessToken: opt.slackAccessToken,
		slackChannelID:   opt.slackChannelID,
		webhook:          opt.webhook,
	})

	return &client{fetcher: ef, poster: sp}, nil
}

// EventData is a data for template.
type EventData struct {
	Msg    string
	Events []*calendar.Event
}

func createSlackMessage(events []*calendar.Event, msg, alt string) (string, error) {
	if len(events) == 0 && alt != "" {
		return alt, nil
	}

	data := EventData{
		Events: events,
		Msg:    msg,
	}

	tmpl, err := template.New("slackMessage").Parse(`{{.Msg}}

{{range .Events -}}
• {{.Summary}}
{{end}}`)
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buffer.String(), nil
}
