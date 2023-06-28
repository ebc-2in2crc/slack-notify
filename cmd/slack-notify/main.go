package main

import (
	"flag"
	"fmt"
	"github.com/slack-go/slack"
	"google.golang.org/api/option"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/calendar/v3"
)

var version = "0.0.1"

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)
}

type opt struct {
	calendarId        string
	credentials       string
	credentialsFile   string
	eventFilterRegexp string
	location          string
	slackAccessToken  string
	slackChannelId    string
	subject           string
	timeout           time.Duration
	version           bool
}

type eventFetcher struct {
	calendarId  string
	credentials []byte
	filter      *regexp.Regexp
	location    *time.Location
}

func newEventFetcher(opt *opt) (*eventFetcher, error) {
	ef := &eventFetcher{calendarId: opt.calendarId}

	// credentials を読み込む
	if len(opt.credentials) > 0 {
		ef.credentials = []byte(opt.credentials)
	} else {
		c, err := os.ReadFile(opt.credentialsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials file: %w", err)
		}
		ef.credentials = c
	}

	// タイムゾーンを読み込む
	loc, err := time.LoadLocation(opt.location)
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	ef.location = loc

	// イベントをフィルタする正規表現をコンパイルする
	re, err := regexp.Compile(opt.eventFilterRegexp)
	if err != nil {
		logger.Fatalf("failed to compile regexp: %v", err)
	}
	ef.filter = re

	return ef, nil
}

func (s *eventFetcher) fetch(ctx context.Context) ([]string, error) {
	logger.Printf("fetching events...")

	service, err := calendar.NewService(ctx, option.WithCredentialsJSON(s.credentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	tMin, tMax := s.eventsTerm()
	events, err := service.Events.List(s.calendarId).
		TimeMin(tMin.Format(time.RFC3339)).
		TimeMax(tMax.Format(time.RFC3339)).
		SingleEvents(true).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	var a []string
	for _, item := range events.Items {
		if s.filter.MatchString(item.Summary) {
			a = append(a, item.Summary)
		}
	}

	logger.Printf("fetched events: %d, target events: %d", len(events.Items), len(a))
	return a, nil
}

func (s *eventFetcher) eventsTerm() (timeMin, timeMax time.Time) {
	now := time.Now().In(s.location)
	min := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.location).UTC()
	max := min.AddDate(0, 0, 1).UTC()
	return min, max
}

type slackPoster struct {
	slackAccessToken string
	*slack.Client
}

func newSlackPoster(opt *opt) *slackPoster {
	return &slackPoster{
		slackAccessToken: opt.slackAccessToken,
		Client:           slack.New(opt.slackAccessToken),
	}
}

func (p *slackPoster) post(ctx context.Context, channelId, msg string) error {
	logger.Printf("posting message...")

	_, _, err := p.PostMessageContext(ctx, channelId, slack.MsgOptionText(msg, false))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	logger.Printf("posted message.")
	return nil
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

	ef, err := newEventFetcher(opt)
	if err != nil {
		logger.Fatalf("failed to create eventFetcher: %v", err)
	}

	// イベントを取得する
	events, err := ef.fetch(ctx)
	if err != nil {
		logger.Fatalf("failed to fetch: %v", err)
	}

	// Slack に投稿するメッセージを作成する
	for i := range events {
		events[i] = fmt.Sprintf("• %s", events[i])
	}
	msg := fmt.Sprintf("<!here> %s\n\n%s", opt.subject, strings.Join(events, "\n"))

	// Slack に投稿する
	sp := newSlackPoster(opt)
	if err := sp.post(ctx, opt.slackChannelId, msg); err != nil {
		logger.Fatalf("failed to post: %v", err)
	}
}

func parseFlag() (*opt, error) {
	calendarId := flag.String("calendar-id", "", "Specify Google Calendar ID")
	credentials := flag.String("credentials", "", "Specify credentials")
	credentialsFile := flag.String("credentials-file", "", "Specify credentials file")
	eventFilterRegexp := flag.String("event-filter-regexp", ".", "Specify event filter regexp")
	location := flag.String("location", "UTC", "Specify Location")
	slackAccessToken := flag.String("slack-token", "", "Specify Slack Access Token")
	slackChannelId := flag.String("slack-channel-id", "", "Specify Slack Channel ID")
	subject := flag.String("subject", "", "Specify subject")
	timeoutOption := flag.Duration("timeout", 15*time.Minute, "Specify timeout")
	version := flag.Bool("v", false, "Show version")
	flag.Parse()

	if version != nil {
		return &opt{version: *version}, nil
	}

	if *credentials == "" && *credentialsFile == "" {
		return nil, fmt.Errorf("credentials or credentials-file must be specified")
	}
	if *calendarId == "" {
		return nil, fmt.Errorf("calendar-id must be specified")
	}
	if *slackAccessToken == "" {
		return nil, fmt.Errorf("slack-token must be specified")
	}
	if *slackChannelId == "" {
		return nil, fmt.Errorf("slack-channel-id must be specified")
	}

	return &opt{
		calendarId:        *calendarId,
		credentials:       *credentials,
		credentialsFile:   *credentialsFile,
		eventFilterRegexp: *eventFilterRegexp,
		location:          *location,
		slackAccessToken:  *slackAccessToken,
		slackChannelId:    *slackChannelId,
		subject:           *subject,
		timeout:           *timeoutOption,
		version:           *version,
	}, nil
}