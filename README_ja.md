[English](README.md) | [日本語](README_ja.md)

# slack-notify

`slack-notify` は、Google カレンダーに登録されたイベントを Slack に連携するコマンドおよび GitHub Action です。

## Description

`slack-notify` は次のことを行います。

- Google カレンダーからイベントを取得します
- 取得したイベントを Slack に投稿します

## Usage

```bash
$ slack-notify \
  -credentials ${GOOGLE_CREDENTIALS} \
  -calendar-id ${GOOGLE_CALENDAR_ID} \
  -slack-token ${SLACK_TOKEN} \
  -slack-channel-id ${SLACK_CHANNEL_ID} \
  -location Asia/Tokyo \
  -message '本日のイベントをお知らせします。' \
  -alternative-message '本日のイベントはありません。'

# ヘルプ
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
  -slack-channel-id string
    	Specify Slack Channel ID
  -slack-token string
    	Specify Slack Access Token
  -target-date string
    	Specify targetDate date. e.g. 2020-01-01
  -timeout duration
    	Specify timeout (default 15m0s)
  -v	Show version
```

## Installation

### Developer

```bash
$ go install github.com/ebc-2in2crc/slack-notify/cmd/slack-notify@latest
```

### User

次の URL からダウンロードします。

- [https://github.com/ebc-2in2crc/slack-notify/releases](https://github.com/ebc-2in2crc/slack-notify/releases)

Homebrew を使うこともできます。

```bash
$ brew install ebc-2in2crc/tap/slack-notify
```

### GitHub Action

アクション `ebc-2in2crc/slack-notify@v0` は、Linux 用の slack-notify バイナリを `/usr/local/bin` にインストールします。
このアクションはインストールのみを実行します。

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
