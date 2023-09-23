package main

import (
	"fmt"

	"github.com/slack-go/slack"
	"golang.org/x/net/context"
)

type slackPoster struct {
	dryRun           bool
	slackAccessToken string
	slackChannelID   string
	slackWebHook     string
	*slack.Client
}

type slackPosterOpt struct {
	dryRun           bool
	slackAccessToken string
	slackChannelID   string
	webhook          string
}

func newSlackPoster(opt *slackPosterOpt) *slackPoster {
	return &slackPoster{
		dryRun:           opt.dryRun,
		slackAccessToken: opt.slackAccessToken,
		slackChannelID:   opt.slackChannelID,
		slackWebHook:     opt.webhook,
		Client:           slack.New(opt.slackAccessToken),
	}
}

func (p *slackPoster) post(ctx context.Context, msg string) error {
	if err := p.postMessage(ctx, msg); err != nil {
		return err
	}

	if err := p.postWebhook(ctx, msg); err != nil {
		return err
	}

	return nil
}

func (p *slackPoster) postMessage(ctx context.Context, msg string) error {
	logger.Printf("posting message...")

	if p.slackAccessToken == "" {
		logger.Printf("slack access token is empty. skip posting message.")
		return nil
	}
	if p.dryRun {
		logger.Printf("dry run mode. skip posting message.")
		return nil
	}

	_, _, err := p.PostMessageContext(ctx, p.slackChannelID, slack.MsgOptionText(msg, false))
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
