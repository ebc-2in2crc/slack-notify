[English](README.md) | [日本語](README_ja.md)

# slack-notify

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ebc-2in2crc/slack-notify)](https://goreportcard.com/report/github.com/ebc-2in2crc/slack-notify)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ebc-2in2crc/slack-notify)](https://img.shields.io/github/go-mod/go-version/ebc-2in2crc/slack-notify)
[![Version](https://img.shields.io/github/release/ebc-2in2crc/slack-notify.svg?label=version)](https://img.shields.io/github/release/ebc-2in2crc/slack-notify.svg?label=version)

`slack-notify` is a command and GitHub Action to link events in your Google Calendar to Slack.

## Description

`slack-notify` does the following

- Retrieve events from Google Calendar
- Post the retrieved events to Slack
- The message to post on Slack is customizable

## Usage

```bash
$ slack-notify \
  -credentials ${GOOGLE_CREDENTIALS} \
  -calendar-id ${GOOGLE_CALENDAR_ID} \
  -slack-token ${SLACK_TOKEN} \
  -slack-channel-id ${SLACK_CHANNEL_ID} \
  -location Asia/Tokyo \
  -message "Here is an update on today's event." \
  -alternative-message "There are no events today."

# usage
$ slack-notify -h
Usage of slack-notify:
  -alternative-message string
    	Specify alternative message
  -calendar-id string
    	Specify Google Calendar ID
  -credentials string
    	Specify credentials
  -credentials-file string
    	Specify credentials file
  -dry-run
    	Specify dry-run
  -event-filter-regexp string
    	Specify event filter regexp (default ".")
  -location string
    	Specify Location (default "UTC")
  -message string
    	Specify message
  -message-template-file string
        Specify custom message template file
  -slack-channel-id string
    	Specify Slack Channel ID
  -slack-token string
    	Specify Slack Access Token
  -target-date string
    	Specify targetDate date. e.g. 2020-01-01
  -timeout duration
    	Specify timeout (default 15m0s)
  -v	Show version
  -webhook string
        Specify Slack Webhook URL
```

## Installation

### Developer

```bash
$ go install github.com/ebc-2in2crc/slack-notify/cmd/slack-notify@latest
```

### User

Download from the following URL.

- [https://github.com/ebc-2in2crc/slack-notify/releases](https://github.com/ebc-2in2crc/slack-notify/releases)


You can also use Homebrew.

```bash
$ brew install ebc-2in2crc/tap/slack-notify
```

### GitHub Action

このアクションはインストールのみを実行します。
The action `ebc-2in2crc/slack-notify@v0` installs the slack-notify binary for Linux in `/usr/local/bin`.
This action only performs the installation.

```yaml
jobs:
  slack-notify:
    name: slack-notify
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Install slack-notify
        uses: ebc-2in2crc/slack-notify@v0
      - name: Notify
        run: |-
          slack-notify -v
```

## Contribution

1. Fork this repository
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Rebase your local changes against the master branch
5. Run test suite with the `make test` command and confirm that it passes
6. Run `make fmt` and `make lint`
7. Create new Pull Request

## Licence

[MIT](https://github.com/ebc-2in2crc/slack-notify/blob/master/LICENSE)

## Author

[ebc-2in2crc](https://github.com/ebc-2in2crc)
