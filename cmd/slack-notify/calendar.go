package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type eventFetcher struct {
	calendarID  string
	credentials []byte
	filter      *regexp.Regexp
	targetDate  time.Time
}

type eventFetcherOpt struct {
	calendarID        string
	credentials       string
	credentialsFile   string
	eventFilterRegexp string
	location          string
	targetDate        string
}

func newEventFetcher(opt *eventFetcherOpt) (*eventFetcher, error) {
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
