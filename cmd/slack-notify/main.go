package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"google.golang.org/api/option"

	"golang.org/x/net/context"
	"google.golang.org/api/calendar/v3"
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

type eventFetcher struct {
	calendarID  string
	credentials []byte
	filter      *regexp.Regexp
	targetDate  time.Time
}

func newEventFetcher(opt *opt) (*eventFetcher, error) {
	ef := &eventFetcher{calendarID: opt.calendarID}

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

	// 日付の文字列をパースする
	if opt.targetDate == "" {
		ef.targetDate = time.Now().In(loc)
	} else {
		t, err := time.ParseInLocation("2006-01-02", opt.targetDate, loc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target date: %w", err)
		}
		ef.targetDate = t
	}

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
	events, err := service.Events.List(s.calendarID).
		TimeMin(tMin.Format(time.RFC3339)).
		TimeMax(tMax.Format(time.RFC3339)).
		SingleEvents(true).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].Start.DateTime < events.Items[j].Start.DateTime
	})

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
	year := s.targetDate.Year()
	month := s.targetDate.Month()
	day := s.targetDate.Day()
	loc := s.targetDate.Location()

	min := time.Date(year, month, day, 0, 0, 0, 0, loc).UTC()
	max := min.AddDate(0, 0, 1).UTC()
	return min, max
}

type slackPoster struct {
	dryRun           bool
	slackAccessToken string
	slackWebHook     string
	*slack.Client
}

func newSlackPoster(opt *opt) *slackPoster {
	return &slackPoster{
		dryRun:           opt.dryRun,
		slackAccessToken: opt.slackAccessToken,
		slackWebHook:     opt.webhook,
		Client:           slack.New(opt.slackAccessToken),
	}
}

func (p *slackPoster) post(ctx context.Context, channelID, msg string) error {
	if err := p.postMessage(ctx, channelID, msg); err != nil {
		return err
	}

	if err := p.postWebhook(ctx, msg); err != nil {
		return err
	}

	return nil
}

func (p *slackPoster) postMessage(ctx context.Context, channelID, msg string) error {
	logger.Printf("posting message...")

	if p.slackAccessToken == "" {
		logger.Printf("slack access token is empty. skip posting message.")
		return nil
	}
	if p.dryRun {
		logger.Printf("dry run mode. skip posting message.")
		return nil
	}

	_, _, err := p.PostMessageContext(ctx, channelID, slack.MsgOptionText(msg, false))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	logger.Printf("posted message.")
	return nil
}

func (p *slackPoster) postWebhook(ctx context.Context, msg string) error {
	logger.Printf("posting webhook...")

	if p.slackWebHook == "" {
		logger.Printf("slack webhook is empty. skip posting webhook.")
		return nil
	}
	if p.dryRun {
		logger.Printf("dry run mode. skip posting webhook.")
		return nil
	}

	message := &slack.WebhookMessage{Text: msg}
	if err := slack.PostWebhookContext(ctx, p.slackWebHook, message); err != nil {
		return fmt.Errorf("failed to post webhook: %w", err)
	}

	logger.Printf("posted webhook.")
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
	desc := opt.message
	if len(events) == 0 && opt.alternativeMessage != "" {
		desc = opt.alternativeMessage
	}
	msg := fmt.Sprintf("%s\n\n%s", desc, strings.Join(events, "\n"))

	// Slack に投稿する
	sp := newSlackPoster(opt)
	if err := sp.post(ctx, opt.slackChannelID, msg); err != nil {
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
